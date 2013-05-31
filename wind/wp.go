package wind

import (
	"math"
)

func Wv2Wp(V []float64, AirDensity float64) (P []float64) {
	for _, v := range V {
		P = append(P, AirDensity*math.Pow(v, 3.0))
	}
	return
}
