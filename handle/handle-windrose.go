package handle

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/woodycarl/wind-go/wind"
)

type WindRoseData struct {
	Dir  string
	Data []float64
	Sum  float64
}
type WindRose struct {
	Title   string
	Channel string
	Height  int
	Cats    []string
	Data    []WindRoseData
	Sums    []float64
}

func handleWdvpWindRose(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data, err := getData(id)
	if err != nil {
		handleErr(w, err)
	}
	s := data.S

	catsD := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

	var windrose []WindRose

	wdvpArg := wind.WdvpArg{
		NAngles:   16,
		IntervalV: 3,
	}

	windroseV, windroseP := formatWindRoseData(wind.WdvpWindRose(s.DataWd, s.DataWv, s.DataWp, wdvpArg), catsD)
	windroseV.Title = "代表年的全年风向频率分布玫瑰图"
	windroseP.Title = "代表年的全年风能频率分布玫瑰图"
	windrose = append(windrose, windroseV, windroseP)

	var wrdatas []wind.Data
	for i, _ := range s.DataTime {
		t := time.Unix(int64(s.DataTime[i]), 0)
		d := wind.Data{
			"Month": float64(int(t.Month())),
			"Wd":    s.DataWd[i],
			"Wv":    s.DataWv[i],
			"Wp":    s.DataWp[i],
		}
		wrdatas = append(wrdatas, d)
	}

	db := wind.DB(wrdatas)

	getWR := func(dbM wind.DB, i int) {
		ds := dbM.Get("Wd Wv Wp")
		wd, wv, wp := ds["Wd"], ds["Wv"], ds["Wp"]

		windroseV, windroseP := formatWindRoseData(wind.WdvpWindRose(wd, wv, wp, wdvpArg), catsD)
		windroseV.Title = strconv.Itoa(i) + "月风向频率分布玫瑰图"
		windroseP.Title = strconv.Itoa(i) + "月风能频率分布玫瑰图"
		windrose = append(windrose, windroseV, windroseP)
	}

	if len(s.Cm) == 12 {
		for i := 1; i < 13; i++ {
			dbM := db.Filter("Month", float64(i))
			getWR(dbM, i)
		}
	} else {
		for _, v := range s.Cm {
			dbM := db.Filter("Month", v.Month)
			getWR(dbM, int(v.Month))
		}
	}

	page := Page{
		"id":   id,
		"wvps": windrose,
		"s":    s,
	}

	page.render("wvp-windrose", w)
}

func formatWindRoseData(wdvpFs wind.WdvpFs, catsD []string) (WindRose, WindRose) {
	var dataVs, dataPs []WindRoseData

	for i, v := range catsD {
		wdvpData := wdvpFs.Wdvpfs[i]

		windRoseDataV := WindRoseData{
			Dir:  v,
			Data: wdvpData.V,
			Sum:  wdvpData.Vf,
		}
		dataVs = append(dataVs, windRoseDataV)
		windRoseDataP := WindRoseData{
			Dir:  v,
			Data: wdvpData.P,
			Sum:  wdvpData.Pf,
		}
		dataPs = append(dataPs, windRoseDataP)
	}

	var catsV []string
	catsV = append(catsV, "<"+fmt.Sprint(wdvpFs.AyV[0])+" m/s")
	for i := 0; i < len(wdvpFs.AyV)-2; i++ {
		catsV = append(catsV, fmt.Sprint(wdvpFs.AyV[i])+"-"+fmt.Sprint(wdvpFs.AyV[i+1])+" m/s")
	}
	catsV = append(catsV, ">"+fmt.Sprint(wdvpFs.AyV[len(wdvpFs.AyV)-2])+" m/s")

	var catsP []string
	catsP = append(catsP, "<"+fmt.Sprint(wdvpFs.AyP[0])+" W/m2")
	for i := 0; i < len(wdvpFs.AyP)-2; i++ {
		catsP = append(catsP, fmt.Sprint(wdvpFs.AyP[i])+"-"+fmt.Sprint(wdvpFs.AyP[i+1])+" W/m2")
	}
	catsP = append(catsP, ">"+fmt.Sprint(wdvpFs.AyP[len(wdvpFs.AyP)-2])+" W/m2")

	windRoseV := WindRose{
		Cats: catsV,
		Data: dataVs,
		Sums: wdvpFs.SumV,
	}
	windRoseP := WindRose{
		Cats: catsP,
		Data: dataPs,
		Sums: wdvpFs.SumP,
	}
	return windRoseV, windRoseP
}
