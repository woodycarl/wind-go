package wind

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

type RawData struct {
	s     Station
	data  []Data
	is10m bool
	err   error
}

func decRaw(data [][]string, c Config) (r []Result, err error) {
	Info("DecRaw Start")
	timeS := time.Now()

	var chRaw = make(chan RawData, len(data))
	for _, v := range data {
		go DecodeData(v, c, chRaw)
	}

	for _, _ = range data {
		v := <-chRaw

		if v.err != nil {
			err = v.err
			return
		}

		if i, b := existStation(v.s, r); b {
			switch {
			case v.is10m && (len(r[i].D2) > 0 || len(r[i].D1) == 0):
				fallthrough
			case !v.is10m && (len(r[i].D1) > 0 || len(r[i].D2) == 0):
				err = errors.New("file cannot match!")
				return
			}
			if v.is10m {
				r[i].D2 = v.data
			} else {
				r[i].D1 = v.data
			}
			continue
		}

		res := Result{
			S: v.s,
		}
		if v.is10m {
			res.D2 = v.data
		} else {
			res.D1 = v.data
		}

		r = append(r, res)
	}

	for i, _ := range r {
		r[i].ID = r[i].S.Site.Site
	}

	Info("DecRaw End", time.Now().Sub(timeS))
	return
}

func existStation(s Station, r []Result) (int, bool) {
	if len(r) == 0 {
		return 0, false
	}
	for i, v := range r {
		if isSameStation(s, v.S) {
			return i, true
		}
	}
	return 0, false
}

func isSameStation(s1, s2 Station) bool {
	switch {
	case s1.System != s2.System:
		return false
	case s1.Version != s2.Version:
		return false
	case s1.Logger != s2.Logger:
		return false
	case s1.Site != s2.Site:
		return false
	}
	return true
}

// 解析原始数据，返回到 channel
func DecodeData(lines []string, c Config, ch chan RawData) {
	Info("---Decode Data ---", len(lines))
	var chData RawData
	var linesR []string
	var data []Data

	if len(lines) < 100 {
		chData.err = errors.New("DecodeData: not enough data")
		ch <- chData
		return
	}

	t := strings.Split(lines[0], "\t")
	s := Station{
		System:  t[0],
		Version: t[1],
	}

	switch s.System {
	case "SDR":
		s.SensorsR, s.Logger, s.Site, linesR, chData.err = decInfoSDR(lines)
		if chData.err != nil {
			ch <- chData
			return
		}
		data = decDataSDRch(linesR, s.SensorsR)
	default:
		chData.err = errors.New("DecodeData: system not surport!")
		ch <- chData
		return
	}

	for _, v := range data {
		t := time.Unix(int64(v["Time"]), 0)
		if t.Minute() != 0 {
			chData.is10m = true
			break
		}
	}

	chData.data = data
	chData.s = s
	ch <- chData
}

func decInfoSDR(lines []string) (sensors []Sensor, logger Logger, site Site, linesR []string, err error) {
	//Info("dec info")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Channel") {
			sensor := Sensor{
				Channel:      getLineStr(lines[i]),
				Type:         getLineStr(lines[i+1]),
				Description:  getLineStr(lines[i+2]),
				Details:      getLineStr(lines[i+3]),
				SerialNumber: getLineStr(lines[i+4]),
				ScaleFactor:  getLineStr(lines[i+6]),
				Offset:       getLineStr(lines[i+7]),
				Units:        getLineStr(lines[i+8]),
			}
			if len(sensor.Channel) < 1 {
				err = errors.New("sensor channel empty!")
				return
			}

			if len(sensor.Units) < 1 {
				err = errors.New("sensor units empty!")
				return
			}

			sensor.Height, err = strconv.Atoi(strings.Fields(getLineStr(lines[i+5]))[0])
			if err != nil && sensor.Description != "No SCM Installed" && getLineStr(lines[i+5]) != "m" {
				// 需要增加没有高度值时的处理
				Error("decInfoSDR: sensor height get err!", err)
				return
			}
			err = nil

			sensors = append(sensors, sensor)
			i = i + 8
			continue
		}
		if strings.Contains(lines[i], "Logger") {
			logger = Logger{
				Model:       getLineStr(lines[i+1]),
				Serial:      getLineStr(lines[i+2]),
				HardwareRev: getLineStr(lines[i+3]),
			}
			i = i + 3
			continue
		}
		if strings.Contains(lines[i], "Site") {
			site = Site{
				Site:         getLineStr(lines[i+1]),
				SiteDesc:     getLineStr(lines[i+2]),
				ProjectCode:  getLineStr(lines[i+3]),
				ProjectDesc:  getLineStr(lines[i+4]),
				SiteLocation: getLineStr(lines[i+5]),
				Latitude:     getLineStr(lines[i+7]),
				Longitude:    getLineStr(lines[i+8]),
				TimeOffset:   getLineStr(lines[i+9]),
			}

			site.SiteElevation, err = strconv.ParseFloat(getLineStr(lines[i+6]), 64)
			if err != nil && getLineStr(lines[i+6]) != "" {
				// 需要增加为 "" 时的处理
				Error("decInfoSDR: site elevation get err!", err)
				return
			}
			err = nil

			i = i + 9
			continue
		}
		if strings.Contains(lines[i], "Date") {
			linesR = lines[i+1:]
			break
		}
	}
	return
}

type ChDecData struct {
	index int
	err   error
	data  []Data
}

func decDataSDRch(lines []string, s []Sensor) (r []Data) {
	//Info("dec data")
	interval := 15000
	length := int(math.Ceil(float64(len(lines)) / float64(interval)))

	ch := make(chan ChDecData, length)
	for i := 0; i < length-1; i++ {
		start := i * interval
		go decDataSDR(lines[start:start+interval], s, i, ch)
	}
	go decDataSDR(lines[(length-1)*interval:], s, length-1, ch)

	chData := map[int][]Data{}
	for i := 0; i < length; i++ {
		tData := <-ch
		chData[tData.index] = tData.data
	}
	for i := 0; i < length; i++ {
		r = append(r, chData[i]...)
	}

	return
}
func decDataSDR(lines []string, s []Sensor, index int, ch chan ChDecData) {
	chDecData := ChDecData{
		index: index,
	}

	var r []Data

	for i := 0; i < len(lines); i++ {
		data := strings.Split(lines[i], "\t")

		if len(data) < 4 {
			// 非数据行
			continue
		}

		var t time.Time
		var my float64

		t, my, chDecData.err = decodeDate(data[0])
		if chDecData.err != nil {
			Error("decodeDate", chDecData.err)
			ch <- chDecData
			return
		}

		tData := Data{
			"Time":  float64(t.Unix()),
			"Hour":  float64(t.Hour()),
			"My":    my,
			"Year":  float64(t.Year()),
			"Month": float64(t.Month()),
		}

		for j, v := range s {
			if v.Description == "No SCM Installed" {
				continue
			}

			start := j*4 + 1
			js := v.Channel

			tData["ChAvg"+js], chDecData.err = strconv.ParseFloat(data[start], 64)
			if chDecData.err != nil {
				Error("decDataSDR: ChAvg"+js, chDecData.err)
				ch <- chDecData
				return
			}

			tData["ChSd"+js], chDecData.err = strconv.ParseFloat(data[start+1], 64)
			if chDecData.err != nil {
				Error("decDataSDR: ChSd"+js, chDecData.err)
				ch <- chDecData
				return
			}
			/*
				tData["ChMax"+js], chDecData.err = strconv.ParseFloat(data[start+2], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChMax"+js, chDecData.err)
					ch <- chDecData
					return
				}

				tData["ChMin"+js], chDecData.err = strconv.ParseFloat(data[start+4], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChMin"+js, chDecData.err)
					ch <- chDecData
					return
				}*/
		}

		r = append(r, tData)
	}

	chDecData.data = r
	ch <- chDecData
}

func getLineStr(str string) string {
	return strings.Split(str, "\t")[1]
}

func decodeDate(data string) (t time.Time, f float64, err error) {
	re := regexp.MustCompile(`^(\d{4})\/(\d{1,2})\/(\d{1,2})(\s\w+|)\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)$`)
	if !re.MatchString(data) {
		err = errors.New("date format err")
		return
	}

	td := re.FindStringSubmatch(data)

	var year, month, date, hour, minute int
	year, err = strconv.Atoi(td[1])
	if err != nil {
		return
	}

	month, err = strconv.Atoi(td[2])
	if err != nil {
		return
	}

	date, err = strconv.Atoi(td[3])
	if err != nil {
		return
	}

	hour, err = strconv.Atoi(td[5])
	if err != nil {
		return
	}

	minute, err = strconv.Atoi(td[6])
	if err != nil {
		return
	}

	location, _ := time.LoadLocation("Local")
	t = time.Date(year, time.Month(month), date, hour, minute, 0, 0, location)

	f, err = strconv.ParseFloat(td[1]+td[2], 64)
	if err != nil {
		return
	}
	return
}

func sensorClassify(sensorsR []Sensor) map[string]([]Sensor) {
	Info("sensors classify")
	sensors := map[string]([]Sensor){}
	for _, v := range sensorsR {
		units := map[string]string{
			"m/s":   "wv",
			"deg":   "wd",
			"Volts": "vol",
			"%RH":   "h",
			"C":     "t",
			"kPa":   "p",
		}

		sensor := units[v.Units]

		sensors[sensor] = append(sensors[sensor], v)
	}
	return sensors
}

func getAm(data []Data) (ams []Am, err error) {
	Info("get am")
	for _, v := range data {
		if !existAm(ams, v["My"]) {
			am := Am{
				My:    v["My"],
				Year:  v["Year"],
				Month: v["Month"],
			}
			ams = append(ams, am)
		}
	}

	// Am 排序

	// 补足缺失月份
	ams, err = rLostAm(ams)

	return
}
func existAm(ams []Am, my float64) bool {
	for _, v := range ams {
		if v.My == my {
			return true
		}
	}

	return false
}

func rLostAm(ams []Am) (ams2 []Am, err error) {
	location, _ := time.LoadLocation("Local")

	ams2 = append(ams2, ams[0])

	for i := 1; i < len(ams); {
		j := len(ams2) - 1
		diff := ams[i].My - ams2[j].My

		if (diff < 89 && diff > 1) || diff > 89 {
			t := time.Date(int(ams2[j].Year), time.Month(int(ams2[j].Month)+2), 0, 0, 0, 0, 0, location)

			am := Am{
				Year:     float64(t.Year()),
				Month:    float64(t.Month()),
				NotExist: true,
			}
			am.My, err = strconv.ParseFloat(t.Format("200601"), 64)
			if err != nil {
				return
			}

			ams2 = append(ams2, am)

			Warn("add am", am.My)
		} else {
			ams2 = append(ams2, ams[i])
			i = i + 1
		}
	}

	return
}
