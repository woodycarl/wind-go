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

type NomadSensor struct {
	Height      float64
	Description string
	Units       string
	Cat         string
}

func decInfoNomad(lines []string) (sensorsN []NomadSensor, sensors []Sensor, logger Logger, site Site, linesR []string, err error) {

	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Nomad2 Name") {
			re := regexp.MustCompile(`Nomad2\sName:(\d+)`)
			td := re.FindStringSubmatch(lines[i])

			logger = Logger{
				Serial: td[1],
			}

			continue
		}
		if strings.Contains(lines[i], "Site Name") {
			re := regexp.MustCompile(`Site\sName:\s+([^#,\"]+)`)
			td := re.FindStringSubmatch(lines[i])

			site = Site{
				Site: td[1],
			}
			continue
		}
		if strings.Contains(lines[i], "TimeStamp") {
			data := strings.Split(lines[i], ",")

			for i := 1; i < len(data); i++ {
				if len(data[i]) < 1 {
					continue
				}
				re := regexp.MustCompile(`^([^\(]+)\((.+)\)(\s+@\s+(\d+)m|)\-\s*(\d+)\s+(min|hour)\s+(Vec\s+|)(Sample|Averag|Max\sValue|Min\sValue|Std\sDev)`)
				td := re.FindStringSubmatch(data[i])

				sensor := NomadSensor{
					Description: td[1],
					Units:       td[2],
					Cat:         td[8],
				}

				if td[2] == "\xa1\xe3" {
					sensor.Units = "deg"
				}
				if td[2] == "\xa1\xe3C" {
					sensor.Units = "C"
				}

				if len(td[4]) > 0 {
					sensor.Height, err = strconv.ParseFloat(td[4], 64)
					if err != nil {
						return
					}
				}

				sensorsN = append(sensorsN, sensor)
			}

			linesR = lines[i+1:]
			break
		}
	}

	sensors = getSfSN(sensorsN)

	return
}

func getSfSN(sensorsN []NomadSensor) (s []Sensor) {
	for _, v := range sensorsN {
		if !existSN(s, v) {
			st := Sensor{
				Height:      v.Height,
				Description: v.Description,
				Units:       v.Units,
				Channel:     strconv.Itoa(len(s) + 1),
			}

			s = append(s, st)
		}
	}

	return
}

func existSN(s []Sensor, n NomadSensor) bool {
	for _, v := range s {
		if isSameNomadSensor(n, v) {
			return true
		}
	}

	return false
}

func decDataNomadch(lines []string, s []NomadSensor, ss []Sensor) (r []Data) {
	interval := 15000
	length := int(math.Ceil(float64(len(lines)) / float64(interval)))

	ch := make(chan ChDecData, length)
	for i := 0; i < length-1; i++ {
		start := i * interval
		go decDataNomad(lines[start:start+interval], s, ss, i, ch)
	}
	go decDataNomad(lines[(length-1)*interval:], s, ss, length-1, ch)

	chData := map[int][]Data{}
	for i := 0; i < length; i++ {
		tData := <-ch
		chData[tData.index] = tData.data
	}
	for i := 0; i < length; i++ {
		r = append(r, chData[i]...)
	}

	go saveRData(strconv.Itoa(g()), r, ss)

	return
}

func decDataNomad(lines []string, s []NomadSensor, sensors []Sensor, index int, ch chan ChDecData) {
	chDecData := ChDecData{
		index: index,
	}

	var r []Data

	for i := 0; i < len(lines); i++ {

		data := strings.Split(lines[i], ",")

		if len(data) < 4 {
			// 非数据行
			continue
		}

		re := regexp.MustCompile(`^(\"|)\d{4}[\-|\/]\d{1,2}[\-|\/]\d{1,2}(\"|)$`)
		if re.MatchString(data[0]) {
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
			var channel string
			var chs string

			channel = getNomadCh(v, sensors)

			switch v.Cat {
			case "Averag", "Sample": // Averag|Max\sValue|Min\sValue|Std\sDev
				chs = "ChAvg" + channel
			case "Max Value":
				chs = "ChMax" + channel
			case "Min Value":
				chs = "ChMin" + channel
			case "Std Dev":
				chs = "ChSd" + channel
			default:
				chDecData.err = errors.New("decDataNomad: cat no match!")
				ch <- chDecData
				return
			}

			tData[chs], chDecData.err = strconv.ParseFloat(strings.TrimSpace(data[j+1]), 64)
			if chDecData.err != nil {
				ch <- chDecData
				return
			}
		}

		r = append(r, tData)
	}

	chDecData.data = r
	ch <- chDecData
}

func getNomadCh(s NomadSensor, ss []Sensor) (ch string) {
	for _, v := range ss {
		if isSameNomadSensor(s, v) {
			ch = v.Channel
			return
		}
	}

	return
}

func isSameNomadSensor(s1 NomadSensor, s2 Sensor) bool {
	if s1.Height == s2.Height && s1.Description == s2.Description && s1.Units == s2.Units {
		return true
	}

	return false
}
