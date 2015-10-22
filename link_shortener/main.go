package main

import (
	"log"
	"net/http"
	"math/rand"
	"strings"

	"github.com/drone/routes"
	"gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
)

// ===== CONSTANTS / GLOBALS =======================================================================

const (
	collectionName = "shortened_urls"
	databaseName = "traction_demo"
	httpPort = "12345"
	shortRegex = ":shortID([a-f0-9]{8})" // 16 ** 8 = 4,294,967,296 possible URLs

	// must be ASCII for this demo because we index into it directly and also use len() on it
	shortIdChars = "01234567890abcdef"
)

// our MongoDB collection
var collection *mgo.Collection

// ===== STRUCTURES ================================================================================

type ShortenedUrl struct {
	LongUrl string
	ShortId string
	Visitors []*Visit
}

type Visit struct {
	IpAddress string
	UserAgent  string
	Country    string
	State      string
	Referrer   string
}

// ===== FUNCTIONS =================================================================================

func connectToMongo(host string) *mgo.Session {
	log.Println("Trying to connect to mongo @", host)

	conn, err := mgo.Dial(host)
	if err != nil {
		log.Fatalln("mgo.Dial:", err)
	}

	log.Println("Connected.")

	return conn
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

func shortIdFromUrl(req *http.Request) string {
	return req.URL.Query().Get(":shortID")
}

func shortIdExists(id string) bool {
	if 0 == len(id) {
		return false
	}

	// check for id in database
	return exists_in_db(id)
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

	// 1. make sure the long URL is _somewhat valid_ (see my comments on :url in models/listing.rb)
	long_url := req.FormValue("url")

	if !urlIsValid(long_url) {
		http.Error(w, "", http.StatusConflict)
		return
	}

	// 2. come up with a unique short ID for this URL
	// Generate a short ID, see if it's in use, repeat until one isn't in use
	short_id := generateShortId()

	for ; shortIdExists(short_id); {
		short_id = generateShortId()
	}

	// 3. store the original URL + short ID into the db

	http.Error(w, short_id, http.StatusCreated)
}

// returns a 301 "Moved Permanently" if the short url exists
// records the click and associated metrics/data
// returns a 404 "Not Found" if no match
func shortUrlRedirectHandler(w http.ResponseWriter, req *http.Request) {

	short_id := shortIdFromUrl(req)

	// 1. see if the short ID exists
	if !shortIdExists(short_id) {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// 2. if it does, extract a few things:
	//   a. IP Address
	//   b. User Agent
	//   c. Country and State (from MaxMind)
	//   d. Referrer

	// 3. record the visit into the database

	// 4. send the 301 redirect to the full URL

	http.Error(w, "", http.StatusMovedPermanently)
}

// formats and returns a 200 "OK" JSON response of stats/metrics about the short URL in question
// returns a 404 "Not Found" if no match
func shortUrlStatsHandler(w http.ResponseWriter, req *http.Request) {

	short_id := shortIdFromUrl(req)

	// 1. see if the short ID exists
	if !shortIdExists(short_id) {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// 2. compile the stats and send them back as a 200 JSON blob

	http.Error(w, "", http.StatusOK)
}

// ===== ENTRYPOINT ================================================================================

func main() {

	// ----- SETUP ---------------------------------------------------------------------------------

	// NOTE: you'd probably want to set GOMAXPROCS or runtime.GOMAXPROCS() here

	mongo_conn := connectToMongo("localhost")
	defer mongo_conn.Close()
	mongo_conn.SetMode(mgo.Strong, true) // don't trust computers, ever
	mongo_conn.SetSafe(&mgo.Safe{})

	collection = mongo_conn.DB(databaseName).C(collectionName)

	mux := routes.New()

	// ----- HTTP HANDLERS -------------------------------------------------------------------------

	mux.Get("/ping", emptyHandler) // useful for poking the server to see if it's alive

	mux.Post("/urls/shorten", shortenUrlHandler) // returns a new shortened URL
	mux.Get("/urls/" + shortRegex + "/stats" , shortUrlStatsHandler) // returns stats as JSON

	mux.Get("/s/" + shortRegex, shortUrlRedirectHandler) // records visit and redirects to full URL

	http.Handle("/", mux)

	// ----- HTTP SERVER ---------------------------------------------------------------------------

	log.Println("Listening on localhost:", httpPort)

	err := http.ListenAndServe(":" + httpPort, nil)
	if err != nil {
		log.Fatalln("http.ListenAndServe:", err)
	}
}
