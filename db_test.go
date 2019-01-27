package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryNames(t *testing.T) {
	db, err := BuildDB()
	defer db.Close()

	assert.Nil(t, err, "failed to build db")

	// this relies on data in the DB!
	lifters, err := db.QueryForNames("d'angelo")

	assert.Nil(t, err, "query for names returned an error")
	fmt.Printf("%v\n", lifters)

	// assert that d'angelo and a few others are in the data set
}
