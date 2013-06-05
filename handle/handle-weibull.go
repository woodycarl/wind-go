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
	W     []float64
	V     []float64
	Cats  []int
	K     float64
	C     float64
}

func handleWeibull(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	db := wind.DB(data.RData)

	var weibulls []Weibull

	for _, v := range s.Sensors["wv"] {
		ch := v.Channel
		wv := db.Get("ChAvg" + ch)["ChAvg"+ch]

		max := wind.ArrayMax(wv) + 1.0

		k, c := wind.WeibullKC(wv)
		Info(k, c)
		wvf := wind.WvAvg(wv)
		var dataW []float64
		var dataV []float64
		var cats []int

		for i := 1; float64(i) < max; i++ {
			index := strconv.Itoa(i)
			dataW = append(dataW, 100.0*wind.Weibull(k, c, float64(i)-0.5))

			if _, ok := wvf[index]; ok {
				dataV = append(dataV, wvf[index])
			} else {
				dataV = append(dataV, 0.0)
			}

			cats = append(cats, i)
		}

		weibull := Weibull{
			Title: "Ch" + ch + " " + fmt.Sprint(v.Height) + "m高度风速威布尔分布图",
			W:     dataW,
			V:     dataV,
			K:     k,
			C:     c,
			Cats:  cats,
		}

		weibulls = append(weibulls, weibull)
	}

	page := Page{
		"id":       id,
		"weibulls": weibulls,
	}

	page.render("weibull", w)
}
