package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gitlab.com/derwolfe/faststats/api"
	"gitlab.com/derwolfe/faststats/db"
)

func main() {
	// open the DB in read only mode. If we get SQLi this should limit damage.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting on %v\n", port)

	db, err := db.BuildDB("./results.db?_query_only=1")
	if err != nil {
		log.Fatal(err)
	}
	api := api.NewAPI(db)

	http.HandleFunc("/", api.SearchForm)
	http.HandleFunc("/search", api.Search)
	http.HandleFunc("/results", api.Results)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil) // setting listening port

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
