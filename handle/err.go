package handle

import "net/http"

func handleErr(w http.ResponseWriter, err error) {
	page := Page{
		"err": err.Error(),
	}

	page.render("err", w)
}
