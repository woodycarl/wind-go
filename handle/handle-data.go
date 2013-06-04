package handle

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	. "github.com/woodycarl/wind-go/logger"
	"github.com/woodycarl/wind-go/wind"
)

func handleData(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
	files := r.MultipartForm.File["files"]

	if len(files) < 1 {
		err := errors.New("no file upload!")
		handleErr(w, err)
		return
	}

	var rawData [][]string
	for _, fh := range files {
		file, err := fh.Open()
		defer file.Close()
		if err != nil {
			handleErr(w, err)
			return
		}

		lines, err := getFile(file)
		if err != nil {
			handleErr(w, err)
			return
		}

		rawData = append(rawData, lines)
	}

	res, err := wind.HandleData(rawData, config.Config)
	if err != nil {
		handleErr(w, err)
		return
	}

	for _, v := range res {
		id := genID() + "-" + v.S.Site.Site
		data := Data{
			Id:      id,
			ID:      v.ID,
			Station: v.S,
			Data1h:  v.D1,
			Data10m: v.D2,
			RData:   v.RD,
		}

		addData(data)

		os.Mkdir(OUTPUT_DIR+id, 0700)

		go saveData(data)
		go saveRData(id, v.RD, v.S)
		go saveWvptData(id, v.S.DataTime, v.S.DataWv, v.S.DataWd)
	}

	http.Redirect(w, r, "/result", http.StatusFound)
}

func saveWvptData(id string, ts, wv, wd []float64) {
	timeS := time.Now()

	if len(ts) != len(wv) || (len(ts)+len(wv))/2 != len(wd) || len(ts) < 8000 {
		Error("saveWvptData err!")
		return
	}

	var lines []string
	for i, v := range ts {
		t := time.Unix(int64(v), 0).Format("2006010215")

		line := t + "\t" + fmt.Sprint(wv[i]) + "\t" + fmt.Sprint(wd[i])
		lines = append(lines, line)
	}

	err := writeLines(lines, OUTPUT_DIR+id+"/data-wvd.txt")
	if err != nil {
		Error("saveWvptData", err)
		return
	}

	Info("saveWvptData", time.Now().Sub(timeS))
}

func saveRData(id string, data []wind.Data, s wind.Station) {
	timeS := time.Now()

	if len(data) < 8000 {
		Error("saveRData\tnot enough data!")
		return
	}

	var lines []string

	line0 := "Date & Time Stamp"
	for _, v := range s.SensorsR {
		ch := v.Channel
		line0 = line0 + "\t" + "Ch" + ch + "Avg"
		line0 = line0 + "\t" + "Ch" + ch + "SD"
		line0 = line0 + "\t" + "Ch" + ch + "Max"
		line0 = line0 + "\t" + "Ch" + ch + "Min"
	}
	lines = append(lines, line0)

	for _, v1 := range data {
		t := time.Unix(int64(v1["Time"]), 0)
		line := ""
		line = line + t.Format("2006/01/02 15:04:05")

		for _, v2 := range s.SensorsR {
			ch := v2.Channel
			line = line + "\t" + fmt.Sprint(v1["ChAvg"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChSd"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChMin"+ch])
			line = line + "\t" + fmt.Sprint(v1["ChMax"+ch])
		}

		lines = append(lines, line)
	}

	err := writeLines(lines, OUTPUT_DIR+id+"/data-r.txt")
	if err != nil {
		Error("saveRData", err)
		return
	}

	Info("saveRData", time.Now().Sub(timeS))
}

func saveData(data Data) {
	timeS := time.Now()

	b, err := json.Marshal(data)
	if err != nil {
		Error("saveData:", err)
		return
	}

	f, err := os.Create(OUTPUT_DIR + data.Id + "/data.txt")
	if err != nil {
		Error("saveData", err)
	}
	defer f.Close()
	f.Write(b)

	Info("save data", time.Now().Sub(timeS))
}

func getFile(file multipart.File) (lines []string, err error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	return
}
