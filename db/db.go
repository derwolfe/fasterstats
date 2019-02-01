package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"log"
)

func BuildDB(dbPath string) (*OurDB, error) {
	// mark the connection as read only!

	// make sure this has been run!
	// create index idx_lifter_hometown on results(lifter, hometown);

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	nameStmt, err := db.Prepare(`SELECT DISTINCT lifter, hometown FROM results WHERE lifter like $1 ORDER BY lifter ASC LIMIT 100`)
	if err != nil {
		return nil, err
	}

	resultsStmt, err := db.Prepare(`SELECT date, meet_name, lifter, weight_class, hometown, cj1, cj2, cj3, sn1, sn2, sn3, total, url FROM results WHERE lifter = $1 and hometown = $2 ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}

	bestCJ, err := db.Prepare(`with bestCJ as (select cj1 from results where lifter = $1 and hometown = $2 UNION select cj2 from results where lifter = $1 and hometown = $2 UNION select cj3 from results where lifter = $1 and hometown = $2) select MAX(cj1) from bestCJ`)

	if err != nil {
		return nil, err
	}

	bestSN, err := db.Prepare(`with bestSN as (select sn1 from results where lifter = $1 and hometown = $2 UNION select sn2 from results where lifter = $1 and hometown = $2 UNION select sn3 from results where lifter = $1 and hometown = $2) select MAX(sn1) from bestSN`)

	if err != nil {
		return nil, err
	}

	bestTotal, err := db.Prepare(`select MAX(total) from results where lifter = $1 and hometown = $2`)

	if err != nil {
		return nil, err
	}

	return &OurDB{
		db:           db,
		nameQuery:    nameStmt,
		resultsQuery: resultsStmt,
		bestCJQuery: bestCJ,
		bestSNQuery: bestSN,
		bestTotalQuery: bestTotal,
	}, nil
}

func (o *OurDB) Close() {
	o.nameQuery.Close()
	o.resultsQuery.Close()
	o.bestCJQuery.Close()
	o.bestSNQuery.Close()
	o.bestTotalQuery.Close()
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
	Weightclass string
	Hometown    string
	CJ1         decimal.Decimal
	CJ2         decimal.Decimal
	CJ3         decimal.Decimal
	SN1         decimal.Decimal
	SN2         decimal.Decimal
	SN3         decimal.Decimal
	Total       decimal.Decimal
	URL         string
	CJSMade 	int
	SNSMade		int
}

func (r *Result) missesToMakes() {
	r.CJSMade = max(0, r.CJ1.Sign()) + max(0, r.CJ2.Sign()) + max(0, r.CJ3.Sign())
	r.SNSMade = max(0, r.SN1.Sign()) + max(0, r.SN2.Sign()) + max(0, r.SN3.Sign())
}

type ResultsSummary struct {
	BestCJ decimal.Decimal
	BestSN decimal.Decimal
	BestTotal  decimal.Decimal
	AvgCJMakes decimal.Decimal
	AvgSNMakes decimal.Decimal
	Results []*Result
}


type OurDB struct {
	db           *sql.DB
	nameQuery    *sql.Stmt
	resultsQuery *sql.Stmt
	bestCJQuery  *sql.Stmt
	bestSNQuery  *sql.Stmt
	bestTotalQuery *sql.Stmt
}

func (o *OurDB) QueryNames(name string) ([]Lifter, error) {
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

func (o *OurDB) QueryResults(name, hometown string) (*ResultsSummary, error) {
	rows, err := o.resultsQuery.Query(name, hometown)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// load the results
	var results []*Result
	for rows.Next() {
		r := &Result{}
		err = rows.Scan(&r.Date, &r.MeetName, &r.Lifter, &r.Weightclass, &r.Hometown, &r.CJ1, &r.CJ2, &r.CJ3, &r.SN1, &r.SN2, &r.SN3, &r.Total, &r.URL)
		if err != nil {
			return nil, err
		}

		// compute misses an makes
		r.missesToMakes()
		results = append(results, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	rs := ResultsSummary{Results: results}

	// load the total
	err = o.bestTotalQuery.QueryRow(name, hometown).Scan(&rs.BestTotal)
	if err != nil {
		return nil, err
	}
	// load the best SN
	err = o.bestSNQuery.QueryRow(name, hometown).Scan(&rs.BestSN)
	if err != nil {
		return nil, err
	}
	// load the best CJ
	err = o.bestCJQuery.QueryRow(name, hometown).Scan(&rs.BestCJ)
	if err != nil {
		return nil, err
	}

	// now compute avg CJ makes. Loop over the results, converting each to a 1 or -1
	return &rs, nil
}

func max(x, y int) int {
    if x > y {
        return x
    }
    return y
}