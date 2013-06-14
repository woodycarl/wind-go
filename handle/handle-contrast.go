package handle

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	Info("=== Handle Contrast ===")
	timeS := time.Now()

	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.S

	db := wind.DB(data.D1).Filter("Time >=", data.RD[0]["Time"]).Filter("Time <=", data.RD[len(data.RD)-1]["Time"])

	var ds []ContrastData
	for _, v := range data.RD {
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

	go saveCData(id, s.SensorsR, ds)

	page := Page{
		"id": id,
		"ds": ds,
		"s":  s.SensorsR,
	}

	page.render("contrast", w)

	Info("=== End Handle Contrast", time.Now().Sub(timeS), "===")
}

func saveCData(id string, s []wind.Sensor, ds []ContrastData) {
	timeS := time.Now()
	var lines []string
	line := "Time"

	for _, v := range s {
		line = line + "\tCh" + v.Channel + "-O" + "\tCh" + v.Channel + "-N" + "\tModified"
	}

	lines = append(lines, line)

	for _, v := range ds {
		line = v.Time.Format(DATA_DATE_FORMAT)

		for _, v1 := range v.Data {
			line = line + "\t" + fmt.Sprintf("%0.2f", v1.O) + "\t" + fmt.Sprintf("%0.2f", v1.N)
			if v1.M {
				line = line + "\t" + "*"
			} else {
				line = line + "\t"
			}
		}

		lines = append(lines, line)
	}

	err := writeLines(lines, OUTPUT_DIR+id+"/"+CONTRAST_FILE_NAME)
	if err != nil {
		Error("saveContrastData", err)
		return
	}
	Info("saveContrastData", time.Now().Sub(timeS))
}
