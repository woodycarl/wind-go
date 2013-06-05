package wind

import (
	"strings"
)

type Data map[string]float64

type DB []Data

//	Filter|filter  对数据行进行过滤
//	str 过滤规则   列名 > | < |= | >= | <=, 默认 =
//	val 规则对应值
func (db DB) Filter(str string, val float64) DB {
	return db.filter(str, val)
}

func (db DB) filter(str string, val float64) DB {
	col, method := decodeFilterStr(str)

	var data DB
	for _, v := range db {
		if judgeColVal(v[col], method, val) {
			data = append(data, v)
		}
	}
	return data
}

// Get|get: 提取数据列
// str 以空格分隔的列名
func (db DB) Get(str string) map[string][]float64 {
	return db.get(str)
}
func (db DB) get(str string) map[string][]float64 {
	r := strings.Fields(str)
	var data = make(map[string][]float64)
	for _, v1 := range r {
		for _, v2 := range db {
			data[v1] = append(data[v1], v2[v1])
		}
	}

	return data
}

//	对数据是否符合规则进行判断
func judgeColVal(d float64, method string, val float64) bool {
	switch {
	case method == "=" && d == val:
		return true
	case method == ">" && d > val:
		return true
	case method == "<" && d < val:
		return true
	case method == ">=" && d >= val:
		return true
	case method == "<=" && d <= val:
		return true
	}

	return false
}

//	解析过滤规则字符串
func decodeFilterStr(str string) (string, string) {
	r := strings.Fields(str)
	if len(r) > 1 {
		return r[0], r[1]
	}
	return r[0], "="
}
