package wind

import (
	"math"
)

const (
	M = 0.577215664901532860606512090082402431042159335
)

func WeibullKC(wv []float64) (k, c float64) {
	avg := ArrayAvg(wv)
	sv := ArraySv(wv)

	k = math.Pow(sv/avg, -1.086)
	c = avg / gamma(1.0+1.0/k)

	return
}

func Weibull(k, c, v float64) float64 {
	return (k / c) * math.Pow(v/c, k-1) * math.Exp(-math.Pow(v/c, k))
}

func gamma(z float64) float64 {
	y := math.Exp(-1*M*z) / z
	n := 1.0
	for true {
		yt := y * math.Pow(1.0+z/n, -1) * math.Exp(z/n)

		if math.Abs(yt-y) < 1e-15 {
			return yt
		}
		y = yt
		n = n + 1.0

	}

	return 0.0
}
