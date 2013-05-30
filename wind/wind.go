package wind

import (
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

func HandleData(data [][]string, c Config) (r []Result, err error) {
	Info("HandleData")

	r, err = decRaw(data, c)
	if err != nil {
		return
	}

	timeS := time.Now()
	for i, _ := range r {
		r[i].S.AirDensity = c.AirDensity
		r[i].S.Sensors = sensorClassify(r[i].S.SensorsR)
		r[i].S.Am, err = getAm(r[i].D1)
		if err != nil {
			return
		}
	}
	Info("sensorClassify + getAm", time.Now().Sub(timeS))

	r = linests(r)        // 计算线性相关
	r = integrities(r, c) // 计算完整率
	r = revises(r, c)     // 修订数据
	r = caculates(r, c)   // 计算需要的数据

	return
}
