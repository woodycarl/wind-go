package handle

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/woodycarl/wind-go/wind"
)

var (
	datas []Data
)

const (
	OUTPUT_DIR         = "./output/"
	WVPT_DATE_FORMAT   = "2006010215"
	WVPT_FILE_NAME     = "data-wvd.txt"
	RD_FILE_NAME       = "data-r.txt"
	CONTRAST_FILE_NAME = "data-c.txt"
	DATA_DATE_FORMAT   = "2006/01/02 15:04:05"
	DATA_FILE_NAME     = "data.txt"
)

type Data struct {
	Id string
	T  time.Time
	ID string
	S  wind.Station
	D1 []wind.Data
	D2 []wind.Data
	RD []wind.Data
}

func getData(id string) (data Data, err error) {
	timeS := time.Now()
	for _, v := range datas {
		if v.Id == id {
			data = v
			return
		}
	}

	data, err = getData2(id)
	Info("getData", time.Now().Sub(timeS))
	return
}

func getData2(id string) (data Data, err error) {
	var datafile []byte

	datafile, err = ioutil.ReadFile(OUTPUT_DIR + id + "/" + DATA_FILE_NAME)
	if err != nil {
		return
	}

	err = json.Unmarshal(datafile, &data)
	if err != nil {
		return
	}

	addData(data)
	return
}

func addData(data Data) {
	if len(datas) > config.MaxNumInMem {
		datas = datas[1:]
	}

	datas = append(datas, data)
}
