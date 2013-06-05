package wind

import (
	"strconv"
)

func WvAvg(V []float64) (wvF map[string]float64) {
	wvsums := map[string]int{}
	for _, v := range V {
		index := strconv.Itoa(int(v) + 1)
		wvsums[index] = wvsums[index] + 1
	}

	allNum := len(V)
	wvF = map[string]float64{}

	for k, v := range wvsums {
		wvF[k] = 100.0 * float64(v) / float64(allNum)
	}

	return
}
