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
<!-- UIkit CSS -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/css/uikit.min.css" />

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
  <div class="uk-card uk-card-default">
  	<div class="uk-card-body">
		<h5 class="uk-card-title">{{ .Lifter }}</h5>
		<h6>{{ .Hometown }}</h6>
		<ul class="">
			<li class="">Best CJ: {{ .BestCJ }}</li>
			<li class="">Best Snatch: {{ .BestSN }}</li>
			<li class="">Best Total: {{ .BestTotal }}</li>
			<li class="">Most recent weight: {{ .RecentWeight }}
			<li class="">Avg SNs made: {{ .AvgSNMakes }}%</li>
			<li class="">Avg CJs made: {{ .AvgCJMakes }}%</li>
		</ul>
		<div class="uk-overflow-auto">
			<table class="uk-table uk-table-divider">
				<thead>
					<tr>
						<th>Meet Date</th>
						<th>Meet</th>
						<th>Class</th>
						<th>Weight</th>
						<th>SN1</th>
						<th>SN2</th>
						<th>SN3</th>
						<th>CJ1</th>
						<th>CJ2</th>
						<th>CJ3</th>
						<th>Total</th>
						<th>Best SN</th>
						<th>Best CJ</th>
						<th>SNs/3</th>
						<th>CJs/3</th>
					</tr>
				</thead>
				<tbody>
				{{ range .Results }}
					{{ if .BestResult }}
					<tr>
					{{ else }}
					<tr>
					{{ end }}
						<td data-label="Meet Date">{{ .Date }}</td>
						<td data-label="Name"><a rel="noopener noreferrer" target="_blank" href="{{ .URL }}&isPopup=&Tab=Results">{{ .MeetName }}</a></td>
						<td data-label="Weight Class">{{ .Weightclass }}</td>
						<td data-label="Competition Weight">{{ .CompetitionWeight }}</td>
						<td data-label="SN1">{{ .SN1 }}</td>
						<td data-label="SN2">{{ .SN2 }}</td>
						<td data-label="SN3">{{ .SN3 }}</td>
						<td data-label="CJ1">{{ .CJ1 }}</td>
						<td data-label="CJ2">{{ .CJ2 }}</td>
						<td data-label="CJ3">{{ .CJ3 }}</td>
						<td data-label="Total">{{ .Total }}</td>
						<td data-label="Best Snatch">{{ .BestSN }}</td>
						<td data-label="Best CJ">{{ .BestCJ }}</td>
						<td data-label="# Snatches made">{{ .SNSMade }}</td>
						<td data-label="# CJs made">{{ .CJSMade }}</td>
					</tr>
					{{ end }}
				</tbody>
			</table>
		</div>
	</div>
</div>
{{ end }}`

var searchForm = `{{define "searchForm" }}<form class="form-inline" action="/search" method="GET">
	<input class="form-control mr-sm-2" name="name" type="search" placeholder="Search" aria-label="Search" required minlength=3>
	<button class="btn btn-outline-success my-2 my-sm-0" type="submit" value="Search">Search</button>
</form>{{ end}}`

var navbar = `{{define "navbar" }}
<nav class="navbar navbar-light bg-light">
	<a class="navbar-brand" href="/">bitofapressout.com</a>
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
		<div class="uk-container">
			{{ template "navbar" }}
			{{ template "content" .}}
		</div>

		<!-- UIkit JS -->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit-icons.min.js"></script>
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
		<div class="uk-container">
			<div class="d-flex flex-column align-content-center">
				<h2 class="display align-self-center">bitofapressout</h1>
				<p class="align-self-center">Enter a name, find a lifer from scraped USAW meet data</p>
				<form class="form-inline align-self-center" action="/search" method="GET">
					<input class="form-control mr-sm-2" name="name" type="search" placeholder="part of a name" aria-label="Search" required minlength=3>
					<button class="btn btn-outline-success my-2 my-sm-0" type="submit" value="Search">Search</button>
				</form>
			</div>
		</div>
		<!-- UIkit JS -->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit-icons.min.js"></script>
	</body>
</html>`
