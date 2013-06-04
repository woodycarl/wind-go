package handle

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/woodycarl/wind-go/wind"
)

func handleLinest(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cat := mux.Vars(r)["cat"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	var title string
	switch cat {
	case "wv":
		title = "风速相关性"
	case "wd":
		title = "风向相关性"
	}

	page := Page{
		"id":      id,
		"sensors": s.Sensors[cat],
		"title":   title,
	}

	page.render("linest", w)
}

type LinestFigData struct {
	Ch1       string
	Height1   float64
	Ch2       string
	Height2   float64
	Data      string
	Max2      float64
	Slope     float64
	Intercept float64
	Rsq       float64
}

func handleLinestFig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cat := mux.Vars(r)["cat"]
	ch1 := mux.Vars(r)["ch"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	var sensor wind.Sensor
	for _, v := range s.Sensors[cat] {
		if v.Channel == ch1 {
			sensor = v
		}
	}

	linests := []LinestFigData{}

	db := wind.DB(data.Data1h)
	d1 := db.Get("ChAvg" + ch1)["ChAvg"+ch1]
	height1 := sensor.Height
	for _, v := range sensor.Rations {
		ch2 := v.Channel
		height2 := s.Sensors[cat][v.Index].Height
		d2 := db.Get("ChAvg" + ch2)["ChAvg"+ch2]
		var d [][]float64
		max2 := 0.0
		for i, _ := range d1 {
			dt := []float64{d2[i], d1[i]}
			d = append(d, dt)
			if max2 < d2[i] {
				max2 = d2[i]
			}
		}

		b, err := json.Marshal(d)
		if err != nil {
			fmt.Println("error:", err)
		}

		linest := LinestFigData{
			Ch1:       ch1,
			Height1:   height1,
			Ch2:       ch2,
			Height2:   height2,
			Data:      string(b),
			Max2:      max2,
			Slope:     v.Slope,
			Intercept: v.Intercept,
			Rsq:       v.Rsq,
		}
		linests = append(linests, linest)
	}

	page := Page{
		"id":      id,
		"linests": linests,
	}

	page.render("linest-fig", w)
}
