package wind

import (
	"math"
	"strconv"
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

func caculates(r []Result, c Config) []Result {
	Info("---Caculate---")
	for i, v := range r {
		r[i].S = caculate(v.S, v.D1, v.D2, v.RD, c)
	}

	return r
}

func caculate(s Station, d1, d2, rd []Data, c Config) Station {
	db := DB(rd)

	// 空气密度

	if v, ok := s.Sensors["t"]; ok {
		chT := "ChAvg" + v[0].Channel
		dataT := db.get(chT)[chT]

		if len(s.Sensors["p"]) > 0 {
			chP := "ChAvg" + s.Sensors["p"][0].Channel
			dataP := db.get(chP)[chP]
			s.AirDensity = AirDensity(dataP, dataT)
		} else {
			s.AirDensity = AirDensity2(dataT, c.CalHeight)
		}
	}

	Info("AirDensity:", s.AirDensity)

	chTurbs := make(chan []float64)
	go calTurbs(s, d1, d2, chTurbs)

	chWvp := make(chan []map[string]Mwvp)
	go calAvgWvp(s, db, chWvp)

	chWss := make(chan Wss)
	go calWindShear(db, s.Sensors["wv"], chWss)

	// 依照设定高度选取channel数据 或计算得到
	bv, wvCh := chooseCh(s.Sensors["wv"], c.CalHeight)
	_, wdCh := chooseCh(s.Sensors["wd"], c.CalHeight)
	s.DataWd = db.get("ChAvg" + wdCh)["ChAvg"+wdCh]

	if bv {
		s.DataWv = db.get("ChAvg" + wvCh)["ChAvg"+wvCh]
	} else {
		s.DataWv = calWvData(rd, s.Sensors["wv"], c.CalHeight)
	}

	for _, v := range s.DataWv {
		s.DataWp = append(s.DataWp, s.AirDensity*math.Pow(v, 3.0)/2.0)
	}

	s.DataTime = db.get("Time")["Time"]

	s.Wss = <-chWss

	s.Wvp = <-chWvp
	// 计算平均速度和风功率
	s.Turbs = <-chTurbs

	Info("done cal")
	return s
}

// 选取设定高度的channel wv wd
func chooseCh(s []Sensor, calHeight float64) (b bool, c string) {
	dhMin, c := 100.0, ""
	for _, v := range s {
		height := float64(v.Height)
		if height == calHeight {
			return true, v.Channel
		}
		dh := math.Abs(height - calHeight)
		if dh < dhMin {
			dhMin = dh
			c = v.Channel
		}
	}
	return false, c
}

// 计算平均风速风功率
func calAvgWvp(s Station, db DB, chWvp chan []map[string]Mwvp) {
	Info("Wvp")
	dbMs := map[string]DB{}
	for _, v := range s.Cm {
		m := strconv.Itoa(int(v.Month))
		dbMs[m] = db.filter("Month", v.Month)
	}

	wvp := []map[string]Mwvp{}
	for _, v1 := range s.Sensors["wv"] {
		ch := "ChAvg" + v1.Channel

		chwvp := map[string]Mwvp{}
		chwvp0 := Mwvp{Hwv: map[string]float64{}, Hwp: map[string]float64{}}
		for _, v2 := range s.Cm {
			m := strconv.Itoa(int(v2.Month))

			chwvpM := Mwvp{Hwv: map[string]float64{}, Hwp: map[string]float64{}}

			dataM := dbMs[m].get(ch)[ch]
			chwvpM.Wv = ArrayAvg(dataM)
			chwvpM.Wp = s.AirDensity * ArrayAvg(ArrayPow(dataM, 3.0)) / 2.0

			chwvp0.Wv = chwvp0.Wv + chwvpM.Wv
			chwvp0.Wp = chwvp0.Wp + chwvpM.Wp

			for k := 0; k < 24; k++ {
				ks := strconv.Itoa(k)
				dataMH := dbMs[m].filter("Hour", float64(k)).get(ch)[ch]

				chwvpM.Hwv[ks] = ArrayAvg(dataMH)
				chwvpM.Hwp[ks] = s.AirDensity * ArrayAvg(ArrayPow(dataMH, 3.0)) / 2.0

				chwvp0.Hwv[ks] = chwvp0.Hwv[ks] + chwvpM.Hwv[ks]
				chwvp0.Hwp[ks] = chwvp0.Hwp[ks] + chwvpM.Hwp[ks]

			}

			chwvp[m] = chwvpM
		}
		chwvp0.Wv = chwvp0.Wv / 12.0
		chwvp0.Wp = chwvp0.Wp / 12.0
		for k := 0; k < 24; k++ {
			h := strconv.Itoa(k)
			chwvp0.Hwv[h] = chwvp0.Hwv[h] / 12.0
			chwvp0.Hwp[h] = chwvp0.Hwp[h] / 12.0
		}
		chwvp["0"] = chwvp0
		wvp = append(wvp, chwvp)
	}
	chWvp <- wvp
	return
}

func CalAvgWvp(T, V, P []float64) (Vym, Pym, Vyh, Pyh []float64, Vmh, Pmh [][]float64) {
	var data []Data
	for i, v := range T {
		t := time.Unix(int64(v), 0)
		m := t.Month()
		h := t.Hour()
		d := Data{
			"Month": float64(int(m)),
			"Hour":  float64(h),
			"V":     V[i],
			"P":     P[i],
		}
		data = append(data, d)
	}

	db := DB(data)

	for i := 1; i < 13; i++ {
		dbM := db.filter("Month", float64(i))
		dataM := dbM.get("V P")
		Vym = append(Vym, ArrayAvg(dataM["V"]))
		Pym = append(Pym, ArrayAvg(dataM["P"]))

		var vmh, pmh []float64
		for k := 0; k < 24; k++ {
			dataMH := dbM.filter("Hour", float64(k)).get("V P")

			vmh = append(vmh, ArrayAvg(dataMH["V"]))
			pmh = append(pmh, ArrayAvg(dataMH["P"]))
		}

		Vmh = append(Vmh, vmh)
		Pmh = append(Pmh, pmh)
	}

	for i := 0; i < 24; i++ {
		dbH := db.filter("Hour", float64(i))
		dataH := dbH.get("V P")

		Vyh = append(Vyh, ArrayAvg(dataH["V"]))
		Pyh = append(Pyh, ArrayAvg(dataH["P"]))
	}

	return
}

// 湍流强度 数据源 data1h data10m
func calTurbs(s Station, data1h, data10m []Data, chTurbs chan []float64) {
	Info("Turbs")
	cm := s.Cm
	location, _ := time.LoadLocation("Local")
	timeS := float64(time.Date(int(cm[0].Year), time.Month(int(cm[0].Month)), 1, 0, 0, 0, 0, location).Unix())
	timeE := float64(time.Date(int(cm[11].Year), time.Month(int(cm[11].Month)+1), 1, 0, 0, 0, 0, location).Unix())
	db1 := DB(data1h).filter("Time >=", timeS).filter("Time <", timeE)
	db10 := DB(data10m).filter("Time >=", timeS).filter("Time <", timeE)

	ch := make(chan ChTurb, len(s.Sensors["wv"]))
	for _, v1 := range s.Sensors["wv"] {
		go calTurb(v1, db1, db10, ch)
	}

	var turbs []float64
	var tData []ChTurb
	for _, _ = range s.Sensors["wv"] {
		tData = append(tData, <-ch)
	}
	for _, v1 := range s.Sensors["wv"] {
		channel := v1.Channel
		for _, v2 := range tData {
			if v2.channel == channel {
				turbs = append(turbs, v2.turbulence)
			}
		}
	}

	chTurbs <- turbs
	return
}

type ChTurb struct {
	channel    string
	turbulence float64
}

func calTurb(v1 Sensor, db1, db10 DB, ch chan ChTurb) {
	location, _ := time.LoadLocation("Local")
	channel := v1.Channel
	data1 := db1.filter("ChAvg"+channel+" >=", 14.5).filter("ChAvg"+channel+" <=", 15.5)

	if len(data1) < 1 {
		Error("calTurb: not enough data!")
		// 需要增加错误返回
	}

	turb := []float64{}
	for _, v2 := range data1 {
		if len(db10) < 1 {
			turbulence := v2["ChSd"+channel] / v2["ChAvg"+channel]
			turb = append(turb, turbulence)
			continue
		}

		tS := time.Unix(int64(v2["Time"]), 0)
		tE := float64(time.Date(tS.Year(), tS.Month(), tS.Day(), tS.Hour()+1, 0, 0, 0, location).Unix())
		data10 := db10.filter("Time >=", float64(tS.Unix())).filter("Time <", tE)

		if len(data10) < 1 {
			continue
		}

		turbulence := data10[0]["ChSd"+channel] / data10[0]["ChAvg"+channel]
		for _, v3 := range data10 {
			t := v3["ChSd"+channel] / v3["ChAvg"+channel]
			if t > turbulence {
				turbulence = t
			}
		}
		turb = append(turb, turbulence)
	}

	chTurb := ChTurb{
		channel:    channel,
		turbulence: ArrayAvg(turb),
	}

	ch <- chTurb
}

// 风切变指数
func calWindShear(db DB, s []Sensor, chWss chan Wss) {
	Info("Wss")
	//按高度分类
	type DD struct {
		Channel string
		Avg     float64
		Index   int
	}
	type DH struct {
		Height float64
		Data   []DD
	}

	dH := []DH{}
	dataH := []float64{}
	dataWv := []float64{}

	existDH := func(dh []DH, height float64) (bool, int) {
		for i, v := range dh {
			if v.Height == height {
				return true, i
			}
		}
		return false, 0
	}
	for i, v := range s {
		ch := v.Channel
		height := v.Height
		avg := ArrayAvg(db.get("ChAvg" + ch)["ChAvg"+ch])

		dataH = append(dataH, math.Log(float64(height)))
		dataWv = append(dataWv, math.Log(avg))
		dd := DD{
			Channel: v.Channel,
			Avg:     avg,
			Index:   i,
		}
		b, j := existDH(dH, height)
		if !b {

			dh := DH{
				Height: height,
			}
			dh.Data = append(dh.Data, dd)
			dH = append(dH, dh)
			continue
		}
		dH[j].Data = append(dH[j].Data, dd)
	}

	wss := []Ws{}
	for i := 0; i < len(dH)-1; i++ {
		data1 := dH[i].Data
		data2 := dH[i+1].Data
		height1 := dH[i].Height
		height2 := dH[i+1].Height

		for _, v1 := range data1 {
			for _, v2 := range data2 {
				// if height1==height2 {continue}
				rws := math.Log(v1.Avg/v2.Avg) / math.Log(float64(height1)/float64(height2))

				ws := Ws{
					YI:  v1.Index,
					XI:  v2.Index,
					YCh: v1.Channel,
					XCh: v2.Channel,
					YH:  height1,
					XH:  height2,
					Ws:  rws,
				}
				wss = append(wss, ws)
			}
		}
	}

	slope, intercept, rsq := CalLinestRsq(dataH, dataWv)

	a := math.Exp(intercept)
	b := slope

	wsss := Wss{
		Ws:     wss,
		A:      a,
		B:      b,
		R:      rsq,
		Height: dataH,
		Wv:     dataWv,
	}

	chWss <- wsss
	return
}

func calWvData(rData []Data, s []Sensor, h float64) (data []float64) {
	Info("new wv data: height", h)
	for _, v1 := range rData {
		var wv []float64
		var height []float64
		for _, v2 := range s {
			wv = append(wv, math.Log(v1["ChAvg"+v2.Channel]))
			height = append(height, math.Log(float64(v2.Height)))
		}

		slope, intercept, _ := CalLinestRsq(height, wv)
		a := math.Exp(intercept)
		b := slope
		data = append(data, a*math.Pow(h, b))
	}

	return data
}
