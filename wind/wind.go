package wind

func HandleData(data [][]string, c Config) (r []Result, err error) {
	r, err = decRaw(data)
	if err != nil {
		return
	}

	for i, _ := range r {
		r[i].S.AirDensity = c.AirDensity
		r[i].S.TurbineHeight = c.CalHeight
	}

	r = linests(r, c)     // 计算线性相关
	r = integrities(r, c) // 计算完整率

	r, err = revises(r, c) // 修订数据
	if err != nil {
		return
	}

	r = caculates(r, c) // 计算需要的数据

	return
}
