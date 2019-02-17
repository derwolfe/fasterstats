package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryNames(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	lifters, err := db.QueryNames("francisco flores", "1")

	assert.Nil(t, err, "query for names returned an error")
	fmt.Printf("%v\n", lifters)
}

func TestQueryResults(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	lifters, err := db.QueryNames("chris wolfe", "1")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, lifters, "no names returned")

	summary, err := db.QueryResults(lifters.Lifters[0].Name, lifters.Lifters[0].Hometown)
	assert.Nil(t, err, "query for results failed")

	assert.NotEmpty(t, summary.Results, "no results for lifter")
	for _, result := range summary.Results {
		// let's make sure we have what we expect
		fmt.Printf("%v\n", result)
	}
}

func TestQueryResultsLikeReplace(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	lifters, err := db.QueryNames("j bradley", "1")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, lifters.Lifters, "no names returned")

	pass := false
	for _, l := range lifters.Lifters {
		if l.Name == "Jessie Bradley" {
			pass = true
		}
	}
	assert.True(t, pass, "lifter not found")
}
