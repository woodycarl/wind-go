package handle

import (
	"net/http"

	"github.com/gorilla/mux"
)

type TurbData struct {
	Channel string
	Height  float64
	Turb    float64
}

func handleTurbs(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.S

	var turbs []TurbData

	for i, v := range s.Sensors["wv"] {
		turb := TurbData{
			Channel: v.Channel,
			Height:  v.Height,
			Turb:    s.Turbs[i],
		}

		turbs = append(turbs, turb)
	}

	page := Page{
		"id":    id,
		"turbs": turbs,
	}

	page.render("turbs", w)
}
