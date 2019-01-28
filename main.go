package main

import (
	"log"
	"net/http"
	"html/template"
	"gitlab.com/derwolfe/faststats/db"
)

type api struct {
	db *db.OurDB
}

var searchNamesResults = `<!doctype html>
<html>
	<head>
		<title>Lifter finder</title>
    	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	</head>
	<body>
		<div>
			<form action="/search" method="post">
				Lifter:<input type="string" name="name">
				<input type="submit" value="">
			</form>
			<ul>
				{{ range .}}
					<li><a href="results?name={{ .Name }}&hometown={{ .Hometown }}">{{ .Name }} - {{ .Hometown }}</li></a>
				{{ end }}
			</ul>
		</div>
	</body>
</html>`

var liftingResults = `<!doctype html>
<html>
	<head>
		<title>Lifter finder</title>
    	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	</head>
	<body>
		<div>
			<form action="/search" method="post">
				Lifter:<input type="string" name="name">
				<input type="submit" value="">
			</form>
			<ul>
				{{ range .}}
					<li>{{ . }}</li>
				{{ end }}
			</ul>
		</div>
	</body>
</html>`

var searchNamesResultsTemplate = template.Must(template.New("names found").Parse(searchNamesResults))
var liftingResultsTemplate = template.Must(template.New("results found").Parse(liftingResults))

// this should be used inside of another template, not sure how to do that now
var findLiftersForm = []byte(`<!doctype html>
<html>
	<head>
		<title>Lifter finder</title>
    	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	</head>
	<body>
		<form action="/search" method="post">
			Lifter:<input type="string" name="name">
			<input type="submit" value="">
		</form>
	</body>
</html>`)

func (a api) search(w http.ResponseWriter, r *http.Request){
	// display a form, if a post, display results
	if r.Method == "POST" {
		r.ParseForm()
		// this needs validation! should be characters, maybe a digit, spaces
		name := r.FormValue("name")
		found, err := a.db.QueryNames(name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Uh oh"))
			return
		}
		searchNamesResultsTemplate.Execute(w, found)
		return
	}
	w.Write(findLiftersForm)
}

func (a api) results(w http.ResponseWriter, r *http.Request) {
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
	liftingResultsTemplate.Execute(w, found)
}

func main() {
	db, err := db.BuildDB("./results.db")
	if err != nil {
		log.Fatal(err)
	}
	api := api{db: db}

	http.HandleFunc("/search", api.search)
	http.HandleFunc("/results", api.results)

	err = http.ListenAndServe(":9090", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
