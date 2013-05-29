package wind

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

func revises(r []Result, c Config) []Result {
	Info("---Revise---")
	for i, v := range r {
		r[i].RD = revise(v.S, v.D1, c)
	}

	return r
}

func revise(s Station, data []Data, c Config) []Data {
	// 1.Lost
	rData := rLost(s, data, c)

	// 2.合理性修订
	db := DB(rData)
	catR := []string{"wv", "wd"}
	for _, v := range catR {
		errs := getErrR(db, v, s.Sensors)
		rData = rRationality(v, errs, s.Sensors[v], rData)
	}

	// 3.趋势性修订
	db = DB(rData)
	catT := []string{"wv", "t", "p"}
	for _, v := range catT {
		errs := getErrT(db, v, s.Sensors)
		rData = rTrends(v, errs, s.Sensors[v], rData)
	}

	// 4.相关性修订
	db = DB(rData)
	catC := []string{"wv"}
	for _, v := range catC {
		errs := getErrC(db, v, s.Sensors)
		rData = rCorrelation(v, errs, s.Sensors[v], rData)
	}

	return rData
}

func rLost(s Station, data []Data, c Config) []Data {
	db := DB(data)

	monthData, hourData := getMonthHourData(s, db, c.AllowMinMax)

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
	var rData []Data
	for _, v := range s.Cm {
		dataMy := db.filter("My", v.My)

		// 依照dataMy的len需要细化

		k := 0
		for j := 0; j < v.Sbm; j++ {
			d := Data{}
			t := getNewDate(int(v.Year), int(v.Month), j)
			if len(dataMy) > k && float64(t.Unix())-dataMy[k]["Time"] > 0 {
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

	return rData
}

type Dt map[string][]float64
type DDt map[string]Dt
type DDDt map[string]DDt

func getMonthHourData(s Station, db DB, allowMinMax bool) (DDt, DDDt) {
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
			if allowMinMax {
				monthDataI["ChMin"+ch] = dbMD.get("ChMin" + ch)["ChMin"+ch]
				monthDataI["ChMax"+ch] = dbMD.get("ChMax" + ch)["ChMax"+ch]
			}
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
				if allowMinMax {
					hourDataIJ["ChMin"+ch] = dbHD.get("ChMin" + ch)["ChMin"+ch]
					hourDataIJ["ChMax"+ch] = dbHD.get("ChMax" + ch)["ChMax"+ch]
				}
			}
			hourDataI[h] = hourDataIJ
		}
		hourData[m] = hourDataI
	}

	return monthData, hourData
}

func rRationality(cat string, errs [][]bool, s []Sensor, data []Data) []Data {
	for i, v1 := range s {
		errI := errs[i]
		chI := v1.Channel

		for j, v2 := range data {
			if jR(v2["ChAvg"+chI], cat) {
				if !errI[j] {
					errI[j] = true
				}

				// 修订方法1，上下正常值的平均值
				if j > 0 && j < len(errI)-1 && !errI[j-1] && !errI[j+1] {
					data[j]["ChAvg"+chI] = (data[j-1]["ChAvg"+chI] + data[j+1]["ChAvg"+chI]) / 2
					continue
				}

				// 修订方法2，如果传感器多于1个，根据相关性来计算
				if len(s) > 1 {
					b, ch := rChannel(s[i].Rations, j, errs)
					if b {
						ration := v1.Rations[ch]
						chJ := ration.Channel

						data[j]["ChAvg"+chI] = ration.Slope*data[j]["ChAvg"+chJ] + ration.Intercept
						continue
					}
				}

				// 需要修订方法3，
			}
		}
	}

	return data
}
func rChannel(r []Ration, i int, errs [][]bool) (bool, int) {
	for j, v := range r {
		k := v.Index
		if !errs[k][i] {
			return true, j
		}
	}
	return false, 0
}

func rTrends(cat string, errs [][]bool, s []Sensor, data []Data) []Data {
	for i, v1 := range s {
		errI := errs[i]
		chI := v1.Channel

		for j := 0; j < len(data)-1; j++ {
			data1 := data[j]["ChAvg"+chI]
			data2 := data[j+1]["ChAvg"+chI]

			if jT(data1, data2, cat) {
				if !errI[j] {
					errI[j] = true
				}

				// 修订方法1，如果传感器多于1个，根据相关性来计算
				if len(s) > 1 {
					b, ch := tChannel(s[i].Rations, j, errs)
					if b {
						ration := v1.Rations[ch]
						chJ := ration.Channel

						data[j]["ChAvg"+chI] = ration.Slope*data[j]["ChAvg"+chJ] + ration.Intercept
						continue
					}
				}

				// 修订方法2，为之前两个值的平均值
				if j > 1 {
					data[j]["ChAvg"+chI] = (data[j-1]["ChAvg"+chI] + data[j-2]["ChAvg"+chI]) / 2
					continue
				}

				// 需要修订方法3，
			}
		}
	}

	return data
}
func tChannel(r []Ration, i int, errs [][]bool) (bool, int) {
	for j, v := range r {
		k := v.Index
		if i == 0 {
			if !errs[k][i] {
				return true, j
			}
			continue
		}
		if !errs[k][i] && !errs[k][i-1] {
			return true, j
		}
	}
	return false, 0
}

func rCorrelation(cat string, errs [][][]bool, s []Sensor, data []Data) []Data {
	for i := 0; i < len(s)-1; i++ {
		sI := s[i]
		chI := sI.Channel
		for j := i + 1; j < len(s); j++ {
			sJ := s[j]
			chJ := sJ.Channel

			for k, v := range data {
				dataI := v["ChAvg"+chI]
				dataJ := v["ChAvg"+chJ]
				heightI := sI.Height
				heightJ := sJ.Height

				if jC(dataI, dataJ, float64(heightI), float64(heightJ), cat) {
					if !errs[i][j][k] {
						errs[i][j][k] = true
						errs[j][i][k] = true
					}

					doneI, doneJ := false, false

					// 通道多于两个时采用相关性修订
					if len(s) > 2 {
						bI, cI := cChannel(s[i].Rations, i, j, k, errs)
						bJ, cJ := cChannel(s[j].Rations, j, i, k, errs)

						if bI {
							ration := sI.Rations[cI]
							ch := ration.Channel
							data[k]["ChAvg"+chI] = ration.Slope*data[k]["ChAvg"+ch] + ration.Intercept
							doneI = true
						}
						if bJ {
							ration := sJ.Rations[cJ]
							ch := ration.Channel
							data[k]["ChAvg"+chJ] = ration.Slope*data[k]["ChAvg"+ch] + ration.Intercept
							doneJ = true
						}
					}

					// 方法2
					if !doneI {
					}
					if !doneJ {
					}
				}
			}
		}
	}

	return data
}
func cChannel(r []Ration, indexI, indexJ, i int, errs [][][]bool) (bool, int) {
	for j, v := range r {
		k := v.Index
		if k == indexJ {
			continue
		}
		if !errs[indexI][k][i] {
			return true, j
		}
	}
	return false, 0
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
