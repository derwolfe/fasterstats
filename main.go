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
				Lifter:<input type="string" name="name" required minlength=3>
				<input type="submit" value="Search!">
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
				Lifter:<input type="string" name="name" required minlength=3>
				<input type="submit" value="Search!">
			</form>
			<div class="table-responsive">
            <table class="table table-striped w-auto">
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
                    </tr>
                </thead>
                <tbody>
                    {{ range . }}
                    <tr>
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
                    </tr>
                    {{ end }}
                </tbody>
            </table>
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
			Lifter:<input type="string" name="name" required minlength=3>
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
