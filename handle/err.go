package handle

import (
	"net/http"

	. "github.com/woodycarl/wind-go/logger"
)

func handleErr(w http.ResponseWriter, err error) {
	Error(err)

	page := Page{
		"err": err.Error(),
	}

	page.render("err", w)
}
