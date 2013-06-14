package wind

import (
	"errors"
	"strconv"
	"time"
)

// 计算平均风速风功率

func CalWvpAvgH(T, D []float64) (data []float64, err error) {
	var db DB
	a := map[string][]float64{
		"D": D,
	}
	db, err = NewDB(T, a)
	if err != nil {
		return
	}

	for i := 0; i < 24; i++ {
		d := db.Filter("Hour =", float64(i)).get("D")["D"]

		if len(d) < 1 {
			data = append(data, 0.0)
		} else {
			data = append(data, ArrayAvg(d))
		}
	}

	return
}

func CalWvpAvgM(T, D []float64, limit bool) (data []float64, A []string, err error) {
	var db DB
	a := map[string][]float64{
		"D": D,
	}
	db, err = NewDB(T, a)
	if err != nil {
		return
	}

	addWvp := func(d []float64, s string) {
		if len(d) < 1 {
			data = append(data, 0.0)
		} else {
			data = append(data, ArrayAvg(d))
		}

		A = append(A, s)
	}

	if limit {
		for i := 1; i < 13; i++ {
			d := db.Filter("Month =", float64(i)).get("D")["D"]
			addWvp(d, strconv.Itoa(i))
		}
	} else {
		my := getMy(T)

		for _, v := range my {
			d := db.Filter("My =", v).get("D")["D"]
			addWvp(d, strconv.Itoa(int(v)))
		}
	}

	return
}

func CalWvpAvgMH(T, D []float64, limit bool) (data [][]float64, A []string, err error) {
	var db DB
	a := map[string][]float64{
		"D": D,
	}
	db, err = NewDB(T, a)
	if err != nil {
		return
	}

	addWvp := func(dbM DB, s string) {
		var ds []float64
		for j := 0; j < 24; j++ {
			d := dbM.Filter("Hour =", float64(j)).get("D")["D"]
			if len(d) < 1 {
				ds = append(ds, 0.0)
			} else {
				ds = append(ds, ArrayAvg(d))
			}
		}

		data = append(data, ds)
		A = append(A, s)
	}

	if limit {
		for i := 1; i < 13; i++ {
			dbM := db.Filter("Month =", float64(i))

			addWvp(dbM, strconv.Itoa(i))
		}
	} else {
		my := getMy(T)

		for _, v := range my {
			dbM := db.Filter("My =", v)
			addWvp(dbM, strconv.Itoa(int(v)))
		}
	}

	return
}

func getMy(T []float64) (mys []float64) {
	min := int64(ArrayMin(T))
	max := int64(ArrayMax(T))

	location, _ := time.LoadLocation("Local")
	for t := time.Unix(min, 0); t.Unix() <= max; {
		mys = append(mys, float64(t.Year()*100+int(t.Month())))
		t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, location)
	}

	return
}

func existMy(my []float64, f float64) bool {
	for _, v := range my {
		if f == v {
			return true
		}
	}

	return false
}

func NewDB(T []float64, a map[string][]float64) (db DB, err error) {
	for _, v := range a {
		if len(v) != len(T) {
			err = errors.New("NewDB: len not equal!")
			return
		}
	}

	for i, v := range T {
		t := time.Unix(int64(v), 0)
		var f float64
		f, err = strconv.ParseFloat(t.Format("200601"), 64)
		if err != nil {
			return
		}

		tData := Data{
			"Time":  v,
			"My":    f,
			"Hour":  float64(t.Hour()),
			"Day":   float64(t.Day()),
			"Year":  float64(t.Year()),
			"Month": float64(t.Month()),
		}

		for k, v2 := range a {
			tData[k] = v2[i]
		}

		db = append(db, tData)
	}

	return
}
