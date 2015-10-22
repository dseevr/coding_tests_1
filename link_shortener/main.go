package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/drone/routes"
	"github.com/oschwald/geoip2-golang"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ===== CONSTANTS / GLOBALS =======================================================================

const (
	databaseName  = "traction_demo"
	debug         = true
	httpPort      = "12345"
	maxmindDbName = "GeoLite2-Country.mmdb"
	shortDomain   = "localhost" + ":" + httpPort // setting to "localhost" is convenient for testing!
	shortRegex    = ":shortID([a-f0-9]{8})"      // 16 ** 8 = 4,294,967,296 possible URLs

	// must be ASCII for this demo because we index into it directly and also use len() on it
	shortIdChars = "01234567890abcdef"

	urlCollectionName   = "shortened_urls"
	visitCollectionName = "visits"
)

// our MongoDB collections
var url_collection *mgo.Collection
var visit_collection *mgo.Collection

// for state/country lookup
var maxmindDB *geoip2.Reader

// ===== STRUCTURES ================================================================================

type ShortenedUrl struct {
	Id      bson.ObjectId `bson:"_id"`
	LongUrl string        `bson:"long_url"`
	ShortId string        `bson:"short_id"`
	Visits  []*Visit      `bson:"visits"`
}

type Visit struct {
	ShortenedUrlId bson.ObjectId `bson:"shortened_url_id"`
	IpAddress      string        `bson:"ip_address"`
	UserAgent      string        `bson:"user_agent"`
	Country        string        `bson:"country"`
	Referrer       string        `bson:"referrer"`
}

// ----- STATS STRUCTURES --------------------------------------------------------------------------

type Stats struct {
	Countries   []StatsCountries `json:"by_country"`
	Referrers   []StatsReferrer  `json:"by_referrer"`
	TotalVisits int              `json:"total_visits"`
}

type StatsCountries struct {
	Name   string `json:"name"`
	Visits int    `json:"visits"`
}

type StatsReferrer struct {
	Url    string `json:"url"`
	Visits int    `json:"visits":`
}

// ===== FUNCTIONS =================================================================================

func connectToMongo(host string) *mgo.Session {
	log.Println("Trying to connect to mongo @", host)

	conn, err := mgo.Dial(host)
	if err != nil {
		log.Fatalln("mgo.Dial:", err)
	}

	conn.SetMode(mgo.Strong, true) // don't trust computers, ever
	conn.SetSafe(&mgo.Safe{})

	log.Println("Connected.")

	return conn
}

func countryFromIp(ip string) string {
	record, err := maxmindDB.Country(net.ParseIP(ip))
	if err != nil {
		log.Fatalln("geoip2.Country:", err)
	}

	country := record.Country.Names["en"]

	log.Printf("Found country \"%v\" for IP %v\n", country, ip)

	return country
}

func createNewShortUrl(long_url, short_id string) string {
	bson_id := bson.NewObjectId()

	shortened_url := ShortenedUrl{
		Id:      bson_id,
		LongUrl: long_url,
		ShortId: short_id,
	}

	err := url_collection.Insert(shortened_url)
	if err != nil {
		log.Fatalln("url_collection.Insert:", err)
	}

	log.Printf("Created new short URL: %v -> %v (id: %v)\n", short_id, long_url, bson_id)

	return shortUrlFromShortId(short_id)
}

func findRecordByShortId(id string) (*ShortenedUrl, bool) {
	result := ShortenedUrl{}

	err := url_collection.Find(bson.M{"short_id": id}).One(&result)

	if nil == err {
		return &result, true
	} else if mgo.ErrNotFound == err {
		return nil, false
	} else {
		log.Fatalln("url_collection.Find:", err)
		return nil, false
	}
}

// creates a random 8 character hex string
func generateShortId() string {
	s := ""
	max := len(shortIdChars) // don't need to -1 because rand.Intn is exclusive of the upper bound

	for i := 0; i < 8; i++ {
		s += string(shortIdChars[rand.Intn(max)])
	}

	return s
}

func ipFromRequest(req *http.Request) string {
	// NOTE: this should really be checking for X-Forwarded-For, X-Real-IP, etc.

	// NOTE: You can set a custom IP for testing, e.g.,:
	//       curl -H "X-IP: 1.2.3.4" http://localhost/s/abc123
	custom_ip := req.Header.Get("X-IP")

	ip := ""

	if 0 == len(custom_ip) {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	} else {
		ip = custom_ip
	}

	// NOTE: on localhost, the IP that Go shows me is the IPv6 loopback address, so if the result
	//       of strings.Split() above isn't parseable as IPv4, just convert it to the IPv4 loopback
	//       so Maxmind can look it up without exploding (and return "").
	if nil == net.ParseIP(ip) {
		ip = "127.0.0.1"
	}

	return ip
}

func loadMaxmindDb() *geoip2.Reader {
	db, err := geoip2.Open(maxmindDbName)
	if err != nil {
		log.Fatalln("geoip2.Open:", err)
	}

	return db
}

func longUrlForShortId(id string) string {
	record, found := findRecordByShortId(id)

	if found {
		return record.LongUrl
	} else {
		panic("Couldn't find long URL for short ID: " + id)
	}
}

func recordVisit(visit Visit) {
	err := visit_collection.Insert(visit)
	if err != nil {
		log.Fatalln("visit_collection.Insert:", err)
	}

	log.Println("Recorded new visit:")
	log.Println("IP:", visit.IpAddress)
	log.Println("User Agent:", visit.UserAgent)
	log.Println("Country:", visit.Country)
	log.Println("Referrer:", visit.Referrer)
}

func shortIdFromUrl(req *http.Request) string {
	return req.URL.Query().Get(":shortID")
}

func shortIdExists(id string) bool {
	if 0 == len(id) {
		return false
	}

	log.Println("Looking for short ID:", id)

	_, found := findRecordByShortId(id)

	if found {
		log.Println("Found!")
	} else {
		log.Println("Not found!")
	}

	return found
}

func shortUrlFromShortId(id string) string {
	return "http://" + shortDomain + "/s/" + id
}

func urlIsValid(url string) bool {
	// see models/listing.rb for more discussion about this
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// ===== HTTP HANDLERS =============================================================================

// returns a 204 "No Content" for testing purposes
func emptyHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusNoContent)
}

// returns a 201 "Resource Created" if a short URL is successfully created
// the json response includes the short URL if successful
// returns a 409 "Conflict" otherwise (there's really no consensus on status code for this case)
func shortenUrlHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("shortenUrlHandler:")

	// 1. make sure the long URL is _somewhat valid_ (see my comments on :url in models/listing.rb)
	long_url := req.FormValue("url")

	if !urlIsValid(long_url) {
		http.Error(w, "", http.StatusConflict)
		return
	}

	// 2. come up with a unique short ID for this URL
	// Generate a short ID, see if it's in use, repeat until one isn't in use
	short_id := generateShortId()

	for shortIdExists(short_id) {
		short_id = generateShortId()
	}

	// 3. store the original URL + short ID into the db
	shortened_url := createNewShortUrl(long_url, short_id)

	http.Error(w, shortened_url, http.StatusCreated)
}

// returns a 301 "Moved Permanently" if the short url exists
// records the click and associated metrics/data
// returns a 404 "Not Found" if no match
func shortUrlRedirectHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("shortUrlRedirectHandler:")

	short_id := shortIdFromUrl(req)

	// 1. see if the short ID exists

	// we need the BSON id later, so use this function instead of shortIdExists()
	record, found := findRecordByShortId(short_id)
	if !found {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// 2. if it does, extract a few things:
	//   a. IP Address
	//   b. User Agent
	//   c. Country (from MaxMind)
	//   d. Referrer

	ip := ipFromRequest(req)

	visit := Visit{
		ShortenedUrlId: record.Id,
		IpAddress:      ip,
		UserAgent:      req.Header.Get("User-Agent"),
		Country:        countryFromIp(ip),
		Referrer:       req.Referer(),
	}

	// 3. record the visit into the database
	recordVisit(visit)

	// 4. send the 301 redirect to the full URL
	w.Header().Set("Location", record.LongUrl)

	log.Println("Redirecting from short ID", short_id, "to long URL:", record.LongUrl)

	http.Error(w, "", http.StatusMovedPermanently)
}

// formats and returns a 200 "OK" JSON response of stats/metrics about the short URL in question
// returns a 404 "Not Found" if no match
func shortUrlStatsHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("shortUrlStatsHandler:")

	short_id := shortIdFromUrl(req)

	// 1. see if the short ID exists

	// we need the BSON id, so use this function instead of shortIdExists()
	record, found := findRecordByShortId(short_id)
	if !found {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// 2. compile the stats and send them back as a 200 JSON blob

	stats := Stats{}

	// ----- TOTAL VISITS --------------------------------------------------------------------------

	query := bson.M{"shortened_url_id": record.Id}

	count, err := visit_collection.Find(query).Count()
	if err != nil {
		log.Println("visit_collection.Find:", err)
	}

	stats.TotalVisits = count

	// ----- VISITS BY COUNTRY ---------------------------------------------------------------------

	// NOTE: I wanted to try out MapReduce in MongoDB... obviously you would never do this live haha

	country_job := mgo.MapReduce{
		Map:    "function() { if(!this.country.length) return; emit(this.country, 1); }",
		Reduce: "function(country, count) { return Array.sum(count) }",
	}

	var country_results []struct {
		Name   string "_id"
		Visits int    "value"
	}

	_, err = visit_collection.Find(query).MapReduce(&country_job, &country_results)
	if err != nil {
		log.Fatalln("visit_collection.Find:", err)
	}

	for _, item := range country_results {
		sc := StatsCountries{Name: item.Name, Visits: item.Visits}
		stats.Countries = append(stats.Countries, sc)
	}

	// ----- TOP REFERRERS -------------------------------------------------------------------------

	referrer_results := []Visit{}

	iter := visit_collection.Find(query).Limit(1000).Iter()
	err = iter.All(&referrer_results)
	if err != nil {
		log.Fatalln("visit_collection.Find:", err)
	}

	referring_urls := make(map[string]int)

	for _, value := range referrer_results {
		if 0 == len(value.Referrer) {
			// skip empty referrers, could also do this at the DB level
			continue
		}

		referring_urls[value.Referrer] += 1
	}

	sr_arr := []StatsReferrer{}

	for key, value := range referring_urls {
		sr := StatsReferrer{Url: key, Visits: value}
		sr_arr = append(sr_arr, sr)
	}

	stats.Referrers = sr_arr

	// ----- RENDER JSON RESPONSE ------------------------------------------------------------------

	bytes, err := json.Marshal(stats)
	if err != nil {
		log.Fatalln("json.Marshal:", err)
	}

	log.Println("Rendered stats JSON for", short_id)

	http.Error(w, string(bytes), http.StatusOK)
}

// ===== ENTRYPOINT ================================================================================

func main() {

	// ----- SETUP ---------------------------------------------------------------------------------

	// NOTE: you'd probably want to set runtime.GOMAXPROCS() here if GOMAXPROCS isn't set.

	// NOTE: comment this out, generate a short URL, restart the server, and generate another short
	//       URL to test that collision detection is handled properly.
	rand.Seed(time.Now().UTC().UnixNano())

	mongo_conn := connectToMongo("localhost")
	defer mongo_conn.Close()

	mongo_db := mongo_conn.DB(databaseName)

	url_collection = mongo_db.C(urlCollectionName)
	visit_collection = mongo_db.C(visitCollectionName)

	maxmindDB = loadMaxmindDb()
	defer maxmindDB.Close()

	// ----- HTTP HANDLERS -------------------------------------------------------------------------

	mux := routes.New()

	mux.Get("/ping", emptyHandler) // useful for poking the server to see if it's alive

	mux.Post("/urls/shorten", shortenUrlHandler)                // returns a new shortened URL
	mux.Get("/urls/"+shortRegex+"/stats", shortUrlStatsHandler) // returns stats as JSON

	mux.Get("/s/"+shortRegex, shortUrlRedirectHandler) // records visit and redirects to full URL

	http.Handle("/", mux)

	// ----- HTTP SERVER ---------------------------------------------------------------------------

	log.Println("Listening on localhost:" + httpPort)

	err := http.ListenAndServe(":"+httpPort, nil)
	if err != nil {
		log.Fatalln("http.ListenAndServe:", err)
	}
}
