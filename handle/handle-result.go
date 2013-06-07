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

	if config.Result == "dir" {
		dirs, _ := ioutil.ReadDir(OUTPUT_DIR)

		for _, v := range dirs {
			if v.IsDir() {
				result := Result{
					Id:   v.Name(),
					Date: v.ModTime(),
				}

				results = append(results, result)
			}
		}
	} else {
		for _, v := range datas {
			result := Result{
				Id:   v.Id,
				Date: v.T,
			}

			results = append(results, result)
		}
	}

	page := Page{
		"results":        results,
		"hideResultMenu": true,
	}

	page.render("result", w)
}
