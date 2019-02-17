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
	r, err := db.QueryNames("francisco flores", "1")

	assert.Nil(t, err, "query for names returned an error")
	fmt.Printf("%v\n", r)
}

func TestQueryResults(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	r, err := db.QueryNames("chris wolfe", "1")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, r, "no names returned")

	summary, err := db.QueryResults(r.Lifters[0].Name, r.Lifters[0].Hometown)
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
	r, err := db.QueryNames("j bradley", "1")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, r.Lifters, "no names returned")

	pass := false
	for _, l := range r.Lifters {
		if l.Name == "Jessie Bradley" {
			pass = true
		}
	}
	assert.True(t, pass, "lifter not found")
}

func TestQueryCountAccurate(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	r, err := db.QueryNames("kyle brown", "1")
	assert.Nil(t, err, "query for names returned an error")

	assert.Equal(t, len(r.Lifters), 2, "two r not returned for kyle brown")
	assert.Equal(t, r.Total, int64(2), "the total was not two")

	// this relies on data in the DB!
	r, err = db.QueryNames("steph", "1")
	assert.Nil(t, err, "query for names returned an error")

	// this could be flaky but should be at least 300 lifters
	assert.Equal(t, len(r.Lifters), 50, "two r not returned for kyle brown")
	assert.True(t, r.Total > 50, "the total was not larger than a page boundary")
}

func TestQueryWorksWhenZero(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	r, err := db.QueryNames("foooooo", "1")
	assert.Nil(t, err, "query for names returned an error")

	assert.Equal(t, len(r.Lifters), 0, "two r not returned for kyle brown")
	assert.Equal(t, r.Total, int64(0), "the total was not two")

	assert.Nil(t, err, "query for names returned an error")
}