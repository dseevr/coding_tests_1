package main

import (
	"log"
	"net/http"

	"github.com/drone/routes"
)

// ===== CONSTANTS / GLOBALS =======================================================================

const (
	httpPort = "12345"
	shortRegex = ":short([a-f0-9]{8})" // 16 ** 8 = 4,294,967,296 possible URLs
)

// ===== FUNCTIONS =================================================================================

// ===== HTTP HANDLERS =============================================================================

// returns a 204 "No Content" for testing purposes
func emptyHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusNoContent)
}

// returns a 201 "Resource Created" if a short URL is successfully created
// the json response includes the short URL if successful
// returns a 409 "Conflict" otherwise (there's really no consensus on status code for this case)
func shortenUrlHandler(w http.ResponseWriter, req *http.Request) {
	// 1. make sure URL is _somewhat valid_ (see my comments on :url in models/listing.rb)

	// 2. come up with a unique short ID for this URL
	// Generate an 8 character hex string, see if it's in use, repeat until one isn't in use

	// 3. store the original URL + short ID into the db

	// 4. return a 200 JSON blob with the short URL or a 409

	// if it succeeds:
	http.Error(w, "", http.StatusCreated)

	// or if it fails:
	// http.Error(w, "", http.StatusConflict)
}

// returns a 301 "Moved Permanently" if the short url exists
// records the click and associated metrics/data
// returns a 404 "Not Found" if no match
func shortUrlRedirectHandler(w http.ResponseWriter, req *http.Request) {
	// 1. see if the short URL exists

	// 2. if it does, extract a few things:
	//   a. IP Address
	//   b. User Agent
	//   c. Country and State (from MaxMind)
	//   d. Referrer

	// 3. record the visit into the database

	// 4. send the 301 redirect to the full URL

	http.Error(w, "", http.StatusMovedPermanently)

	// or if it fails:
	// http.Error(w, "", http.StatusNotFound)
}

// formats and returns a 200 "OK" JSON response of stats/metrics about the short URL in question
// returns a 404 "Not Found" if no match
func shortUrlStatsHandler(w http.ResponseWriter, req *http.Request) {
	// 1. see if the short URL exists

	// 2. compile the stats and send them back as a 200 JSON blob

	// if it succeeds:
	http.Error(w, "", http.StatusOK)

	// or if it fails:
	http.Error(w, "", http.StatusNotFound)
}

// ===== ENTRYPOINT ================================================================================

func main() {
	mux := routes.New()

	mux.Get("/ping", emptyHandler) // useful for poking the server to see if it's alive

	mux.Post("/urls/shorten", shortenUrlHandler) // returns a new shortened URL
	mux.Get("/urls/" + shortRegex + "/stats" , shortUrlStatsHandler) // returns stats as JSON

	mux.Get("/s/" + shortRegex, shortUrlRedirectHandler) // records visit and redirects to full URL

	http.Handle("/", mux)

	log.Println("Listening on localhost:", httpPort)

	err := http.ListenAndServe(":" + httpPort, nil)
	if err != nil {
		log.Fatalln("http.ListenAndServe:", err)
	}
}
