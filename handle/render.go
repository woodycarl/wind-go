package handle

import (
	"net/http"
	"text/template"
)

type Page map[string]interface{}

func (page *Page) render(file string, w http.ResponseWriter) {
	base := "template/"

	tmpl, err := template.New("main.html").Funcs(funcMap).ParseFiles(
		base+"main.html",
		base+file+".html",
	)

	if err != nil {
		handleErr(w, err)
		return
	}

	if err = tmpl.Execute(w, page); err != nil {
		handleErr(w, err)
		return
	}
}
