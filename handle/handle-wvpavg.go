package handle

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	. "github.com/woodycarl/wind-go/logger"
	"github.com/woodycarl/wind-go/wind"
)

type WvpFigData struct {
	Name    string
	Data    []float64
	Avg     float64
	Channel string
	Height  float64
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

func handleWvpFig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cat1 := mux.Vars(r)["cat1"]
	cat2 := mux.Vars(r)["cat2"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.S

	var cats []string

	switch cat2 {
	case "ym":
		cats = []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
	case "yh":
		cats = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	}

	chds := []WvpFigData{}

	for _, v := range s.Sensors["wv"] {
		channel := v.Channel
		height := v.Height

		ch := "ChAvg" + channel
		db := wind.DB(data.RD).Get("Time " + ch)
		t := db["Time"]
		d := db[ch]

		if cat1 == "wp" {
			d = wind.Wv2Wp(d, s.AirDensity)
		}

		var ds []float64
		if cat2 == "ym" {
			ds, _, err = wind.CalWvpAvgM(t, d, true)
		} else if cat2 == "yh" {
			ds, err = wind.CalWvpAvgH(t, d)
		}
		if err != nil {
			handleErr(w, err)
			return
		}

		wvpdata := WvpFigData{
			Name:    channel + " (" + fmt.Sprint(height) + "m)",
			Data:    ds,
			Avg:     wind.ArrayAvg(ds),
			Channel: channel,
			Height:  height,
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

	page := Page{
		"id":   id,
		"wvps": wvps,
	}

	page.render("wvp", w)
}

type WvpFigMH struct {
	Title  string
	Cats   []string
	WvData []float64
	WpData []float64
}

func handleWvpMhFig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
		return
	}
	s := data.S

	cats := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}

	var wvpss []WvpFigMH

	for _, v := range s.Sensors["wv"] {
		channel, height := v.Channel, v.Height

		ch := "ChAvg" + channel
		db := wind.DB(data.RD).Get("Time " + ch)
		T := db["Time"]
		V := db[ch]
		P := wind.Wv2Wp(V, s.AirDensity)

		vs, _, err := wind.CalWvpAvgMH(T, V, true)
		if err != nil {
			handleErr(w, err)
			return
		}

		ps, _, err := wind.CalWvpAvgMH(T, P, true)
		if err != nil {
			handleErr(w, err)
			return
		}

		for i := 1; i < 13; i++ {
			wvps := WvpFigMH{
				Title:  fmt.Sprint("Ch", channel, "(", height, "m)", i, "月份风速风能日变化图"),
				Cats:   cats,
				WvData: vs[i-1],
				WpData: ps[i-1],
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

type ChWvpData struct {
	data  WvpFig
	index int
	err   error
}
type ChWvpMHData struct {
	data []WvpFigMH
	err  error
}

func handleWvpAvg(w http.ResponseWriter, r *http.Request) {
	Info("=== Handle Wvp Avg ===")
	timeS := time.Now()
	id := mux.Vars(r)["id"]
	cat := mux.Vars(r)["cat"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
		return
	}
	s := data.S

	chWvp := make(chan ChWvpData, 4)
	chWvpMH := make(chan ChWvpMHData)

	if cat == "all" {
		go getWvpAvg(data, "wv", "ym", 0, true, chWvp)
		go getWvpAvg(data, "wp", "ym", 1, true, chWvp)
		go getWvpAvg(data, "wv", "yh", 2, true, chWvp)
		go getWvpAvg(data, "wp", "yh", 3, true, chWvp)
		go getWvpAvgMh(data, true, chWvpMH)
	} else if cat == "turb" {
		go getTurbineWvpAvg(s, "wv", "ym", 0, true, chWvp)
		go getTurbineWvpAvg(s, "wp", "ym", 1, true, chWvp)
		go getTurbineWvpAvg(s, "wv", "yh", 2, true, chWvp)
		go getTurbineWvpAvg(s, "wp", "yh", 3, true, chWvp)
		go getTurbineWvpAvgMh(s, true, chWvpMH)
	}

	d := map[int]WvpFig{}
	for i := 0; i < 4; i++ {
		td := <-chWvp
		if td.err != nil {
			handleErr(w, err)
			return
		}
		d[td.index] = td.data
	}

	var wvps []WvpFig
	for i := 0; i < 4; i++ {
		wvps = append(wvps, d[i])
	}

	td := <-chWvpMH
	if td.err != nil {
		handleErr(w, err)
		return
	}
	wvpsmh := td.data

	page := Page{
		"id":     id,
		"wvps":   wvps,
		"wvpsmh": wvpsmh,
	}

	page.render("wvp-avg", w)
	Info("=== End Handle Wvp Avg", time.Now().Sub(timeS), "===")
}

func getWvpAvg(data Data, cat1, cat2 string, index int, limit bool, ch chan ChWvpData) {
	timeS := time.Now()
	s := data.S

	var chWvpData ChWvpData
	var cats, A []string

	chds := []WvpFigData{}

	for _, v := range s.Sensors["wv"] {
		channel := v.Channel
		height := v.Height

		db := wind.DB(data.RD).Get("Time ChAvg" + channel)
		t := db["Time"]
		d := db["ChAvg"+channel]

		if cat1 == "wp" {
			d = wind.Wv2Wp(d, s.AirDensity)
		}

		var ds []float64
		if cat2 == "ym" {
			ds, A, chWvpData.err = wind.CalWvpAvgM(t, d, limit)
			if !limit {
				cats = A
			} else {
				cats = []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
			}
		} else if cat2 == "yh" {
			ds, chWvpData.err = wind.CalWvpAvgH(t, d)
			cats = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
		}
		if chWvpData.err != nil {
			ch <- chWvpData
			return
		}

		wvpdata := WvpFigData{
			Name:    channel + " (" + fmt.Sprint(height) + "m)",
			Data:    ds,
			Avg:     wind.ArrayAvg(ds),
			Channel: channel,
			Height:  height,
		}

		chds = append(chds, wvpdata)
	}

	title := "不同高度测风年"

	chWvpData.data = getWvpAvgData(cats, chds, cat1, cat2, title)
	chWvpData.index = index
	ch <- chWvpData
	Info("getWvpAvg", time.Now().Sub(timeS))
}

func getWvpAvgData(cats []string, chds []WvpFigData, cat1, cat2, title string) WvpFig {
	wvps := WvpFig{
		Cats: cats,
		Data: chds,
	}

	switch cat1 + cat2 {
	case "wvym":
		wvps.Title = title + "逐月平均风速"
	case "wvyh":
		wvps.Title = title + "逐时平均风速"
	case "wpym":
		wvps.Title = title + "逐月平均风功率"
	case "wpyh":
		wvps.Title = title + "逐时平均风功率"
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

	return wvps
}

func getWvpAvgMh(data Data, limit bool, ch chan ChWvpMHData) {
	timeS := time.Now()
	var chWvpMHData ChWvpMHData
	s := data.S

	var wvpss []WvpFigMH

	for _, v := range s.Sensors["wv"] {
		channel, height := v.Channel, v.Height

		db := wind.DB(data.RD).Get("Time ChAvg" + channel)
		T := db["Time"]
		V := db["ChAvg"+channel]
		P := wind.Wv2Wp(V, s.AirDensity)

		title := fmt.Sprint("Ch", channel, "(", height, "m)")

		wvps, err := getWvpAvgM(T, V, P, limit, title)
		if err != nil {
			chWvpMHData.err = err
			ch <- chWvpMHData
			return
		}

		wvpss = append(wvpss, wvps...)
	}

	chWvpMHData.data = wvpss
	ch <- chWvpMHData
	Info("getWvpAvgMh", time.Now().Sub(timeS))
}

func getWvpAvgM(T, V, P []float64, limit bool, title string) (wvpss []WvpFigMH, err error) {
	var vs, ps [][]float64
	var cats, A []string

	vs, A, err = wind.CalWvpAvgMH(T, V, limit)
	if err != nil {
		return
	}
	if !limit {
		cats = A
	} else {
		cats = []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
	}

	ps, _, err = wind.CalWvpAvgMH(T, P, limit)
	if err != nil {
		return
	}

	for i, v := range cats {
		wvps := WvpFigMH{
			Title:  title + " " + v + "月份风速风能日变化图",
			Cats:   cats,
			WvData: vs[i],
			WpData: ps[i],
		}
		wvpss = append(wvpss, wvps)
	}

	return
}

func getTurbineWvpAvg(s wind.Station, cat1, cat2 string, index int, limit bool, ch chan ChWvpData) {
	timeS := time.Now()

	var chWvpData ChWvpData

	var D []float64
	if cat1 == "wp" {
		D = s.DataWp
	} else if cat1 == "wv" {
		D = s.DataWv
	}

	var cats []string
	var ds []float64
	var A []string
	switch cat2 {
	case "ym":
		ds, A, chWvpData.err = wind.CalWvpAvgM(s.DataTime, D, limit)
		if !limit {
			cats = A
		} else {
			cats = []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		}
	case "yh":
		ds, chWvpData.err = wind.CalWvpAvgH(s.DataTime, D)
		cats = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	}
	if chWvpData.err != nil {
		ch <- chWvpData
		return
	}

	wvpdata := WvpFigData{
		Name: "数据",
		Data: ds,
		Avg:  wind.ArrayAvg(ds),
	}

	chds := []WvpFigData{wvpdata}

	title := "轮毂高度(" + fmt.Sprint(config.Config.CalHeight) + "m)测风年"

	chWvpData.data = getWvpAvgData(cats, chds, cat1, cat2, title)
	chWvpData.index = index
	ch <- chWvpData
	Info("getTurbineWvpAvg", time.Now().Sub(timeS))
}

func getTurbineWvpAvgMh(s wind.Station, limit bool, ch chan ChWvpMHData) {
	timeS := time.Now()
	var chWvpMHData ChWvpMHData
	var wvpss []WvpFigMH

	T := s.DataTime
	V := s.DataWv
	P := s.DataWp

	title := fmt.Sprint("轮毂高度(", s.TurbineHeight, "m)")

	wvpss, err := getWvpAvgM(T, V, P, limit, title)
	if err != nil {
		chWvpMHData.err = err
		ch <- chWvpMHData
		return
	}

	chWvpMHData.data = wvpss
	ch <- chWvpMHData
	Info("getTurbineWvpAvgMh", time.Now().Sub(timeS))
}
