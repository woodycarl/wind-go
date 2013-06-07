package handle

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	. "github.com/woodycarl/wind-go/logger"
	"github.com/woodycarl/wind-go/wind"
)

type Weibull struct {
	Title string
	W     [][]float64
	V     [][]float64
	K     string
	C     string
}

func handleWeibull(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.S

	db := wind.DB(data.RD)

	var weibulls []Weibull

	for _, v := range s.Sensors["wv"] {
		ch := v.Channel
		wv := db.Get("ChAvg" + ch)["ChAvg"+ch]

		max := wind.ArrayMax(wv) + 1.0

		k, c := wind.WeibullKC(wv)
		Info(k, c)
		wvf := wind.WvAvg(wv)
		var dataW [][]float64
		var dataV [][]float64

		for i := 0.0; i < max; {
			tw := []float64{i, 100.0 * wind.Weibull(k, c, i)}
			dataW = append(dataW, tw)
			i = i + 0.1
		}

		for i := 1; float64(i) < max; i++ {
			index := strconv.Itoa(i)

			if _, ok := wvf[index]; ok {
				dataV = append(dataV, []float64{float64(i) - 0.5, wvf[index]})
			} else {
				dataV = append(dataV, []float64{float64(i) - 0.5, 0.0})
			}

		}

		weibull := Weibull{
			Title: "Ch" + ch + " " + fmt.Sprint(v.Height) + "m高度风速威布尔分布图",
			W:     dataW,
			V:     dataV,
			K:     fmt.Sprintf("%0.2f", k),
			C:     fmt.Sprintf("%0.2f", c),
		}

		weibulls = append(weibulls, weibull)
	}

	page := Page{
		"id":       id,
		"weibulls": weibulls,
	}

	page.render("weibull", w)
}
