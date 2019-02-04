package main

import (
	"log"
	"net/http"

	"gitlab.com/derwolfe/faststats/api"
	"gitlab.com/derwolfe/faststats/db"
)

func main() {
	// open the DB in read only mode. If we get SQLi this should limit damage.
	db, err := db.BuildDB("./results.db?_query_only=1")
	if err != nil {
		log.Fatal(err)
	}
	api := api.NewAPI(db)

	http.HandleFunc("/", api.SearchForm)
	http.HandleFunc("/search", api.Search)
	http.HandleFunc("/results", api.Results)

	log.Println("Starting on :9090")
	err = http.ListenAndServe(":9090", nil) // setting listening port

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
