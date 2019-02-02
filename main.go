package main

import (
	"html/template"
	"log"
	"net/http"

	"gitlab.com/derwolfe/faststats/db"
)

type api struct {
	db *db.OurDB
	searchPage	*template.Template
	namesPage *template.Template
	liftersPage *template.Template
}

var css = `{{ define "css" }}<style>
.Aligner {
	display: flex;
	align-items: center;
	justify-content: center;
}

.bestRow { bgcolor: "limegreen"; }
</style>{{ end }}`

var searchNamesResults = `{{ define "content" }}<div>
		<ul>
			{{ range .}}
				<li><a href="results?name={{ .Name }}&hometown={{ .Hometown }}">{{ .Name }} - {{ .Hometown }}</li></a>
			{{ end }}
		</ul>
</div>{{ end }}`

var resultsTable = `{{ define "content" }}<div>
	<h4>Bests</h4>
	<ul>
		<li>CJ: {{ .BestCJ }}</li>
		<li>Snatch {{ .BestSN }}</li>
		<li>Total {{ .BestTotal }}</li>
	</ul>
	<h4>Averages (approx)</h4>
	<ul>
		<li>CJ: {{ .AvgCJMakes }}</li>
		<li>Snatch {{ .AvgSNMakes }}</li>
	</ul>
</div>
<div class="">
<table class="">
	<thead>
		<tr>
			<th scope="col">Meet Date</th>
			<th scope="col">Meet</th>
			<th scope="col">Class</th>
			<th scope="col">Lifter</th>
			<th scope="col">Hometown</th>
			<th scope="col">SN1</th>
			<th scope="col">SN2</th>
			<th scope="col">SN3</th>
			<th scope="col">CJ1</th>
			<th scope="col">CJ2</th>
			<th scope="col">CJ3</th>
			<th scope="col">Total</th>
			<th scope="col">SNs/3</th>
			<th scope="col">CJs/3</th>
		</tr>
	</thead>
	<tbody>
	{{ range .Results }}
		{{ if .BestResult }}
		<tr bgcolor="lime">
		{{ else }}
		<tr>
		{{ end }}
		<td>{{ .Date }}</td>
			<td><a rel="noopener noreferrer" target="_blank" href="{{ .URL }}&isPopup=&Tab=Results">{{ .MeetName }}</a></td>
			<td>{{ .Weightclass }}</td>
			<td>{{ .Lifter }}</td>
			<td>{{ .Hometown }}</td>
			<td>{{ .SN1 }}</td>
			<td>{{ .SN2 }}</td>
			<td>{{ .SN3 }}</td>
			<td>{{ .CJ1 }}</td>
			<td>{{ .CJ2 }}</td>
			<td>{{ .CJ3 }}</td>
			<td>{{ .Total }}</td>
			<td>{{ .SNSMade }}</td>
			<td>{{ .CJSMade }}</td>
		</tr>
		{{ end }}
	</tbody>
</table>
</div>{{ end }}`

var searchForm = `{{define "searchForm" }}<form action="/search" method="GET">
	Lifter:<input type="string" name="name" required minlength=3>
	<input type="submit" value="Search">
</form>{{ end }}`

var liftingResults = `<!doctype html>
<html>
	<head>
		<title>Lifter finder</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{ template "css"}}
	</head>
	<body>
		<div class="Aligner">
		{{ template "searchForm" }}
		{{ template "content" .}}
		</div>
	</body>
</html>`

// this should be used inside of another template, not sure how to do that now
var findLiftersForm = `<!doctype html>
<html>
	<head>
		<title>Lifter finder</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{ template "css" }}
	</head>
	<body>
		{{ template "searchForm" }}
	</body>
</html>`

func (a api) search(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		// this needs validation! should be characters, maybe a digit, spaces
		name := r.FormValue("name")
		// this could be allowed and use pagination
		if len(name) < 3 {
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
			log.Printf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (a api) searchForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := a.searchPage.Execute(w, nil); err != nil {
			log.Printf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (a api) results(w http.ResponseWriter, r *http.Request) {
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
		found, err := a.db.QueryResults(names[0], hometowns[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Uh oh"))
			return
		}
		// lifts
		if err := a.liftersPage.Execute(w, found); err != nil {
			log.Printf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func newAPI(db *db.OurDB) *api{
	//
	lifts := template.Must(template.New("liftingResults").Parse(liftingResults))
	lifts.Parse(css)
	lifts.Parse(searchForm)
	lifts.Parse(resultsTable)

	// names
	names := template.Must(template.New("liftingResults").Parse(liftingResults))
	names.Parse(searchForm)
	names.Parse(css)
	names.Parse(searchNamesResults)

	// search form
	search := template.Must(template.New("findLiftersForm").Parse(findLiftersForm))
	search.Parse(css)
	search.Parse(searchForm)
	search.Parse(searchNamesResults)

	return &api{db: db, searchPage: search, namesPage: names, liftersPage: lifts}
}

func main() {
	db, err := db.BuildDB("./results.db")
	if err != nil {
		log.Fatal(err)
	}
	api := newAPI(db)

	http.HandleFunc("/", api.searchForm)
	http.HandleFunc("/search", api.search)
	http.HandleFunc("/results", api.results)

	log.Println("Starting on :9090")
	err = http.ListenAndServe(":9090", nil) // setting listening port

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
