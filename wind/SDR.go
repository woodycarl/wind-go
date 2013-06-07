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

			switch sensor.Units {
			case "", "-----", "unit":
				sensor.NotInstalled = true
			}

			switch sensor.Description {
			case "No SCM Installed", "No Sensor", "Custom": // Custom
				sensor.NotInstalled = true
			}

			re := regexp.MustCompile(`^Height\s+([\d\.]+)[\s]*(m|ft)`)
			if re.MatchString(lines[i+5]) {
				td := re.FindStringSubmatch(lines[i+5])

				sensor.Height, err = strconv.ParseFloat(td[1], 64)
				if err != nil {
					return
				}

				// 单位转换
				if td[2] == "ft" {
					sensor.Height = sensor.Height * 0.3048
				}
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

			if data[start] != "" {
				tData["ChAvg"+js], chDecData.err = strconv.ParseFloat(data[start], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChAvg"+js, chDecData.err)
					ch <- chDecData
					return
				}
			}

			if data[start+1] != "" {
				tData["ChSd"+js], chDecData.err = strconv.ParseFloat(data[start+1], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChSd"+js, chDecData.err)
					ch <- chDecData
					return
				}
			}

			if data[start+2] != "" {
				tData["ChMax"+js], chDecData.err = strconv.ParseFloat(data[start+2], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChMax"+js, chDecData.err)
					ch <- chDecData
					return
				}
			}
			if data[start+3] != "" {
				tData["ChMin"+js], chDecData.err = strconv.ParseFloat(data[start+3], 64)
				if chDecData.err != nil {
					Error("decDataSDR: ChMin"+js, chDecData.err)
					ch <- chDecData
					return
				}
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
