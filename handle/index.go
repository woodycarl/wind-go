package handle

import (
	"net/http"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	page := Page {
		"hideResultMenu": true,
	}

	page.render("index", w)
}