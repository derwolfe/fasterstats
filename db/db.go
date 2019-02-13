package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"log"
	"strings"
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

	resultsStmt, err := db.Prepare(`SELECT date, meet_name, lifter, weight_class, competition_weight, hometown, cj1, cj2, cj3, sn1, sn2, sn3, total, best_snatch, best_cleanjerk, url FROM results WHERE lifter = $1 and hometown = $2 ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}

	bestCJ, err := db.Prepare(`select max(best_cleanjerk) from results where lifter = $1 and hometown = $2`)

	if err != nil {
		return nil, err
	}

	bestSN, err := db.Prepare(`select max(best_snatch) from results where lifter = $1 and hometown = $2`)

	if err != nil {
		return nil, err
	}

	bestTotal, err := db.Prepare(`select MAX(total) from results where lifter = $1 and hometown = $2`)

	if err != nil {
		return nil, err
	}

	return &OurDB{
		db:             db,
		nameQuery:      nameStmt,
		resultsQuery:   resultsStmt,
		bestCJQuery:    bestCJ,
		bestSNQuery:    bestSN,
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
	Date              string
	MeetName          string
	Lifter            string
	Weightclass       string
	CompetitionWeight decimal.Decimal
	Hometown          string
	CJ1               decimal.Decimal
	CJ2               decimal.Decimal
	CJ3               decimal.Decimal
	SN1               decimal.Decimal
	SN2               decimal.Decimal
	SN3               decimal.Decimal
	Total             decimal.Decimal
	BestCJ            decimal.Decimal
	BestSN            decimal.Decimal
	URL               string
	CJSMade           decimal.Decimal
	SNSMade           decimal.Decimal
	BestResult        bool
}

func (r *Result) shortenMeetName() {
	runes := []rune(r.MeetName)
	maxLen := 30
	if len(runes) > (maxLen + 1) {
		r.MeetName = fmt.Sprintf("%s ...", string(r.MeetName[0:maxLen]))
	}
}

func (r *Result) missesToMakes() {
	r.CJSMade = decimal.New(int64(max(0, r.CJ1.Sign())+max(0, r.CJ2.Sign())+max(0, r.CJ3.Sign())), 0)
	r.SNSMade = decimal.New(int64(max(0, r.SN1.Sign())+max(0, r.SN2.Sign())+max(0, r.SN3.Sign())), 0)
}

type ResultsSummary struct {
	Lifter       string
	Hometown     string
	BestCJ       decimal.Decimal
	BestSN       decimal.Decimal
	BestTotal    decimal.Decimal
	AvgCJMakes   decimal.Decimal
	AvgSNMakes   decimal.Decimal
	RecentWeight decimal.Decimal
	Results      []*Result
}

type OurDB struct {
	db             *sql.DB
	nameQuery      *sql.Stmt
	resultsQuery   *sql.Stmt
	bestCJQuery    *sql.Stmt
	bestSNQuery    *sql.Stmt
	bestTotalQuery *sql.Stmt
}

func (o *OurDB) QueryNames(name string) ([]Lifter, error) {
	log.Printf("name: %v\n", name)
	name = strings.Replace(name, " ", "%", -1)
	q := "%" + name + "%"
	rows, err := o.nameQuery.Query(q)
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
	log.Printf("name: %v, hometown: %v\n", name, hometown)
	rows, err := o.resultsQuery.Query(name, hometown)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// load the results
	var results []*Result
	for rows.Next() {
		r := &Result{}
		err = rows.Scan(&r.Date, &r.MeetName, &r.Lifter, &r.Weightclass, &r.CompetitionWeight, &r.Hometown, &r.CJ1, &r.CJ2, &r.CJ3, &r.SN1, &r.SN2, &r.SN3, &r.Total, &r.BestSN, &r.BestCJ, &r.URL)
		if err != nil {
			return nil, err
		}

		// compute misses an makes
		r.BestResult = false
		r.missesToMakes()
		r.shortenMeetName()
		results = append(results, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("No results")
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
	totalCJs := decimal.Zero
	totalSNs := decimal.Zero

	numLiftsBase := decimal.New(int64(len(results)*3), 1)

	for _, r := range results {
		// update the avg made
		totalSNs = r.SNSMade.Add(totalSNs)
		totalCJs = r.CJSMade.Add(totalCJs)

		// find the bests over the entire result set
		if r.BestCJ.Equal(rs.BestCJ) {
			r.BestResult = true
		}
		if r.BestSN.Equal(rs.BestSN) {
			r.BestResult = true
		}
		if r.Total.Equal(rs.BestTotal) {
			r.BestResult = true
		}
	}
	factor := decimal.New(100, 1)

	rs.AvgSNMakes = totalSNs.DivRound(numLiftsBase, 5).Mul(factor)
	rs.AvgCJMakes = totalCJs.DivRound(numLiftsBase, 5).Mul(factor)

	// we shouldn't get here if there are no results
	rs.Lifter = results[0].Lifter
	rs.Hometown = results[0].Hometown
	rs.RecentWeight = results[0].CompetitionWeight

	return &rs, nil
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func maxDec(x, y decimal.Decimal) decimal.Decimal {
	if x.GreaterThan(y) {
		return x
	}
	return y
}
