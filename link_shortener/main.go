package main

import (
	"net/http"

	"github.com/drone/routes"
)

// ===== CONSTANTS / GLOBALS =======================================================================

const (
	shortRegex = ":short([a-f0-9]{8})" // 16 ** 8 = 4,294,967,296 possible URLs
)

// ===== HTTP HANDLERS =============================================================================

func emptyHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusNoContent)
}

// ===== ENTRYPOINT ================================================================================

func main() {
	mux := routes.New()

	mux.Get("/ping", emptyHandler) // useful for poking the server to see if it's alive

	mux.Post("/urls/shorten", emptyHandler) // returns a new shortened URL
	mux.Get("/urls/" + shortRegex + "/stats" , emptyHandler) // returns stats as JSON

	mux.Get("/s/" + shortRegex, emptyHandler) // records visit and redirects to full URL

	http.Handle("/", mux)
}
