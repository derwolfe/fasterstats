package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"log"
)

func buildDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./results.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type Lifter struct {
	Name     string
	Hometown string
}

type Result struct {
	Date        string
	MeetName    string
	Lifter      string
	WeightClass string
	Hometown    string
	CJ1         decimal.Decimal
	CJ2         decimal.Decimal
	CJ3         decimal.Decimal
	SN1         decimal.Decimal
	SN2         decimal.Decimal
	SN3         decimal.Decimal
	Total       decimal.Decimal
	URL         string
}

func QueryForNames(db *sql.DB, name string) ([]Lifter, error) {
	rows, err := db.Query(`SELECT lifter, name FROM results WHERE lifter like ?`, name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// get the length of results
	lifters := make([]Lifter, 100)
	for rows.Next() {
		var l Lifter
		err = rows.Scan(&l.Name, &l.Hometown)
		if err != nil {
			log.Fatal(err)
		}
		lifters = append(lifters, l)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return lifters, nil
}

func QueryResults(db *sql.DB, name, weightclass string) ([]Result, error) {
	rows, err := db.Query(`SELECT * FROM results WHERE name = ? and weightclass = ? ORDER BY date ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Result, 100)
	for rows.Next() {
		var r Result
		err = rows.Scan(&r.Date, &r.MeetName, &r.Lifter, &r.WeightClass, &r.Hometown, &r.CJ1, &r.CJ2, &r.CJ3, &r.SN1, &r.SN2, &r.SN3, &r.Total, &r.URL)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	db := buildDB()
	defer db.Close()
}
