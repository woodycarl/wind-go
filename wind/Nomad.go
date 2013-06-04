package wind

/*
import (
	""
)

func decInfoNomad(lines []string) (sensors []Sensor, logger Logger, site Site, linesR []string, err error) {

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
			re := regexp.MustCompile(`Site\sName:\s+([^,\"]+)`)
			td := re.FindStringSubmatch(lines[i])

			site = Site{
				Site: td[1],
			}
			continue
		}
		if strings.Contains(lines[i], "TimeStamp") {
			data := strings.Split(lines[i], ",")

			ch := 1

			for i := 1; i < len(data); i++ {
				re := regexp.MustCompile(`^([^\(]+)\((.+)\)(\s+@\s+(\d+)m|)\-\s*(\d+)\s+(min|hour)\s+(Vec\s+|)(Average|Max\sValue|Min\sValue|Std\sDev)$`)
				td := re.FindStringSubmatch(data[i])

				sensor := Sensor{
					Channel: strconv.Itoa(ch),
					Units:   td[2],
					Value:   td[8],
				}

				sensor.Height, err = strconv.ParseFloat(td[4], 64)
				if err != nil {
					return
				}

				sensors = append(sensors, sensor)
			}

			linesR = lines[i+1:]
			break
		}
	}
	return
}

func decDataCh(lines []string, s []Sensor, dec func([]string, []Sensor, int, chan ChDecData)) (r []Data) {
	interval := 15000
	length := int(math.Ceil(float64(len(lines)) / float64(interval)))

	ch := make(chan ChDecData, length)
	for i := 0; i < length-1; i++ {
		start := i * interval
		go dec(lines[start:start+interval], s, i, ch)
	}
	go dec(lines[(length-1)*interval:], s, length-1, ch)

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

func decDataNomad(lines []string, s []Sensor, index int, ch chan ChDecData) {
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
*/
