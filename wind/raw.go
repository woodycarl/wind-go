package wind

import (
	"errors"
	"math"
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
		chData.err = errors.New("not enough data")
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
		s.SensorsR, s.Logger, s.Site, linesR = decInfoSDR(lines)
		data = decDataSDRch(linesR, c.AllowMinMax)
	default:
		chData.err = errors.New("system not surport!")
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

func decInfoSDR(lines []string) (sensors []Sensor, logger Logger, site Site, linesR []string) {
	//Info("dec info")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Channel") {
			sensor := Sensor{
				Channel:      getLineStr(lines[i]),
				Type:         getLineInt(lines[i+1]),
				Description:  getLineStr(lines[i+2]),
				Details:      getLineStr(lines[i+3]),
				SerialNumber: getLineStr(lines[i+4]),
				Height:       parseInt(strings.Split(getLineStr(lines[i+5]), " ")[0]),
				ScaleFactor:  getLineFloat(lines[i+6]),
				Offset:       getLineFloat(lines[i+7]),
				Units:        getLineStr(lines[i+8]),
			}
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
				Site:          getLineStr(lines[i+1]),
				SiteDesc:      getLineStr(lines[i+2]),
				ProjectCode:   getLineStr(lines[i+3]),
				ProjectDesc:   getLineStr(lines[i+4]),
				SiteLocation:  getLineStr(lines[i+5]),
				SiteElevation: getLineFloat(lines[i+6]),
				Latitude:      getLineStr(lines[i+7]),
				Longitude:     getLineStr(lines[i+8]),
				TimeOffset:    getLineFloat(lines[i+9]),
			}
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
	data  []Data
}

func decDataSDRch(lines []string, allowMinMax bool) (r []Data) {
	//Info("dec data")
	interval := 15000
	length := int(math.Ceil(float64(len(lines)) / float64(interval)))

	ch := make(chan ChDecData, length)
	for i := 0; i < length-1; i++ {
		start := i * interval
		go decDataSDR(lines[start:start+interval], allowMinMax, i, ch)
	}
	go decDataSDR(lines[(length-1)*interval:], allowMinMax, length-1, ch)

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
func decDataSDR(lines []string, allowMinMax bool, index int, ch chan ChDecData) {
	var r []Data
	for i := 0; i < len(lines); i++ {
		data := strings.Split(lines[i], "\t")

		if len(data) < 4 {
			// 非数据行
			continue
		}

		td := strings.Split(data[0], " ")
		var t time.Time
		var my float64
		switch len(td) {
		case 2:
			t, my = decodeDate(td[0], td[1])
		case 3:
			t, my = decodeDate(td[0], td[2])
		}

		tData := Data{
			"Time":  float64(t.Unix()),
			"Hour":  float64(t.Hour()),
			"My":    my,
			"Year":  float64(t.Year()),
			"Month": float64(t.Month()),
		}

		for j := 1; j <= len(data)/4; j++ {
			// if Description=="No SCM Installed" {continue}
			var start = (j-1)*4 + 1
			sj := strconv.Itoa(j)
			tData["ChAvg"+sj] = parseFloat(data[start])
			tData["ChSd"+sj] = parseFloat(data[start+1])
			if allowMinMax {
				tData["ChMax"+sj] = parseFloat(data[start+2])
				tData["ChMin"+sj] = parseFloat(data[start+3])
			}

		}

		r = append(r, tData)
	}

	chDecData := ChDecData{
		index: index,
		data:  r,
	}
	ch <- chDecData
}

func getLineStr(str string) string {
	return strings.Split(str, "\t")[1]
}
func getLineInt(str string) int {
	return parseInt(getLineStr(str))
}
func getLineFloat(str string) float64 {
	return parseFloat(getLineStr(str))
}
func parseInt(str string) int {
	r, _ := strconv.Atoi(str)
	return r
}
func parseFloat(str string) float64 {
	r, _ := strconv.ParseFloat(str, 64)
	return r
}

func decodeDate(str1, str2 string) (time.Time, float64) {
	a := strings.Split(str1, "/")
	b := strings.Split(str2, ":")

	location, _ := time.LoadLocation("Local")
	t := time.Date(parseInt(a[0]), time.Month(parseInt(a[1])), parseInt(a[2]), parseInt(b[0]), parseInt(b[1]), 0, 0, location)
	return t, parseFloat(a[0] + a[1])
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

func getAm(data []Data) (ams []Am) {
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
	ams = rLostAm(ams)

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

func rLostAm(ams []Am) []Am {
	location, _ := time.LoadLocation("Local")
	ams2 := []Am{ams[0]}

	for i := 1; i < len(ams); {
		j := len(ams2) - 1
		diff := ams[i].My - ams2[j].My

		if (diff < 89 && diff > 1) || diff > 89 {
			t := time.Date(int(ams2[j].Year), time.Month(int(ams2[j].Month)+2), 0, 0, 0, 0, 0, location)

			am := Am{
				My:       parseFloat(t.Format("200601")),
				Year:     float64(t.Year()),
				Month:    float64(t.Month()),
				NotExist: true,
			}

			ams2 = append(ams2, am)

			Warn("add am", am.My)
		} else {
			ams2 = append(ams2, ams[i])
			i = i + 1
		}
	}

	return ams2
}
