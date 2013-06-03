package handle

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	//. "github.com/woodycarl/wind-go/logger"
	"github.com/woodycarl/wind-go/wind"
)

type ContrastDD struct {
	O float64
	N float64
	M bool
}

type ContrastData struct {
	Time time.Time
	Data []ContrastDD
}

func handleContrast(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	db := wind.DB(data.Data1h).Filter("Time >=", data.RData[0]["Time"]).Filter("Time <=", data.RData[len(data.RData)-1]["Time"])

	var ds []ContrastData
	for _, v := range data.RData {
		d := ContrastData{
			Time: time.Unix(int64(v["Time"]), 0),
		}

		od := db.Filter("Time =", v["Time"])

		for _, v1 := range s.SensorsR {
			/*if v1.Description == "No SCM Installed" {
				continue
			}*/

			ch := v1.Channel

			dd := ContrastDD{
				N: v["ChAvg"+ch],
			}

			if len(od) > 0 {
				dd.O = od[0]["ChAvg"+ch]
			}

			if dd.N != dd.O {
				dd.M = true
			}

			d.Data = append(d.Data, dd)
		}

		ds = append(ds, d)
	}

	page := Page{
		"id": id,
		"ds": ds,
		"s":  s.SensorsR,
	}

	page.render("contrast", w)
}
