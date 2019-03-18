package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type OurDB struct {
	db *sql.DB
}

func BuildDB(dbPath string) (*OurDB, error) {
	// mark the connection as read only!

	// make sure this has been run!
	// create index idx_lifter_hometown on results(lifter, hometown);
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	return &OurDB{
		db: db,
	}, nil
}

func (o *OurDB) Close() {
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

func (r *Result) missesToMakes() {
	r.CJSMade = decimal.New(int64(max(0, r.CJ1.Sign())+max(0, r.CJ2.Sign())+max(0, r.CJ3.Sign())), 0)
	r.SNSMade = decimal.New(int64(max(0, r.SN1.Sign())+max(0, r.SN2.Sign())+max(0, r.SN3.Sign())), 0)
}

type ResultsSummary struct {
	Lifter       string
	IWFFirstName string
	IWFLastName  string
	Hometown     string
	BestCJ       decimal.Decimal
	BestSN       decimal.Decimal
	BestTotal    decimal.Decimal
	AvgCJMakes   decimal.Decimal
	AvgSNMakes   decimal.Decimal
	RecentWeight decimal.Decimal
	Results      []*Result
}

type PageInfo struct {
	Display int
}

type LiftersResponse struct {
	Lifters    []Lifter
	Name       string
	Total      int64
	Pages      []PageInfo
	Current    int64
	TotalPages int64
}

func (o *OurDB) QueryNames(name, offset string) (*LiftersResponse, error) {
	log.Printf("name: %v, offset: %v\n", name, offset)
	nameLike := "%" + strings.Replace(name, " ", "%", -1) + "%"

	// get the number of results so we can compute pages. Max result number is 50 per page.
	var total int64
	err := o.db.QueryRow(`SELECT IFNULL(SUM(ct), 0) from (SELECT 1 as ct FROM results WHERE lifter like $1 GROUP BY hometown, lifter)`, nameLike).Scan(&total)
	if err != nil {
		fmt.Printf("cw: %v", err)
		return nil, err
	}

	// if we found nothing return nothing and stop
	if total == 0 {
		resp := &LiftersResponse{
			Lifters:    nil,
			Name:       name,
			Total:      0,
			Current:    0,
			TotalPages: 0,
			Pages:      nil,
		}
		return resp, nil
	}

	// get the offset
	pageLimit := int64(50)
	var onum int64
	if len(offset) != 0 {
		onum, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			log.Printf("failed to parse offset %v", offset)
			// go to page 0
			onum = int64(1)
		}
	} else {
		onum = int64(1)
	}
	totalThisPage := getPageSize(onum, total, pageLimit)

	// page is meant to be min 1 for humans, offset is internal and should be 0-based
	if onum >= 1 {
		onum--
	}
	// get the names
	rows, err := o.db.Query(`SELECT DISTINCT lifter, hometown FROM results WHERE lifter like $1 ORDER BY lifter ASC LIMIT $2 OFFSET $3`, nameLike, pageLimit, onum*pageLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lifters := make([]Lifter, totalThisPage, totalThisPage)
	ct := 0
	for rows.Next() {
		l := Lifter{}
		err = rows.Scan(&l.Name, &l.Hometown)
		if err != nil {
			return nil, err
		}
		lifters[ct] = l
		ct++
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// there are 50 per page
	numPages := total / pageLimit
	pages := makePageInfoRange(int(numPages))

	// total is the number of pages
	// current is the page being returned, if this had an offset, it would be the next page
	resp := &LiftersResponse{
		Lifters:    lifters,
		Total:      total,
		Current:    onum + 1,
		Name:       name,
		Pages:      pages,
		TotalPages: int64(len(pages)),
	}
	return resp, nil
}

func getPageSize(pageNum, total, limit int64) int64 {
	if pageNum < 1 {
		panic("offset must be positive")
	}
	numPages := int64(math.Ceil(float64(total) / float64(limit)))
	if pageNum < numPages {
		return limit
	}
	rem := total % limit
	if total > 0 && rem == 0 {
		return limit
	}
	return rem
}

func (o *OurDB) QueryResults(name, hometown string) (*ResultsSummary, error) {
	log.Printf("name: %v, hometown: %v\n", name, hometown)
	// check results count
	var resultCt int64
	err := o.db.QueryRow(`SELECT IFNULL(SUM(ct), 0) from (SELECT 1 as ct FROM results WHERE lifter = $1 and hometown = $2)`, name, hometown).Scan(&resultCt)
	if err != nil {
		return nil, err
	}

	rows, err := o.db.Query(`SELECT date, meet_name, lifter, weight_class, competition_weight, hometown, cj1, cj2, cj3, sn1, sn2, sn3, total, best_snatch, best_cleanjerk, url FROM results WHERE lifter = $1 and hometown = $2 ORDER BY date DESC`, name, hometown)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// load the results
	results := make([]*Result, resultCt, resultCt)

	ct := 0
	for rows.Next() {
		r := &Result{}
		err = rows.Scan(&r.Date, &r.MeetName, &r.Lifter, &r.Weightclass, &r.CompetitionWeight, &r.Hometown, &r.CJ1, &r.CJ2, &r.CJ3, &r.SN1, &r.SN2, &r.SN3, &r.Total, &r.BestSN, &r.BestCJ, &r.URL)
		if err != nil {
			return nil, err
		}

		// compute misses an makes
		r.BestResult = false
		r.missesToMakes()
		results[ct] = r
		ct++
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
	err = o.db.QueryRow(`select max(total) from results where lifter = $1 and hometown = $2`, name, hometown).Scan(&rs.BestTotal)
	if err != nil {
		return nil, err
	}
	// load the best SN
	err = o.db.QueryRow(`select max(best_snatch) from results where lifter = $1 and hometown = $2`, name, hometown).Scan(&rs.BestSN)
	if err != nil {
		return nil, err
	}
	// load the best CJ
	err = o.db.QueryRow(`select max(best_cleanjerk) from results where lifter = $1 and hometown = $2`, name, hometown).Scan(&rs.BestCJ)
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
	rs.IWFFirstName, rs.IWFLastName = ToIWFName(results[0].Lifter)

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

func makePageInfoRange(max int) []PageInfo {
	// pagination always starts at 1
	a := make([]PageInfo, max+1)
	for i := range a {
		a[i] = PageInfo{
			Display: i + 1,
		}
	}
	return a
}

var nameReg = regexp.MustCompile("[^a-z ]+")

// ToIWFName returns the name needed to perform a search against the IWF
func ToIWFName(name string) (string, string) {
	lowered := strings.ToLower(name)
	lowered = nameReg.ReplaceAllString(lowered, "")

	ss := strings.Split(lowered, " ")
	last := ss[len(ss)-1]

	if len(ss) > 1 {
		return ss[0], last
	}
	return lowered, ""
}
