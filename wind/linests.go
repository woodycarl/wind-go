package wind

import (
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

type ChLinestData struct {
	ir      int
	v       string
	is      int
	rations []Ration
}

func linests(r []Result) []Result {
	Info("---Linests---")
	timeS := time.Now()

	cats := []string{"wv", "wd", "p", "t", "h"}
	chLinest := make(chan ChLinestData, 4)
	for _, v := range cats {
		for ir, v1 := range r {
			for is, _ := range v1.S.Sensors[v] {
				go linest(v, ir, is, r, chLinest)
			}
		}
	}

	for _, v := range cats {
		for _, v1 := range r {
			for _, _ = range v1.S.Sensors[v] {
				tD := <-chLinest
				r[tD.ir].S.Sensors[tD.v][tD.is].Rations = tD.rations
			}
		}
	}

	Info("Linests", time.Now().Sub(timeS))
	return r
}

func linest(v string, ir, is int, r []Result, ch chan ChLinestData) {
	var rations []Ration
	v1 := r[ir]
	v2 := v1.S.Sensors[v][is]

	for jr, v3 := range r {
		for js, v4 := range v3.S.Sensors[v] {
			if ir == jr && is == js {
				continue
			}

			chI, chJ := v2.Channel, v4.Channel
			var dataI, dataJ []float64
			if ir == jr {
				dataI = DB(v1.D1).get("ChAvg" + chI)["ChAvg"+chI]
				dataJ = DB(v3.D1).get("ChAvg" + chJ)["ChAvg"+chJ]
			} else if v2.Height == v4.Height {
				dI := DB(v1.D1).get("Time ChAvg" + chI)
				dJ := DB(v3.D1).get("Time ChAvg" + chJ)
				dataI, dataJ = getUnite(dI["Time"], dJ["Time"], dI["ChAvg"+chI], dJ["ChAvg"+chJ])

				if len(dataI) < 2000 {
					Info("not enough data", v1.ID, v3.ID, chI, chJ, v2.Height, v4.Height, len(dataI))
					continue
				}
			} else {
				continue
			}

			if v == "wd" {
				dataI, dataJ = adjustWd(dataI, dataJ)
			}

			slope, intercept, rsq := CalLinestRsq(dataJ, dataI)
			r2 := rsq * rsq
			if r2 < 0.8 {
				Warn(v1.ID, v3.ID, chI, chJ, v2.Height, v4.Height, r2, "Abandoned")
				continue
			}

			Info(v1.ID, v3.ID, v2.Height, v4.Height, r2)

			ration := Ration{
				Index:     js,
				ID:        v3.ID,
				Channel:   chJ,
				Rsq:       r2,
				Slope:     slope,
				Intercept: intercept,
			}

			rations = append(rations, ration)

		}
	}

	rations = rationsArrange(rations)
	chLinestData := ChLinestData{
		ir:      ir,
		is:      is,
		v:       v,
		rations: rations,
	}

	ch <- chLinestData
}

func rationsArrange(r []Ration) []Ration {
	for i := 0; i < len(r)-1; i++ {
		for j := len(r) - 1; j > i; j-- {
			if r[j].Rsq > r[j-1].Rsq {
				t := r[j-1]
				r[j-1] = r[j]
				r[j] = t
			}
		}
	}

	return r
}

func adjustWd(dataI, dataJ []float64) ([]float64, []float64) {
	for i := 0; i < len(dataI); i++ {
		if (dataI[i] - dataJ[i]) > 180.0 {
			dataI[i] = dataI[i] - 360.0
		} else if (dataJ[i] - dataI[i]) > 180.0 {
			dataJ[i] = dataJ[i] - 360.0
		}
	}

	return dataI, dataJ
}

func getUnite(tI, tJ, dI, dJ []float64) (dataI, dataJ []float64) {
	i, j := 0, 0
	for i < len(tI) && j < len(tJ) {
		switch {
		case tI[i] > tJ[j]:
			j++
		case tI[i] < tJ[j]:
			i++
		case tI[i] == tJ[j]:
			dataI = append(dataI, dI[i])
			dataJ = append(dataJ, dJ[j])
			i++
			j++
		}
	}

	return
}
