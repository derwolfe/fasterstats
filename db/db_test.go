package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
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

func TestQueryNamesRegression(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	r, err := db.QueryNames("mos", "1")
	assert.Nil(t, err, "query for names returned an error")
	assert.NotEmpty(t, r, "no names returned")
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

	r, err := db.QueryNames("foooooo", "1")
	assert.Nil(t, err, "query for names returned an error")

	assert.Equal(t, len(r.Lifters), 0, "no results should be returned")
	assert.Equal(t, r.Total, int64(0), "the total was non-zero")
}

func TestIWFNames(t *testing.T) {
	names := []struct {
		in    string
		first string
		last  string
	}{
		{"Martha (Mattie) Rogers", "martha", "rogers"},
		{"D'angelo Osorio", "dangelo", "osorio"}, // strip the apostrophe
		{"chris", "chris", ""},                   // all names should have a last, but this is how it would run
		{"", "", ""},                             // this should never be an input
	}
	for _, tt := range names {
		t.Run(tt.in, func(t *testing.T) {
			first, last := ToIWFName(tt.in)
			assert.Equal(t, tt.first, first, "failed to convert first")
			assert.Equal(t, tt.last, last, "failed to convert last")
		})
	}
}

var lifterResponseResult *LiftersResponse

func BenchmarkNameQuery(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	db, err := BuildDB("../results.db")
	defer db.Close()
	if err != nil {
		panic("failed to setup DB")
	}

	for i := 0; i < b.N; i++ {
		r, _ := db.QueryNames("mattie rogers", "1")
		lifterResponseResult = r
	}
}

var resultsResponse *ResultsSummary

func BenchmarkResultsQuery(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	db, err := BuildDB("../results.db")
	defer db.Close()
	if err != nil {
		panic("failed to setup DB")
	}

	for i := 0; i < b.N; i++ {
		r, _ := db.QueryResults("D'Angelo Osorio", "Vallejo, CA")
		resultsResponse = r
	}
}

func TestNoBadData(t *testing.T) {
	db, err := BuildDB("../results.db")
	defer db.Close()
	if err != nil {
		panic("failed to get total")
	}

	// look every name in the DB up, check that it finds results without error
	var total int64
	err = db.db.QueryRow(`SELECT COUNT(*) FROM (SELECT DISTINCT lifter, hometown FROM results ORDER BY lifter ASC)`).Scan(&total)
	if err != nil {
		panic("failed to get total")
	}

	rows, err := db.db.Query(`SELECT DISTINCT lifter, hometown FROM results ORDER BY lifter ASC`)
	lifters := make([]Lifter, total, total)
	ct := 0
	for rows.Next() {
		l := Lifter{}
		err = rows.Scan(&l.Name, &l.Hometown)
		if err != nil {
			panic("failed loading lifters")
		}
		lifters[ct] = l
		ct++
	}
	defer rows.Close()
	for _, tt := range lifters {
		testName := fmt.Sprintf("%v, %v", tt.Name, tt.Hometown)
		t.Run(testName, func(t *testing.T) {
			_, err := db.QueryResults(tt.Name, tt.Hometown)
			assert.Nil(t, err, "error querying for", tt.Name, tt.Hometown)
		})
	}
}
