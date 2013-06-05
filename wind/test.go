package wind

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/woodycarl/wind-go/logger"
)

var g = getId()

func writeLines(lines []string, path string) (err error) {
	var (
		file *os.File
	)

	if file, err = os.Create(path); err != nil {
		return
	}
	defer file.Close()

	for _, item := range lines {
		_, err = file.WriteString(strings.TrimSpace(item) + "\r\n")
		if err != nil {
			return
		}
	}

	return
}

func saveRData(id string, data []Data, s []Sensor) {
	timeS := time.Now()

	var lines []string

	for _, v1 := range data {
		t := time.Unix(int64(v1["Time"]), 0)
		line := ""
		line = line + t.Format("2006/01/02 15:04:05")

		for _, v2 := range s {
			ch := v2.Channel

			line = line + "\t" + fmt.Sprint(v1["ChAvg"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChSd"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChMin"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChMax"+ch])
		}

		lines = append(lines, line)
	}

	err := writeLines(lines, "./output/data-"+id+".txt")
	if err != nil {
		Error("saveRData", err)
		return
	}

	Info("saveRData", time.Now().Sub(timeS))
}

func getId() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}
