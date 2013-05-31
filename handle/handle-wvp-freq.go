package handle

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/woodycarl/wind-go/wind"
)

type WvpFreqData struct {
	Title string
	Vf    []float64
	Pf    []float64
	Cats  []string
}

func handleWvpFreq(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	db := wind.DB(data.RData)

	var wvpFreDatas []WvpFreqData
	for _, v := range s.Sensors["wv"] {
		channel := v.Channel
		height := v.Height
		v := db.Get("ChAvg" + channel)["ChAvg"+channel]
		p := wind.Wv2Wp(v, s.AirDensity)

		arg := wind.WvpArg{
			Interval: 1,
		}
		ay, vf, pf := wind.WvpWindrose(v, p, arg)

		var cats []string
		cats = append(cats, "<0.5")
		for i := 1; i < len(ay); i++ {
			cats = append(cats, strconv.Itoa(i))
		}

		wvpFreData := WvpFreqData{
			Title: "(Ch" + channel + ")" + strconv.Itoa(height) + "m高度风速风功率密度分布直方图",
			Vf:    vf,
			Pf:    pf,
			Cats:  cats,
		}
		wvpFreDatas = append(wvpFreDatas, wvpFreData)

	}

	page := Page{
		"id":   id,
		"wvps": wvpFreDatas,
	}

	page.render("wvp-freq", w)
}
