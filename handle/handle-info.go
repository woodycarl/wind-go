package handle

import (
	"net/http"

	"github.com/gorilla/mux"
)

func handleInfo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cat := mux.Vars(r)["cat"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	page := Page{
		"id": id,
	}

	switch cat {
	case "info-site":
		page["site"] = s.Site
	case "info-logger":
		page["logger"] = s.Logger
	case "info-sensors":
		page["sensors"] = s.SensorsR
		page["system"] = s.System
		page["version"] = s.Version
	case "integrity":
		page["start"] = s.Cm[0].My
		page["end"] = s.Cm[11].My
		page["cm"] = s.Cm
	case "integrity-all":
		type Years struct {
			Year float64
			Num  int
		}
		years := map[int]Years{}

		for _, v := range s.Am {
			years[int(v.Year)] = Years{
				Year: v.Year,
				Num:  years[int(v.Year)].Num + 1,
			}
		}
		page["years"] = years
		page["am"] = s.Am
	}

	page.render(cat, w)
}
