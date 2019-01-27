package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"log"
)

func BuildDB() (*OurDB, error) {
	// mark the connection as read only!
	db, err := sql.Open("sqlite3", "./results.db")
	if err != nil {
		log.Fatal(err)
	}

	nameStmt, err := db.Prepare(`SELECT DISTINCT lifter, hometown FROM results WHERE lifter like ? ORDER BY lifter ASC`)
	if err != nil {
		return nil, err
	}

	resultsStmt, err := db.Prepare(`SELECT * FROM results WHERE lifter = ? and weight_class = ? ORDER BY date ASC`)
	if err != nil {
		return nil, err
	}

	return &OurDB{
		db: db,
		nameQuery: nameStmt,
		resultsQuery: resultsStmt,
	}, nil
}

func (o *OurDB) Close() {
	o.nameQuery.Close()
	o.resultsQuery.Close()
	o.db.Close()
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

type OurDB struct {
	db *sql.DB
	nameQuery *sql.Stmt
	resultsQuery *sql.Stmt
}

func (o *OurDB) QueryForNames(name string) ([]Lifter, error) {
	rows, err := o.nameQuery.Query("%" + name + "%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// get the length of results
	var lifters []Lifter
	for rows.Next() {
		l := Lifter{}
		err = rows.Scan(&l.Name, &l.Hometown)
		if err != nil {
			return nil, err
		}
		lifters = append(lifters, l)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return lifters, nil
}

func (o *OurDB) QueryResults(name, weightclass string) ([]Result, error) {
	rows, err := o.resultsQuery.Query(name, weightclass)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		r := Result{}
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
