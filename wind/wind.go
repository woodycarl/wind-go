package wind

import (
	. "github.com/woodycarl/wind-go/logger"
)

func HandleData(data [][]string, c Config) (r []Result, err error) {
	Info("HandleData")

	r, err = decRaw(data)
	if err != nil {
		return
	}

	for i, _ := range r {
		r[i].S.AirDensity = c.AirDensity
	}

	r = linests(r)             // 计算线性相关
	r, err = integrities(r, c) // 计算完整率
	if err != nil {
		return
	}

	r, err = revises(r, c) // 修订数据
	if err != nil {
		return
	}

	r = caculates(r, c) // 计算需要的数据

	return
}
