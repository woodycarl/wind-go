package handle

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func handleWvpFig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cat1 := mux.Vars(r)["cat1"]
	cat2 := mux.Vars(r)["cat2"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	var cats []string
	var startIndex, endIndex int

	switch cat2 {
	case "ym":
		cats = []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		startIndex, endIndex = 1, 13
	case "yh":
		cats = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
		startIndex, endIndex = 0, 24
	}

	type WvpFigData struct {
		Name    string
		Data    []float64
		Avg     float64
		Channel string
		Height  int
	}
	type WvpFig struct {
		Title       string
		Subtitle    string
		Cats        []string
		YaxisTitle  string
		ValueSuffix string
		Data        []WvpFigData
		Unit        string
	}

	chds := []WvpFigData{}

	for i, v := range s.Wvp {
		channel := s.Sensors["wv"][i].Channel
		height := s.Sensors["wv"][i].Height
		var ds []float64

		for j := startIndex; j < endIndex; j++ {
			js := strconv.Itoa(j)
			var d float64
			switch cat1 + cat2 {
			case "wvym":
				d = v[js].Wv
			case "wvyh":
				d = v["0"].Hwv[js]
			case "wpym":
				d = v[js].Wp
			case "wpyh":
				d = v["0"].Hwp[js]
			}
			ds = append(ds, d)
		}

		wvpdata := WvpFigData{
			Name:    channel + " (" + strconv.Itoa(height) + "m)",
			Data:    ds,
			Channel: channel,
			Height:  height,
		}
		switch cat1 {
		case "wv":
			wvpdata.Avg = v["0"].Wv
		case "wp":
			wvpdata.Avg = v["0"].Wp
		}
		chds = append(chds, wvpdata)
	}

	wvps := WvpFig{
		Title:    "不同高度测风年逐时平均风速",
		Subtitle: "",
		Cats:     cats,
		Data:     chds,
	}

	switch cat1 {
	case "wv":
		wvps.YaxisTitle = "风速 (m/s)"
		wvps.ValueSuffix = "m/s"
	case "wp":
		wvps.YaxisTitle = "风功率 (W/m2)"
		wvps.ValueSuffix = "W/m2"
	}

	switch cat2 {
	case "ym":
		wvps.Unit = "月"
	case "yh":
		wvps.Unit = "小时"
	}

	wvpss := []WvpFig{wvps}

	page := Page{
		"id":   id,
		"wvps": wvpss,
	}

	page.render("wvp", w)
}

func handleWvpMhFig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.Station

	cats := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}

	type WvpFig struct {
		Title  string
		Cats   []string
		WvData []float64
		WpData []float64
	}
	var wvpss []WvpFig

	for i, v := range s.Sensors["wv"] {
		channel, height := v.Channel, v.Height
		mwvps := s.Wvp[i]

		for i := 1; i < 13; i++ {
			is := strconv.Itoa(i)
			mwvp := mwvps[is]

			var dsv []float64
			var dsp []float64
			for j := 0; j < 24; j++ {
				js := strconv.Itoa(j)

				dsv = append(dsv, mwvp.Hwv[js])
				dsp = append(dsp, mwvp.Hwp[js])
			}

			wvps := WvpFig{
				Title:  "Ch" + channel + "(" + strconv.Itoa(height) + "m)" + is + "月份风速风能日变化图",
				Cats:   cats,
				WvData: dsv,
				WpData: dsp,
			}
			wvpss = append(wvpss, wvps)
		}

	}

	page := Page{
		"id":   id,
		"wvps": wvpss,
	}

	page.render("wvp-mh", w)
}

func handleTurbineWvp(w http.ResponseWriter, r *http.Request) {
	/*
		id := mux.Vars(r)["id"]

		data := getData(id)
		s := data.Station
	*/
}
