package wind

import (
	"errors"
	"math"
	"math/rand"
	"strconv"
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

type ErrRTC struct {
	id  string
	cat string
	err [][]bool
}

func getErrRTC(errs []ErrRTC, id, cat string, index int) (e []bool, err error) {
	for _, v := range errs {
		if v.id == id && v.cat == cat {
			if index > len(v.err)-1 {
				err = errors.New("getErrRTC: out of range!")
				return
			}
			e = v.err[index]
			return
		}
	}

	if e == nil {
		err = errors.New("getErrRTC: get none!")
		return
	}

	return
}
func calErrNumC1(errs [][][]bool) (errt [][]bool) {
	for _, v := range errs {
		errt = append(errt, calErrNumRT(v))
	}

	return
}

func revises(r []Result, c Config) (rr []Result, err error) {
	Info("---Revise---")

	// 1.Lost
	for i, v := range r {
		r[i].RD, err = rLost(v.S, v.D1, c)
		if err != nil {
			return
		}
	}

	// 2.合理性修订
	var errsR []ErrRTC
	catsR := []string{"wv", "wd"}
	for _, v := range r {
		db := DB(v.RD)
		for _, v1 := range catsR {
			errR := ErrRTC{
				id:  v.ID,
				cat: v1,
				err: getErrR(db, v1, v.S.Sensors[v1]),
			}

			errsR = append(errsR, errR)
		}
	}

	for i, v := range r {
		for _, cat := range catsR {
			for iSensor, sensor := range v.S.Sensors[cat] {
				r[i].RD, err = rRationality(r, i, cat, iSensor, sensor, errsR, v.RD)
				if err != nil {
					return
				}
			}
		}
	}

	// 3.趋势性修订
	var errsT []ErrRTC
	catsT := []string{"wv", "t", "p"}
	for _, v := range r {
		db := DB(v.RD)
		for _, cat := range catsT {
			errT := ErrRTC{
				id:  v.ID,
				cat: cat,
				err: getErrT(db, cat, v.S.Sensors[cat]),
			}

			errsT = append(errsT, errT)
		}
	}

	for i, v := range r {
		for _, cat := range catsT {
			for iSensor, sensor := range v.S.Sensors[cat] {
				r[i].RD, err = rTrends(r, i, cat, iSensor, sensor, errsT, v.RD)
				if err != nil {
					return
				}
			}
		}
	}

	// 4.相关性修订
	var errsC []ErrRTC
	catsC := []string{"wv"}
	for _, v := range r {
		db := DB(v.RD)
		for _, cat := range catsC {

			errC := ErrRTC{
				id:  v.ID,
				cat: cat,
				err: calErrNumC1(getErrC(db, cat, v.S.Sensors[cat])),
			}

			errsC = append(errsC, errC)
		}
	}

	for i, v := range r {
		for _, cat := range catsC {
			for iSensor, sensor := range v.S.Sensors[cat] {
				r[i].RD, err = rCorrelation(r, i, cat, iSensor, sensor, errsC, v.RD)
				if err != nil {
					return
				}
			}
		}
	}

	rr = r
	return
}

func rLost(s Station, data []Data, c Config) (rData []Data, err error) {
	// 需要增加const
	if len(data) < 5000 {
		err = errors.New("rLost: not enough data!")
		return
	}

	db := DB(data)

	monthData, hourData := getMonthHourData(s, db)

	/* 按给定月份小时数据平均值来给定一个新值 */
	avgMonthHour := func(ch, month, hour string) float64 {
		md := monthData[month][ch]
		hd := hourData[month][hour][ch]

		m := ArrayAvg(md)
		h := ArrayAvg(hd)

		return m*0.3 + h*0.7
	}
	/* 在给定月份小时中随机取一个值作为新值 */
	randomData := func(ch, month, hour string) float64 {
		md := monthData[month][ch]
		hd := hourData[month][hour][ch]

		m := md[int(round(rand.Float64()*float64(len(md)), 0))]
		h := hd[int(round(rand.Float64()*float64(len(hd)), 0))]

		return m*0.3 + h*0.7
	}

	/* 补充缺失数据 */
	for _, v := range s.Cm {
		dataMy := db.filter("My", v.My)

		// 依照dataMy的len需要细化

		k := 0
		for j := 0; j < v.Sbm; j++ {
			d := Data{}
			t := getNewDate(int(v.Year), int(v.Month), j)
			if len(dataMy) > k && float64(t.Unix())-dataMy[k]["Time"] >= 0.0 {
				d = dataMy[k]
				k = k + 1
			} else {
				month := strconv.Itoa(int(v.Month))
				hour := strconv.Itoa(t.Hour())

				d = Data{
					"Time":  float64(t.Unix()),
					"Hour":  float64(int(v.Month)),
					"My":    v.My,
					"Year":  v.Year,
					"Month": v.Month,
				}

				for _, v1 := range s.SensorsR {
					ch := v1.Channel

					switch c.RlostMethod {
					case "avg":
						d["ChAvg"+ch] = avgMonthHour("ChAvg"+ch, month, hour)
						d["ChSd"+ch] = avgMonthHour("ChSd"+ch, month, hour)
						if c.AllowMinMax {
							d["ChMin"+ch] = avgMonthHour("ChMin"+ch, month, hour)
							d["ChMax"+ch] = avgMonthHour("ChMax"+ch, month, hour)
						}

					case "random":
						d["ChAvg"+ch] = randomData("ChAvg"+ch, month, hour)
						d["ChSd"+ch] = randomData("ChSd"+ch, month, hour)
						if c.AllowMinMax {
							d["ChMin"+ch] = randomData("ChMin"+ch, month, hour)
							d["ChMax"+ch] = randomData("ChMax"+ch, month, hour)
						}
					}
				}

			}

			rData = append(rData, d)
		}
	}

	// >12个相同作缺失处理
	cats := []string{"wv", "wd", "t", "h"} // ,"p"
	type Same struct {
		Channel string
		Start   int
		Val     float64
		Num     int
		Cat     string
	}
	sames := []Same{}
	for _, v1 := range cats {
		for _, v2 := range s.Sensors[v1] {
			same := Same{
				Channel: v2.Channel,
				Start:   0,
				Val:     0,
				Num:     0,
				Cat:     v1,
			}
			sames = append(sames, same)
		}
	}
	isSame := func(same Same, d []Data) bool {
		/* 判断12个Sd关系以确定是否仪器故障 */
		r := 0.0
		for i := 0; i < same.Num; i++ {
			r = r + d[same.Start+i]["ChSd"+same.Channel]
		}
		return r < 0.1
	}

	for i, v1 := range rData {
		for j, v2 := range sames {
			ch := v2.Channel

			if v1["ChAvg"+ch] == v2.Val {
				sames[j].Num = v2.Num + 1
			} else {
				num := v2.Num
				start := v2.Num
				val := v2.Val
				if num > 12 && isSame(v2, rData) {
					// println("传感器"+ch+"失效 开始:"+iToS(start) + "相同数:"+iToS(num))
					// 需要增加方法，用相关性来修订
					avg := (rData[start-1]["ChAvg"+ch] + rData[start+num]["ChAvg"+ch]) / 2

					for k := 0; k < num; k++ {
						rData[start+k]["ChAvg"+ch] = avg + rand.Float64()*(val-avg)
					}
				}

				sames[j].Start = i
				sames[j].Val = rData[i]["ChAvg"+ch]
				sames[j].Num = 1
			}

		}
	}

	return
}

type Dt map[string][]float64
type DDt map[string]Dt
type DDDt map[string]DDt

func getMonthHourData(s Station, db DB) (DDt, DDDt) {
	/* 分类提取月份小时的数据 */
	monthData := DDt{}
	hourData := DDDt{}
	for _, v1 := range s.Cm {
		dbMD := db.filter("My", v1.My)

		if len(dbMD) < 100 {
			dbMD = db.filter("Month", v1.Month)
			Error(s.Site.Site, "no data, instead with Month Data", v1.My, len(dbMD))

			if len(dbMD) < 100 {
				Error(s.Site.Site, "no Month Data")
				// 需要增加err 返回，或处理方法
			}
		}

		monthDataI := Dt{}
		for _, v2 := range s.SensorsR {
			ch := v2.Channel
			monthDataI["ChAvg"+ch] = dbMD.get("ChAvg" + ch)["ChAvg"+ch]
			monthDataI["ChSd"+ch] = dbMD.get("ChSd" + ch)["ChSd"+ch]

			monthDataI["ChMin"+ch] = dbMD.get("ChMin" + ch)["ChMin"+ch]
			monthDataI["ChMax"+ch] = dbMD.get("ChMax" + ch)["ChMax"+ch]

		}
		m := strconv.Itoa(int(v1.Month))
		monthData[m] = monthDataI

		hourDataI := DDt{}
		for j := 0; j < 24; j++ {
			h := strconv.Itoa(j)
			hourDataIJ := Dt{}
			dbHD := dbMD.filter("Hour", float64(j))
			for _, v3 := range s.SensorsR {
				ch := v3.Channel
				hourDataIJ["ChAvg"+ch] = dbHD.get("ChAvg" + ch)["ChAvg"+ch]
				hourDataIJ["ChSd"+ch] = dbHD.get("ChSd" + ch)["ChSd"+ch]

				hourDataIJ["ChMin"+ch] = dbHD.get("ChMin" + ch)["ChMin"+ch]
				hourDataIJ["ChMax"+ch] = dbHD.get("ChMax" + ch)["ChMax"+ch]

			}
			hourDataI[h] = hourDataIJ
		}
		hourData[m] = hourDataI
	}

	return monthData, hourData
}

func getDataByTime(data []Data, t float64, ch string) (f float64, err error) {
	for _, v := range data {
		if v["Time"] == t {
			if d, ok := v["ChAvg"+ch]; ok {
				f = d
				return
			}
			err = errors.New("getDataByTime: no data in this channel!")
			return
		}
	}

	err = errors.New("getDataByTime: get none!")
	return
}

func getRByID(r []Result, id string) (index int, err error) {
	for i, v := range r {
		if v.ID == id {
			index = i
			return
		}
	}

	err = errors.New("getRByID: get none!")
	return
}

func rRationality(r []Result, i int, cat string, iSensor int, sensor Sensor, errsR []ErrRTC, data []Data) (rd []Data, err error) {
	chI := sensor.Channel
	var errI []bool
	errI, err = getErrRTC(errsR, r[i].ID, cat, iSensor)
	if err != nil {
		return
	}

	for j, dataJ := range data {
		if errI[j] {
			// 修订方法1，如果存在相关性数据，根据相关性来计算
			if len(sensor.Rations) > 0 {
				var ration Ration
				var d float64
				var b bool

				b, d, ration, err = rChannel(r, i, cat, sensor.Rations, j, dataJ, errsR)
				if err != nil {
					return
				}

				if !b {

					data[j]["ChAvg"+chI] = ration.Slope*d + ration.Intercept

					continue
				}
			}

			// 修订方法2，上下正常值的平均值
			if j > 0 && j < len(errI)-1 && !errI[j-1] && !errI[j+1] {
				data[j]["ChAvg"+chI] = (data[j-1]["ChAvg"+chI] + data[j+1]["ChAvg"+chI]) / 2
				continue
			}
			// 3.

			Warn("rRationality: not handle!", r[i].ID, "j")
			//err = errors.New("rRationality:" + r[i].ID + " j ")
			//return
		}
	}

	rd = data

	return
}
func rChannel(r []Result, i int, cat string, rations []Ration, j int, data Data, errsR []ErrRTC) (b bool, f float64, ration Ration, err error) {
	for _, v := range rations {
		if r[i].ID == v.ID {
			var e []bool
			e, err = getErrRTC(errsR, v.ID, cat, v.Index)
			if err != nil {
				return
			}

			if !e[j] {
				ration = v
				f = data["ChAvg"+v.Channel]
				return
			}
		}

		if r[i].ID != v.ID {
			var indexR int
			indexR, err = getRByID(r, v.ID)
			if err != nil {
				return
			}

			f, err = getDataByTime(r[indexR].D1, data["Time"], v.Channel)
			if err != nil {
				return
			}

			if !jR(f, cat) {
				ration = v
				return
			}
		}
	}

	b = true
	return
}

func rTrends(r []Result, i int, cat string, iSensor int, sensor Sensor, errsT []ErrRTC, data []Data) (rd []Data, err error) {
	chI := sensor.Channel

	for j := 0; j < len(data)-1; j++ {
		data1 := data[j]["ChAvg"+chI]
		data2 := data[j+1]["ChAvg"+chI]

		if jT(data1, data2, cat) {
			// 修订方法1，如果传感器多于1个，根据相关性来计算
			if len(sensor.Rations) > 0 {
				var ration Ration
				var d1, d2 float64
				var b bool

				b, d1, d2, ration, err = tChannel(r, i, cat, sensor.Rations, j, errsT)
				if err != nil {
					return
				}

				if !b {
					if td1 := ration.Slope*d1 + ration.Intercept; jT(td1, data2, cat) {
						data[j]["ChAvg"+chI] = td1
						continue
					} else if td2 := ration.Slope*d2 + ration.Intercept; jT(td1, td2, cat) {
						data[j]["ChAvg"+chI] = td1
						data[j+1]["ChAvg"+chI] = td2
						continue
					}
				}
			}

			// 修订方法2，为之前两个值的平均值
			if j > 1 {
				data[j]["ChAvg"+chI] = (data[j-1]["ChAvg"+chI] + data[j-2]["ChAvg"+chI]) / 2
				continue
			}

			// 需要修订方法3，
			Warn("rTrends: not handle!", r[i].ID, j)
			//err = errors.New("rTrends:" + r[i].ID + " j ")
			//return
		}
	}

	rd = data

	return
}
func tChannel(r []Result, i int, cat string, rations []Ration, j int, errsT []ErrRTC) (b bool, d1, d2 float64, ration Ration, err error) {
	for _, v := range rations {
		if r[i].ID == v.ID {

			var e []bool
			e, err = getErrRTC(errsT, v.ID, cat, v.Index)
			if err != nil {
				return
			}

			switch {
			case j == 0 && !e[j] && !e[j+1]:
				fallthrough
			case j == len(r[i].RD)-1 && !e[j-1] && !e[j]:
				fallthrough
			case j > 0 && j < len(r[i].RD)-1 && !e[j-1] && !e[j] && !e[j+1]:
				ration = v
				d1 = r[i].RD[j]["ChAvg"+v.Channel]
				d2 = r[i].RD[j+1]["ChAvg"+v.Channel]
				return
			}
		}

		if r[i].ID != v.ID {
			Info("!=")
			var indexR int
			indexR, err = getRByID(r, v.ID)
			if err != nil {
				return
			}

			d1, err = getDataByTime(r[indexR].D1, r[i].RD[j]["Time"], v.Channel)
			if err != nil {
				return
			}

			d2, err = getDataByTime(r[indexR].D1, r[i].RD[j+1]["Time"], v.Channel)
			if err != nil {
				return
			}

			if !jT(d1, d2, cat) {
				ration = v
				return
			}
		}
	}

	b = true
	return
}

func rCorrelation(r []Result, i int, cat string, iSensor int, sensor Sensor, errsC []ErrRTC, data []Data) (rd []Data, err error) {
	chI := sensor.Channel
	var errI []bool
	errI, err = getErrRTC(errsC, r[i].ID, cat, iSensor)
	if err != nil {
		return
	}

	for j, dataJ := range data {
		if errI[j] {
			// 修订方法1，如果存在相关性数据，根据相关性来计算
			if len(sensor.Rations) > 0 {
				var ration Ration
				var d float64
				var b bool

				b, d, ration, err = cChannel(r, i, cat, sensor.Rations, j, dataJ, errsC)
				if err != nil {
					return
				}

				if !b {

					data[j]["ChAvg"+chI] = ration.Slope*d + ration.Intercept

					continue
				}
			}

			// 修订方法2，上下正常值的平均值
			if j > 0 && j < len(errI)-1 && !errI[j-1] && !errI[j+1] {

				data[j]["ChAvg"+chI] = (data[j-1]["ChAvg"+chI] + data[j+1]["ChAvg"+chI]) / 2

				continue
			}
			// 3.

			Warn("rCorrelation: not handle!", r[i].ID, j)
		}
	}

	rd = data

	return
}
func cChannel(r []Result, i int, cat string, rations []Ration, j int, data Data, errsC []ErrRTC) (b bool, f float64, ration Ration, err error) {
	for _, v := range rations {
		if r[i].ID == v.ID {
			var e []bool
			e, err = getErrRTC(errsC, v.ID, cat, v.Index)
			if err != nil {
				return
			}

			if !e[j] {
				ration = v
				f = data["ChAvg"+v.Channel]
				return
			}
		}

		if r[i].ID != v.ID {
			var indexR int
			indexR, err = getRByID(r, v.ID)
			if err != nil {
				return
			}

			var tb bool
			f, err = getDataByTime(r[indexR].D1, data["Time"], v.Channel)
			if err != nil {
				return
			}
			height := float64(r[indexR].S.Sensors[cat][v.Index].Height)

			for _, v1 := range r[indexR].S.Sensors[cat] {
				if v1.Channel != v.Channel {
					var f2 float64
					f2, err = getDataByTime(r[indexR].D1, data["Time"], v1.Channel)
					if err != nil {
						return
					}

					height1 := float64(v1.Height)

					if jC(f, f2, height, height1, cat) {
						tb = true
						break
					}
				}
			}

			if !tb {
				ration = v
				return
			}
		}
	}

	b = true
	return
}

func round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}
func getNewDate(year, month, hours int) time.Time {
	day := hours/24 + 1
	hour := hours % 24
	location, _ := time.LoadLocation("Local")

	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, location)
}
