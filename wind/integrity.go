package wind

import (
	"errors"
	"math"
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

func integrities(r []Result, c Config) (rr []Result, err error) {
	Info("---Integrity---")
	timeS := time.Now()

	for i, v := range r {
		r[i].S, err = integrity(v.S, v.D1, c)
		if err != nil {
			return
		}

		Info(r[i].ID, r[i].S.Cm[0].My, r[i].S.Cm[11].My)
	}

	Info("Integrity", time.Now().Sub(timeS))

	rr = r
	return
}

func integrity(s Station, data []Data, c Config) (ss Station, err error) {
	s.Am = calErrs(s.Am, data, s.Sensors, c)

	s.Cm, err = chooseOneYear(s.Am)
	if err != nil {
		return
	}

	ss = s
	return
}

/*
	//替换上面代码后从 440ms -> 340ms
	type ChIntegrity struct {
		index int
		s     Station
	}

	func integrities(r []Result, c Config) []Result {
		Info("---Integrity---")
		timeS := time.Now()

		chIntegrity := make(chan ChIntegrity, len(r))
		for i, v := range r {
			go integrity(i, v.S, v.D1, c, chIntegrity)
		}

		for _, _ = range r {
			tD := <-chIntegrity
			r[tD.index].S = tD.s
			Info(r[tD.index].ID, r[tD.index].S.Cm[0].My, r[tD.index].S.Cm[11].My)
		}

		Info("Integrity", time.Now().Sub(timeS))
		return r
	}

	func integrity(i int, s Station, data []Data, c Config, ch chan ChIntegrity) {
		s.Am = calErrs(s.Am, data, s.Sensors, c)
		s.Am = calIntegrity(s.Am)
		s.Cm = chooseOneYear(s.Am)

		chData := ChIntegrity{
			index: i,
			s:     s,
		}

		ch <- chData
	}
*/

func calErrs(am []Am, data []Data, s map[string][]Sensor, c Config) []Am {
	db := DB(data)

	for i, v := range am {
		dbMy := db.filter("My =", v.My)

		errRwv := calErrNumRT(getErrR(dbMy, "wv", s["wv"]))
		errRwd := calErrNumRT(getErrR(dbMy, "wd", s["wd"]))
		errTwv := calErrNumRT(getErrT(dbMy, "wv", s["wv"]))
		errTt := calErrNumRT(getErrT(dbMy, "t", s["t"]))
		errTp := calErrNumRT(getErrT(dbMy, "p", s["p"]))
		errCwv := calErrNumC(getErrC(dbMy, "wv", s["wv"]))
		errCwd := calErrNumC(getErrC(dbMy, "wd", s["wd"]))

		am[i].Rwv = trueNum(errRwv)
		am[i].Rwd = trueNum(errRwd)
		am[i].Twv = trueNum(errTwv)
		am[i].Tt = trueNum(errTt)
		am[i].Tp = trueNum(errTp)
		am[i].Cwv = trueNum(errCwv)
		am[i].Cwd = trueNum(errCwd)

		days := countDays(int(am[i].Year), int(am[i].Month))
		sbm := days * 24
		lost := sbm - len(dbMy)

		var x, y [][]bool
		x = append(x, errRwv, errRwd, errCwv, errCwd)
		y = append(y, errTwv, errTt, errTp)
		errx := calErrNumRT(x)
		erry := calErrNumRT(y)
		err := arrayOrT(erry, errx)
		am[i].All = trueNum(err) + lost

		am[i].Lost = lost
		am[i].Sbm = sbm

		am[i].Sr = 100 * float64(am[i].All) / float64(sbm)
	}

	return am
}

func getErrR(db DB, cat string, s []Sensor) (errs [][]bool) {

	for _, v := range s {
		ch := v.Channel
		data := db.get("ChAvg" + ch)["ChAvg"+ch]

		err := jRs(data, cat)

		errs = append(errs, err)
	}

	return
}
func jRs(data []float64, cat string) (err []bool) {
	for _, v := range data {
		if jR(v, cat) {

			err = append(err, true)
		} else {
			err = append(err, false)
		}
	}

	return
}
func jR(data float64, cat string) (b bool) {
	switch cat {
	case "wv":
		b = data < 0 || data > 40
	case "wd":
		b = data < 0 || data > 360
	}
	return
}
func getErrT(db DB, cat string, s []Sensor) (errs [][]bool) {

	for _, v := range s {
		ch := v.Channel
		t := db.get("ChAvg" + ch + " time")
		data := t["ChAvg"+ch]
		time := t["time"]

		err := jTs(data, time, cat)

		errs = append(errs, err)
	}

	return
}
func jTs(data, time []float64, cat string) (err []bool) {
	for i := 0; i < len(data)-1; i++ {
		if math.Abs(time[i]-time[i+1]) > 3600000.0 {
			err = append(err, false)
			continue
		}
		if jT(data[i], data[i+1], cat) {

			err = append(err, true)
		} else {
			err = append(err, false)
		}
	}

	return
}
func jT(data1, data2 float64, cat string) (b bool) {
	switch cat {
	case "wv":
		b = math.Abs(data1-data2) >= 6
	case "t":
		b = math.Abs(data1-data2) >= 5
	case "p":
		b = math.Abs(data1-data2) >= 1
	}
	return
}
func calErrNumRT(errs [][]bool) (err []bool) {
	if len(errs) < 1 {
		return
	}

	err = errs[0]

	for i := 1; i < len(errs); i++ {
		err = arrayOr(err, errs[i])
	}

	return
}
func arrayOr(x, y []bool) (r []bool) {
	if len(x) < 1 && len(y) > 1 {
		r = y
		return
	}
	if len(x) > 1 && len(y) < 1 {
		r = x
		return
	}
	if len(x) < 1 && len(y) < 1 {
		return
	}
	for i, _ := range x {
		r = append(r, x[i] || y[i])
	}
	return
}
func trueNum(err []bool) (num int) {
	for _, v := range err {
		if v {
			num = num + 1
		}
	}
	return
}
func getErrC(db DB, cat string, s []Sensor) (errs [][][]bool) {

	if len(s) < 2 {
		return
	}

	var data [][]float64
	for _, v := range s {
		ch := v.Channel
		data = append(data, db.get("ChAvg" + ch)["ChAvg"+ch])
	}

	for i, v1 := range s {
		heightI := float64(v1.Height)

		var errI [][]bool
		for j, v2 := range s {
			heightJ := float64(v2.Height)

			if i == j {
				var err []bool
				errI = append(errI, err)
				continue
			}

			err := jCs(data[i], data[j], heightI, heightJ, cat)
			errI = append(errI, err)
		}
		errs = append(errs, errI)
	}

	return
}
func jCs(data1, data2 []float64, height1, height2 float64, cat string) (err []bool) {
	for i, _ := range data1 {
		if jC(data1[i], data2[i], height1, height2, cat) {
			err = append(err, true)
		} else {
			err = append(err, false)
		}
	}

	return
}
func jC(data1, data2, height1, height2 float64, cat string) (b bool) {
	switch cat {
	case "wv":
		n1 := math.Abs(height1-height2) / 10
		if n1 < 1.0 {
			n1 = 1.0
		}
		b = math.Abs(data1-data2) >= n1
	case "wd":
		n2 := math.Abs(data1 - data2)
		if n2 > 180.0 {
			n2 = 360.0 - n2
		}
		n1 := float64(math.Abs(height1-height2) / 20)
		if n1 < 1.0 {
			n1 = 1.0
		}
		b = n2 >= 22.5*n1 //
	}

	return
}
func calErrNumC(errs [][][]bool) (err []bool) {
	if len(errs) < 1 {
		return
	}
	var errt [][]bool

	for _, v := range errs {
		errt = append(errt, calErrNumRT(v))
	}

	err = calErrNumRT(errt)
	return
}
func countDays(year, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30
		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}

	return
}

func chooseOneYear(am []Am) (cm []Am, err error) {
	if len(am) < 12 {
		err = errors.New("chooseOneYear: len am < 12")
		return
	}

	choise := map[int]float64{}
	for i := 0; i < len(am)-11; i++ {
		for j := i; j < i+12; j++ {
			choise[i] = choise[i] + am[j].Sr
		}
	}

	min, _ := minFloatS(choise)

	for i := min; i < min+12; i++ {
		cm = append(cm, am[i])
	}

	return
}
func minFloatS(v map[int]float64) (index int, m float64) {
	if len(v) > 0 {
		m = v[0]
		index = 0
	}
	for i := 1; i < len(v); i++ {
		if v[i] < m {
			m = v[i]
			index = i
		}
	}
	return
}

func arrayOrT(x, y []bool) []bool {

	for i, _ := range x {
		y[i] = y[i] || x[i]
		y[i+1] = y[i+1] || x[i]

	}
	return y
}
