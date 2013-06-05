package wind

import (
	"math"
	"strconv"
)

type WvpArg struct {
	Interval int
}

func WvpWindrose(V, P []float64, arg WvpArg) (Ay []float64, wvF, wpF []float64) {
	wvsums, wpsums := map[string]int{}, map[string]float64{}
	for i, v := range V {
		index := strconv.Itoa(getWvIndex(v, arg.Interval))
		wvsums[index] = wvsums[index] + 1
		wpsums[index] = wpsums[index] + P[i]
	}

	step := math.Ceil(ArrayMax(V) / float64(arg.Interval))
	Ay = []float64{0.5}
	for i := 1; float64(i) <= step; i++ {
		Ay = append(Ay, float64(i*arg.Interval))
	}

	allNum := len(V)

	allWp := 0.0
	for _, v := range wpsums {
		allWp = allWp + v
	}

	for i := 0; i < len(Ay); i++ {
		index := strconv.Itoa(i)
		wvF = append(wvF, 100*float64(wvsums[index])/float64(allNum))
		wpF = append(wpF, 100*wpsums[index]/allWp)
	}

	return
}

func getWvIndex(v float64, interval int) int {
	if v < 0.5 {
		return 0
	}
	return int(v/float64(interval)) + 1
}
