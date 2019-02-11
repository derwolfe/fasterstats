package api

import (
	"html/template"
	"log"
	"net/http"

	"gitlab.com/derwolfe/faststats/db"
)

// API private struct for shared state.
type API struct {
	db          *db.OurDB
	searchPage  *template.Template
	namesPage   *template.Template
	liftersPage *template.Template
}

// NewAPI returns an api that can be used to process http requests
func NewAPI(db *db.OurDB) *API {
	// results
	lifts := template.Must(template.New("liftingResults").Parse(liftingResults))
	lifts.Parse(css)
	lifts.Parse(navbar)
	lifts.Parse(searchForm)
	lifts.Parse(resultsTable)

	// names
	names := template.Must(template.New("liftingResults").Parse(liftingResults))
	names.Parse(searchForm)
	names.Parse(navbar)
	names.Parse(css)
	names.Parse(searchNamesResults)

	// search form
	search := template.Must(template.New("landingPage").Parse(landingPage))
	search.Parse(css)
	search.Parse(searchForm)
	search.Parse(searchNamesResults)

	return &API{db: db, searchPage: search, namesPage: names, liftersPage: lifts}
}

// Search parses query parameters for name and returns a list of names
func (a API) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		// this needs validation! should be characters, maybe a digit, spaces
		name := r.FormValue("name")
		// this could be allowed and use pagination
		if len(name) < 3 {
			// this should be better!
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("400 - Search name must be greater than 3 characters"))
			return
		}
		found, err := a.db.QueryNames(name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Uh oh"))
			return
		}

		if err := a.namesPage.Execute(w, found); err != nil {
			log.Printf("%v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// SearchForm is the landing page and displays the search form.
func (a API) SearchForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := a.searchPage.Execute(w, nil); err != nil {
			log.Printf("%v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (a API) Results(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		names, ok := r.URL.Query()["name"]
		if !ok || len(names) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 - Bad Request - Missing/too many name parameter!"))
			return
		}
		hometowns, ok := r.URL.Query()["hometown"]
		if !ok || len(hometowns) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 - Bad Request - Missing/too many hometown parameter!"))
			return
		}
		// this will produce errors! what if the lifter has no results and someone modifies the search query
		found, err := a.db.QueryResults(names[0], hometowns[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Uh oh"))
			return
		}
		// lifts
		if err := a.liftersPage.Execute(w, found); err != nil {
			log.Printf("%v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

var css = `{{ define "css" }}
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.2.1/css/bootstrap.min.css" integrity="sha384-GJzZqFGwb1QTTN6wy59ffF1BuGJpLSa9DkKMp0DgiMDm4iYMj70gZWKYbI706tWS" crossorigin="anonymous">
<style>
.align-items-center {
	display: flex;
	align-items: center;  /*Aligns vertically center */
	justify-content: center; /*Aligns horizontally center */
  }
  </style>
{{ end }}`

var searchNamesResults = `{{ define "content" }}<div class="w-75 p-3 mx-auto">
	{{ if eq (len .) 0 }}
		<p>No names found</p>
	{{ end }}
	{{ if gt (len .) 50 }}
		<p>Too many results, please provide more letters</p>
	{{ end }}
	{{ if and (gt (len .) 0) (lt (len .) 50) }}
		<ul class="list-group">
			{{ range .}}
				<a href="results?name={{ .Name }}&hometown={{ .Hometown }}">
					<li class="list-group-item">{{ .Name }} - {{ .Hometown }}</li>
				</a>
			{{ end }}
		</ul>
	{{ end }}
</div>{{ end }}`

var resultsTable = `{{ define "content" }}
<div class="mx-auto p-3">
	<h3>{{ .Lifter }}</h3>
	<h4>{{ .Hometown }}</h4>
	<ul>
		<li>CJ: {{ .BestCJ }}</li>
		<li>Snatch: {{ .BestSN }}</li>
		<li>Total: {{ .BestTotal }}</li>
		<li>Avg SNs made: {{ .AvgSNMakes }}%</li>
		<li>Avg CJs made: {{ .AvgCJMakes }}%</li>
	</ul>
<div class="table-responsive">
	<table class="table table-striped">
		<thead class="thead-light">
			<tr>
				<th scope="col">Meet Date</th>
				<th scope="col">Meet</th>
				<th scope="col">Class</th>
				<th scope="col">Weight</th>
				<th scope="col">SN1</th>
				<th scope="col">SN2</th>
				<th scope="col">SN3</th>
				<th scope="col">CJ1</th>
				<th scope="col">CJ2</th>
				<th scope="col">CJ3</th>
				<th scope="col">Total</th>
				<th scope="col">Best SN</th>
				<th scope="col">Best CJ</th>
				<th scope="col">SNs/3</th>
				<th scope="col">CJs/3</th>
			</tr>
		</thead>
		<tbody class="table-striped">
		{{ range .Results }}
			{{ if .BestResult }}
			<tr bgcolor="lime">
			{{ else }}
			<tr>
			{{ end }}
				<td scope="row">{{ .Date }}</td>
				<td><a rel="noopener noreferrer" target="_blank" href="{{ .URL }}&isPopup=&Tab=Results">{{ .MeetName }}</a></td>
				<td>{{ .Weightclass }}</td>
				<td>{{ .CompetitionWeight }}</td>
				<td>{{ .SN1 }}</td>
				<td>{{ .SN2 }}</td>
				<td>{{ .SN3 }}</td>
				<td>{{ .CJ1 }}</td>
				<td>{{ .CJ2 }}</td>
				<td>{{ .CJ3 }}</td>
				<td>{{ .Total }}</td>
				<td>{{ .BestSN }}</td>
				<td>{{ .BestCJ }}</td>
				<td>{{ .SNSMade }}</td>
				<td>{{ .CJSMade }}</td>
			</tr>
			{{ end }}
		</tbody>
	</table>
</div>
{{ end }}`

var searchForm = `{{define "searchForm" }}<form class="form-inline" action="/search" method="GET">
	<input class="form-control mr-sm-2" name="name" type="search" placeholder="Search" aria-label="Search" required minlength=3>
	<button class="btn btn-outline-success my-2 my-sm-0" type="submit" value="Search">Search</button>
</form>{{ end}}`

var navbar = `{{define "navbar" }}
<nav class="navbar navbar-light bg-light">
	<a class="navbar-brand">bitofapressout.com</a>
	{{ template "searchForm" }}
</nav>{{ end }}`

var liftingResults = `<!doctype html>
<html>
	<head>
		<title>bitofapressout.com</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{ template "css"}}
	</head>
	<body>
		<div class="container-fluid">
			{{ template "navbar" }}
			{{ template "content" .}}
		</div>
	</body>
</html>`

// this should be used inside of another template, not sure how to do that now
var landingPage = `<!doctype html>
<html>
	<head>
		<title>bitofapressout.com</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{ template "css" }}
	</head>
	<body>
		<div class="container">
			<div class="d-flex flex-column align-content-center">
				<h2 class="display align-self-center">bitofapressout</h1>
				<p class="align-self-center">Enter a name, find a lifer from scraped USAW meet data</p>
				<form class="form-inline align-self-center" action="/search" method="GET">
					<input class="form-control mr-sm-2" name="name" type="search" placeholder="part of a name" aria-label="Search" required minlength=3>
					<button class="btn btn-outline-success my-2 my-sm-0" type="submit" value="Search">Search</button>
				</form>
			</div>
		</div>
	</body>
</html>`
