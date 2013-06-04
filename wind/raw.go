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
type ChAjustR struct {
	err error
	r   Result
}

func decRaw(data [][]string) (r []Result, err error) {
	Info("DecRaw Start")
	timeS := time.Now()

	var chRaw = make(chan RawData, len(data))
	for _, v := range data {
		go DecodeData(v, chRaw)
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

	for i, v := range r {
		Info(len(v.D1), len(v.D2))
		if len(v.D1) < 1 {
			Warn(v.S.Site.Site + ": no 1h data! gen from 10m data.")
			r[i].D1 = genD1fD2(v.D2, v.S.SensorsR)
			if len(r[i].D1) < 1 {
				Error(v.S.Site.Site + "can not gen 1h data!")
				err = errors.New(v.S.Site.Site + ": can not gen 1h data!")
				return
			}
			Info(len(r[i].D1), "save D1")
			// go saveRData(strconv.Itoa(g()), r[i].D1, r[i].S.SensorsR)
		}

		if len(v.D2) < 1 {
			Warn(v.S.Site.Site + ": no 10m data!")
			// err = errors.New(v.S.Site.Site + ": no 10m data!")
			// return
		}
	}

	chR := make(chan ChAjustR, len(r))
	for _, v := range r {
		go adjustR(v, chR)
	}
	for i, _ := range r {
		tD := <-chR
		if tD.err != nil {
			err = tD.err
			return
		}
		r[i] = tD.r
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
func DecodeData(lines []string, ch chan RawData) {
	Info("---Decode Data ---", len(lines))
	var chData RawData
	var linesR []string
	var data []Data

	// 需要增加const
	if len(lines) < 100 {
		chData.err = errors.New("DecodeData: not enough data!")
		ch <- chData
		return
	}

	t := strings.Split(lines[0], "\t")
	if len(t) < 2 {
		chData.err = errors.New("DecodeData: file system format err!")
		ch <- chData
		return
	}
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

			/*
				if len(sensor.Units) < 1 {
					err = errors.New("sensor units empty!")
					return
				}
			*/

			switch sensor.Units {
			case "", "-----", "unit":
				sensor.NotInstalled = true
			}

			switch sensor.Description {
			case "No SCM Installed", "No Sensor", "Custom": // Custom
				sensor.NotInstalled = true
			}

			/*
				re := regexp.MustCompile(`^Height\s+([\d\.]+)(\s+|)(m|ft)`)
			*/
			sensor.Height, err = strconv.ParseFloat(strings.Fields(getLineStr(lines[i+5]))[0], 64)
			if err != nil && sensor.NotInstalled != true {
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

	// go saveRData(strconv.Itoa(g()), r, s)

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
			"Day":   float64(t.Day()),
			"Year":  float64(t.Year()),
			"Month": float64(t.Month()),
		}

		for j, v := range s {
			if v.Description == "No SCM Installed" {
				continue
			}

			start := j*4 + 1
			if start+3 >= len(data) {
				continue
			}
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

			tData["ChMax"+js], chDecData.err = strconv.ParseFloat(data[start+2], 64)
			if chDecData.err != nil {
				Error("decDataSDR: ChMax"+js, chDecData.err)
				ch <- chDecData
				return
			}

			tData["ChMin"+js], chDecData.err = strconv.ParseFloat(data[start+3], 64)
			if chDecData.err != nil {
				Error("decDataSDR: ChMin"+js, chDecData.err)
				ch <- chDecData
				return
			}
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
	var year, month, date, hour, minute int
	var my string
	location, _ := time.LoadLocation("Local")

	re := regexp.MustCompile(`^(\d{4})[\/|-](\d{1,2})[\/|-](\d{1,2})(\s\w+|)\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)$`)
	if re.MatchString(data) {
		td := re.FindStringSubmatch(data)

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

		my = td[1] + td[2]
	} else if re = regexp.MustCompile(`^(\d{1,2})[\/|-](\d{1,2})[\/|-](\d{4})\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)$`); re.MatchString(data) {
		td := re.FindStringSubmatch(data)

		year, err = strconv.Atoi(td[3])
		if err != nil {
			return
		}

		month, err = strconv.Atoi(td[1])
		if err != nil {
			return
		}

		date, err = strconv.Atoi(td[2])
		if err != nil {
			return
		}

		hour, err = strconv.Atoi(td[4])
		if err != nil {
			return
		}

		minute, err = strconv.Atoi(td[5])
		if err != nil {
			return
		}

		my = td[3] + td[1]
	} else {
		err = errors.New("date format err")
		return
	}

	t = time.Date(year, time.Month(month), date, hour, minute, 0, 0, location)

	f, err = strconv.ParseFloat(my, 64)
	return
}

func sensorClassify(sensorsR []Sensor) map[string]([]Sensor) {
	Info("sensors classify")
	sensors := map[string]([]Sensor){}
	for _, v := range sensorsR {
		units := map[string]string{
			"m/s":       "wv",
			"mph":       "wv",
			"deg":       "wd",
			"Degrees":   "wd",
			"Volts":     "vol",
			"v":         "vol",
			"%RH":       "h",
			"C":         "t",
			"Degrees F": "t",
			"F":         "t",
			"kPa":       "p",
			"mb":        "p",
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
				Error("rLostAm", err)
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

func adjustR(r Result, ch chan ChAjustR) {
	var chAjustR ChAjustR

	r.ID = r.S.Site.Site
	r.S.Sensors = sensorClassify(r.S.SensorsR)
	r.S.Am, chAjustR.err = getAm(r.D1)
	if chAjustR.err != nil {
		ch <- chAjustR
		return
	}

	for _, v := range r.S.SensorsR {
		switch v.Units {
		case "mph":
			r.D1 = adjustRTimes(r.D1, v.Channel, 1.6/3.6)
			r.D2 = adjustRTimes(r.D2, v.Channel, 1.6/3.6)
		case "Degrees F", "F":
			r.D1 = adjustRAdd(r.D1, v.Channel, -273.15)
			r.D2 = adjustRAdd(r.D2, v.Channel, -273.15)
		case "mb":
			r.D1 = adjustRTimes(r.D1, v.Channel, 0.1)
			r.D2 = adjustRTimes(r.D2, v.Channel, 0.1)
		}
	}

	chAjustR.r = r
	ch <- chAjustR
}

func adjustRTimes(data []Data, ch string, t float64) []Data {
	for i, _ := range data {
		data[i]["ChAvg"+ch] = data[i]["ChAvg"+ch] * t
	}
	return data
}
func adjustRAdd(data []Data, ch string, t float64) []Data {
	for i, _ := range data {
		data[i]["ChAvg"+ch] = data[i]["ChAvg"+ch] + t
	}
	return data
}

func genD1fD2(d2 []Data, s []Sensor) (d1 []Data) {
	// 假设数据按是严格时间顺序排列
	location, _ := time.LoadLocation("Local")
	var ds DB
	var val time.Time
	var my float64

	for i, v := range d2 {
		t := time.Unix(int64(v["Time"]), 0)
		if i == 0 {
			val = t
			my = v["My"]
		}
		if i == len(d2)-1 {
			// 最后一个数据，不要了，写起来真麻烦
		}
		if val.Format("2006-01-02-15") == t.Format("2006-01-02-15") {
			ds = append(ds, v)
			continue
		}
		t1 := time.Date(val.Year(), val.Month(), val.Day(), val.Hour(), 0, 0, 0, location)
		data := Data{
			"Time":  float64(t1.Unix()),
			"Hour":  float64(t1.Hour()),
			"My":    my,
			"Day":   float64(t1.Day()),
			"Year":  float64(t1.Year()),
			"Month": float64(t1.Month()),
		}
		for _, v1 := range s {
			ch := v1.Channel
			data["ChAvg"+ch] = ArrayAvg(ds.get("ChAvg" + ch)["ChAvg"+ch])
			data["ChSd"+ch] = ArrayAvg(ds.get("ChSd" + ch)["ChSd"+ch])
			data["ChMin"+ch] = ArrayMin(ds.get("ChMin" + ch)["ChMin"+ch])
			data["ChMax"+ch] = ArrayMax(ds.get("ChMax" + ch)["ChMax"+ch])
		}

		d1 = append(d1, data)

		val = t
		my = v["My"]
		ds = DB{}
	}

	return
}
