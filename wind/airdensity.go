package wind

import (
	"math"
)

/*
	计算空气密度
*/

//	计算空气密度方法1：当有气压和气温数据时
//	dataP 压力数据
//	dataT 温度数据
func AirDensity(dataP, dataT []float64) float64 {
	P := ArrayAvg(dataP) * 1000.0
	T := ArrayAvg(dataT) + 273.0

	return P / (287.0 * T)
}

//	计算空气密度方法2：当没有气压数据时
//	dataT 温度数据
//	height 海拔高度
func AirDensity2(dataT []float64, height float64) float64 {
	T := ArrayAvg(dataT) + 273.0

	return (353.05 / T) * math.Exp((-0.034)*(height/T))
}
