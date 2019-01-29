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
	lifters, err := db.QueryNames("francisco flores")

	assert.Nil(t, err, "query for names returned an error")
	fmt.Printf("%v\n", lifters)
}

func TestQueryResults(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	lifters, err := db.QueryNames("chris wolfe")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, lifters, "no names returned")

	summary, err := db.QueryResults(lifters[0].Name, lifters[0].Hometown)
	assert.Nil(t, err, "query for results failed")

	assert.NotEmpty(t, summary.Results, "no results for lifter")
	for _, result := range summary.Results {
		// let's make sure we have what we expect
		fmt.Printf("%v\n", result)
	}
}
