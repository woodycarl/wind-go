package handle

import (
	"text/template"
	"net/http"
)

type Page map[string]interface{}

func (page *Page) render(file string, w http.ResponseWriter) {
	base := "template/"

	tmpl, err := template.New("main.html").Funcs(funcMap).ParseFiles(
		base + "main.html",
		base + file + ".html",
	)

	if err != nil {
		// serveError(w, err)
		return
	}

	if err = tmpl.Execute(w, page); err != nil {
		// serveError(w, err)
		return
	}
}