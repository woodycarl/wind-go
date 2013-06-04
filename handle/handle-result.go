package handle

import (
	"io/ioutil"
	"net/http"
	"time"
)

func handleResult(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		Id   string
		Date time.Time
	}
	var results []Result

	ds, _ := ioutil.ReadDir(OUTPUT_DIR)

	for _, v := range ds {
		if v.IsDir() {
			result := Result{
				Id:   v.Name(),
				Date: v.ModTime(),
			}

			results = append(results, result)
		}
	}

	page := Page{
		"results": results,
	}

	page.render("result", w)
}
