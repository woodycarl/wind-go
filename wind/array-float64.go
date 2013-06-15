package wind

import (
	"math"
)

func ArrayAvg(array []float64) float64 {
	return ArraySum(array) / float64(len(array))
}

func ArraySum(array []float64) float64 {
	sum := 0.0
	for _, v := range array {
		sum = sum + v
	}

	return sum
}

func ArrayAdd(array []float64, n float64) (r []float64) {
	for _, v := range array {
		r = append(r, v+n)
	}

	return r
}

func ArrayTime(arrayX, arrayY []float64) (r []float64) {
	for i, _ := range arrayX {
		r = append(r, arrayX[i]*arrayY[i])
	}

	return r
}

func ArrayPow(array []float64, pow float64) (r []float64) {
	for _, v := range array {
		r = append(r, math.Pow(v, pow))
	}

	return r
}

func ArraySv(array []float64) float64 {
	avg := ArrayAvg(array)

	var a []float64

	for _, v := range array {
		a = append(a, math.Pow(v-avg, 2))
	}

	return math.Sqrt(ArraySum(a) / float64(len(a)))
}

func ArrayTimeN(array []float64, n float64) (r []float64) {
	for _, v := range array {
		r = append(r, v*n)
	}
	return
}

func ArrayMod(array []float64, m float64) (r []float64) {
	for _, v := range array {
		r = append(r, math.Mod(v, m))
	}
	return
}

func ArrayMax(array []float64) float64 {
	max := array[0]
	for _, v := range array {
		if max < v {
			max = v
		}
	}
	return max
}

func ArrayMin(array []float64) float64 {
	min := array[0]
	for _, v := range array {
		if min > v {
			min = v
		}
	}
	return min
}

func ArrayMinI(array []float64) (index int, min float64) {
	if len(array) < 1 {
		index = -1
		return
	}

	min = array[0]
	for i, v := range array {
		if min > v {
			index = i
			min = v
		}
	}
	return
}

func Linspace(x, y float64, n int) (r []float64) {
	for i, s := 0, (y-x)/float64(n); i < n+1; i++ {
		r = append(r, x+s*float64(i))
	}
	return
}

func CalLinestRsq(arrayX, arrayY []float64) (s, i, r float64) {
	avgX := ArrayAvg(arrayX)
	avgY := ArrayAvg(arrayY)

	data1 := ArrayAdd(arrayX, -avgX)
	data2 := ArrayAdd(arrayY, -avgY)
	data3 := ArraySum(ArrayTime(data1, data2))
	data4 := ArraySum(ArrayPow(data1, 2.0))
	data5 := ArraySum(ArrayPow(data2, 2.0))

	s = data3 / data4
	i = avgY - s*avgX

	r = data3 / math.Sqrt(data4*data5)
	return
}
