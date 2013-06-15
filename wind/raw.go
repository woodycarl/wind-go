package wind

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
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

		if err = v.err; err != nil {
			return
		}

		if i, b := existStation(v.s, r); b {
			switch {
			case v.is10m && (len(r[i].D2) > 0 || len(r[i].D1) == 0):
				fallthrough
			case !v.is10m && (len(r[i].D1) > 0 || len(r[i].D2) == 0):
				err = errors.New("decRaw file cannot match!")
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
		if len(v.D1) < 1 {
			Warn(v.S.Site.Site+": no 1h data! gen from 10m data.", len(v.D2))
			r[i].D1, err = genD1fD2(v.D2, v.S.SensorsR)
			if err != nil || len(r[i].D1) < 1 {
				Error(v.S.Site.Site + "can not gen 1h data!")
				err = errors.New(v.S.Site.Site + ": can not gen 1h data!")
				return
			}
			Info(v.S.Site.Site+": gen D1", len(r[i].D1))
		}
	}

	chR := make(chan ChAjustR, len(r))
	for _, v := range r {
		go adjustR(v, chR)
	}
	for i, _ := range r {
		tD := <-chR
		if err = tD.err; err != nil {
			return
		}
		r[i] = tD.r
	}

	Info("DecRaw End", time.Now().Sub(timeS))
	return
}

func existStation(s Station, r []Result) (int, bool) {
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
	default:
		return true
	}
}

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

	var s Station

	switch {
	case strings.Contains(lines[0], "SDR"):
		s.System = "SDR"
		s.Version = strings.Fields(lines[0])[1]

		s.SensorsR, s.Logger, s.Site, linesR, chData.err = decInfoSDR(lines)
		if chData.err != nil {
			ch <- chData
			return
		}
		data = decDataSDRch(linesR, s.SensorsR)
	case strings.Contains(lines[0], "Multi-Track Export -"):
		s.System = "Nomad2"

		var ss []NomadSensor
		ss, s.SensorsR, s.Logger, s.Site, linesR, chData.err = decInfoNomad(lines)
		if chData.err != nil {
			ch <- chData
			return
		}
		data = decDataNomadch(linesR, ss, s.SensorsR)
	default:
		chData.err = errors.New("DecodeData: file system format err!")
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

type ChDecData struct {
	index int
	err   error
	data  []Data
}

func decodeDate(data string) (t time.Time, err error) {
	if re1 := regexp.MustCompile(`(\d{4})[\/|-](\d{1,2})[\/|-](\d{1,2})(\s\w+|)\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)`); re1.MatchString(data) {
		td := re1.FindStringSubmatch(data)

		t, err = NewDate(td[1], td[2], td[3], td[5], td[6])

	} else if re2 := regexp.MustCompile(`(\d{1,2})[\/|-](\d{1,2})[\/|-](\d{4})\s(\d{1,2}):(\d{1,2})(:\d{1,2}|)`); re2.MatchString(data) {
		td := re2.FindStringSubmatch(data)

		t, err = NewDate(td[3], td[1], td[2], td[4], td[5])

	} else {
		err = errors.New("decodeDate date format err" + data)
		return
	}

	return
}

func NewDate(y, mon, d, h, min string) (t time.Time, err error) {
	var year, month, date, hour, minute int

	year, err = strconv.Atoi(y)
	if err != nil {
		return
	}

	month, err = strconv.Atoi(mon)
	if err != nil {
		return
	}

	date, err = strconv.Atoi(d)
	if err != nil {
		return
	}

	hour, err = strconv.Atoi(h)
	if err != nil {
		return
	}

	minute, err = strconv.Atoi(min)
	if err != nil {
		return
	}

	t = time.Date(year, time.Month(month), date, hour, minute, 0, 0, LOCATION)
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
			"mB":        "p",
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
	if len(ams) < 1 {
		err = errors.New("rLostAm not enough data!")
		return
	}

	ams2 = []Am{ams[0]}

	for i := 1; i < len(ams); {
		j := len(ams2) - 1
		diff := ams[i].My - ams2[j].My

		if diff < 1 {
			err = errors.New("rLostAm am order err!")
			return
		} else if diff == 89 || diff == 1 {
			ams2 = append(ams2, ams[i])
			i = i + 1
		} else {
			// if (diff < 89 && diff > 1) || diff > 89
			t := time.Date(int(ams2[j].Year), time.Month(int(ams2[j].Month)+2), 0, 0, 0, 0, 0, LOCATION)

			am := Am{
				Year:     float64(t.Year()),
				Month:    float64(t.Month()),
				NotExist: true,
			}
			am.My, err = strconv.ParseFloat(t.Format(DATE_FORMAT_MY), 64)
			if err != nil {
				Error("rLostAm", err)
				return
			}

			ams2 = append(ams2, am)

			Warn("add am", am.My)
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
			r.D1 = adjustRF(r.D1, v.Channel)
			r.D2 = adjustRF(r.D2, v.Channel)
		case "mb", "mB", "MB":
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
func adjustRF(data []Data, ch string) []Data {
	for i, _ := range data {
		data[i]["ChAvg"+ch] = (data[i]["ChAvg"+ch] - 32.0) / 1.8
	}
	return data
}

func genD1fD2(d2 []Data, s []Sensor) (d1 []Data, err error) {
	// 假设数据按是严格时间顺序排列
	var ds DB
	var val time.Time

	for i, v := range d2 {
		t := time.Unix(int64(v["Time"]), 0)
		if i == 0 {
			val = t
		}

		if val.Format(DATE_FORMAT_YMHM) == t.Format(DATE_FORMAT_YMHM) && i != len(d2)-1 {
			ds = append(ds, v)
			continue
		}

		t1 := time.Date(val.Year(), val.Month(), val.Day(), val.Hour(), 0, 0, 0, LOCATION)

		var data Data
		data, err = NewData(t1)
		if err != nil {
			return
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
		ds = DB{v}
	}

	return
}

func NewData(t time.Time) (data Data, err error) {
	my, err := strconv.ParseFloat(t.Format(DATE_FORMAT_MY), 64)
	if err != nil {
		return
	}

	data = Data{
		"Time":  float64(t.Unix()),
		"Hour":  float64(t.Hour()),
		"My":    my,
		"Day":   float64(t.Day()),
		"Year":  float64(t.Year()),
		"Month": float64(t.Month()),
	}

	return
}
