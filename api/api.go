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
	aboutPage   *template.Template
}

// NewAPI returns an api that can be used to process http requests
func NewAPI(db *db.OurDB) *API {
	// results
	lifts := template.Must(template.New("liftingResults").Parse(liftingResults))
	lifts.Parse(css)
	lifts.Parse(resultsTable)

	// names
	names := template.Must(template.New("liftingResults").Parse(liftingResults))
	names.Parse(css)
	names.Parse(searchNamesResults)

	// search form
	search := template.Must(template.New("landingPage").Parse(landingPage))
	search.Parse(css)
	search.Parse(searchNamesResults)

	about := template.Must(template.New("about").Parse(liftingResults))
	about.Parse(css)
	about.Parse(aboutPage)

	return &API{db: db, searchPage: search, namesPage: names, liftersPage: lifts, aboutPage: about}
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
		// there might be a page; if so try to use it to look up
		// this should always be a number XXX add validtion
		offset := r.FormValue("page")
		found, err := a.db.QueryNames(name, offset)

		if err != nil {
			log.Printf("error fetching names: %v", err)
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

func (a API) About(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := a.aboutPage.Execute(w, nil); err != nil {
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

var aboutPage = `{{ define "content"}}
<div class="uk-margin" uk-margin>
	<form class="uk-form" action="/search" method="GET" uk-form>
		<input class="uk-input uk-form-width-large" name="name" type="search" placeholder="Find a lifter by name" value="{{ .Name }}" required minlength=3 autofocus>
		<button class="uk-button uk-button-default" type="submit" value="Search">Search</button>
	</form>
</div>
<article class="uk-article">
	<h1 class="uk-article">About</h1>
	<p class="uk-text-lead">I like Olympic Weightlifting statistics. If you're here, you probably do too.</p>
	<p class="uk-article-text">
  If you look at even a little bit of data from this site you'll find some
  errors. Maybe a total doesn't add up or a lifter's best snatch is 40 kg
  greater than their best ever clean & jerk. This is an artifact of the data
  that powers this site having originated from USAW lifting data. To try to
  make it easier to reconcile whether the USAW has incorrect data versus this
  site, every result links back to the original data from the USAW site. This
  link will show up in the <span style="font-weight: bold">MEET (USAW
  LINK)</span> column of the results table.
	</p>
</article>
{{ end }}`

var css = `{{ define "css" }}
<!-- UIkit CSS -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/css/uikit.min.css" />
<style>
body {
	font-family: sans-serif;
	color: #1C242E;
}
.best-row {
	font-weight:bold;
}
</style>
{{ end }}`

var searchNamesResults = `{{ define "content" }}
<div class="uk-margin" uk-margin>
	<form class="uk-form" action="/search" method="GET" uk-form>
		<input class="uk-input uk-form-width-large" name="name" type="search" placeholder="Find a lifter by name" value="{{ .Name }}" required minlength=3 autofocus>
		<button class="uk-button uk-button-default" type="submit" value="Search">Search</button>
	</form>
</div>

<div class="uk-card">
	{{ if eq .Total 0 }}
		<p> No lifters found</p>
	{{ else }}
		<p>Found {{ .Total }} matching lifters</li>

		<hr>

		<div>
			{{ range .Lifters }}
				<a href="results?name={{ .Name }}&hometown={{ .Hometown }}">
					<div class="uk-card">
						<h4 class="uk-card-title">{{ .Name }} - {{ .Hometown }}</h3>
					</div>
				</a>
			{{ end }}
		</div>

		<hr>

		{{ if (ne .TotalPages 1)}}
		<div>
			<ul class="uk-pagination uk-margin">
			{{ range .Pages }}
				{{ if (eq .Display $.Current)}}
					<li class="uk-active">
				{{ else }}
					<li>
				{{ end }}
					<a href="search?name={{ $.Name }}&page={{ .Display }}">{{ .Display }}</a>
				</li>
			{{ end }}
			</ul>
		</div>
		{{ end }}

	{{ end }}
</div>{{ end }}`

var resultsTable = `{{ define "content" }}
<div class="uk-margin" uk-margin>
	<form class="uk-form" action="/search" method="GET" uk-form>
		<input class="uk-input uk-form-width-large" name="name" type="search" placeholder="Find a lifter by name" required minlength=3 autofocus>
		<button class="uk-button uk-button-default" type="submit" value="Search">Search</button>
	</form>
</div>

<article class="uk-article">
	<h1 class="uk-article-title">{{ .Lifter }} / {{ .Hometown }}</h1>
	<h3>Statistics</h3>
	<div class="uk-grid-divider uk-child-width-expand@s" uk-grid>
		<div>
			<ul class="uk-list">
				<li>Best CJ: {{ .BestCJ }} kg</li>
				<li>Best Snatch: {{ .BestSN }} kg </li>
				<li>Best Total: {{ .BestTotal }} kg</li>
			<ul>
		</div>
		<div>
			<ul class="uk-list">
				<li>Most recent weight: {{ .RecentWeight }} kg</li>
				<li>Avg # Snatches made: {{ .AvgSNMakes }}%</li>
				<li>Avg # Clean & Jerks made: {{ .AvgCJMakes }}%</li>
			</ul>
		</div>
	</div>
	<h3>Competitions</h3>
	<p><a rel="noopener noreferrer" target="_blank" href="https://www.iwf.net/new_bw/results_by_events/?athlete_name={{ .IWFLastName }}+{{ .IWFFirstName }}&athlete_gender=all&athlete_nation=USA">IWF results</a></p>
	<p class="uk-text-muted">*Bests are bolded</p>
	<div class="uk-overflow-auto">
		<table class="uk-table uk-table-divider uk-table-hover">
			<thead>
				<tr>
					<th class="uk-table-expand">Meet Date</th>
					<th class="uk-text-nowrap">Meet (USAW link)</th>
					<th class="uk-text-nowrap">Class@weight</th>
					<th>SN1</th>
					<th>SN2</th>
					<th>SN3</th>
					<th>CJ1</th>
					<th>CJ2</th>
					<th>CJ3</th>
					<th>Total</th>
					<th class="uk-text-nowrap">Best SN</th>
					<th class="uk-text-nowrap">Best CJ</th>
					<th class="uk-text-nowrap">SNs/3</th>
					<th class="uk-text-nowrap">CJs/3</th>
				</tr>
			</thead>
			<tbody>
			{{ range .Results }}
				{{ if .BestResult }}
				<tr class="best-row">
				{{ else }}
				<tr>
				{{ end }}
					<td data-label="Meet Date">{{ .Date }}</td>
					<td data-label="Name"><a rel="noopener noreferrer" target="_blank" href="{{ .URL }}&isPopup=&Tab=Results">{{ .MeetName }}</a></td>
					<td data-label="Weight Class">{{ .Weightclass }} @ {{ .CompetitionWeight }}</td>
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
{{ end }}`

var liftingResults = `<!doctype html>
<html>
	<head>
		<title>bitofapressout.com</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{ template "css"}}
	</head>
	<body>
		<div class="uk-container">
			<div class="uk-margin-top">
				{{ template "content" .}}
			</div>
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
			<div class="uk-position-center">
				<h2 class="">bitofapressout</h1>
				<p class="">Search USA Weightlifting data from 2012 onward. See <a href="/about">about</a> to learn more!</p>
				<div class="uk-margin" uk-margin>
					<form class="uk-form" action="/search" method="GET" uk-form>
						<input class="uk-input uk-form-width-large" name="name" type="search" placeholder="Find a lifter by name" required minlength=3 autofocus>
						<button class="uk-button uk-button-default" type="submit" value="Search">Search</button>
					</form>
				</div>
			</div>
		</div>
		<!-- UIkit JS -->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.0.3/js/uikit-icons.min.js"></script>
	</body>
</html>`
