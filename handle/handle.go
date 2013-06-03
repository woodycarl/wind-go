package handle

import (
	"net/http"
	"runtime"

	"github.com/gorilla/mux"
	. "github.com/woodycarl/wind-go/logger"
)

func init() {
	nCPUs := 4
	runtime.GOMAXPROCS(nCPUs)
}

func Main() {
	Info("System Start...")

	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))
	http.Handle("/img/", http.FileServer(http.Dir("template")))

	r := mux.NewRouter()

	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/data", handleData)
	r.HandleFunc("/result", handleResult)

	r.HandleFunc("/result/{id}/{cat:info-site|info-logger|info-sensors}", handleInfo)
	r.HandleFunc("/result/{id}/{cat:integrity|integrity-all}", handleInfo)

	r.HandleFunc("/result/{id}/linest/{cat:wv|wd}", handleLinest)
	r.HandleFunc("/result/{id}/linest-fig/{cat:wv|wd}/{ch}", handleLinestFig)

	r.HandleFunc("/result/{id}/{cat1:wv|wp}/{cat2:yh|ym}", handleWvpFig)
	r.HandleFunc("/result/{id}/wvp/mh", handleWvpMhFig)

	r.HandleFunc("/result/{id}/turbs", handleTurbs)
	r.HandleFunc("/result/{id}/windshear", handleWindshear)
	r.HandleFunc("/result/{id}/windrose", handleWdvpWindRose)
	r.HandleFunc("/result/{id}/wvp-freq", handleWvpFreq)
	//r.HandleFunc("/result/{id}/turbine/wvp", handleTurbineWvp)

	r.HandleFunc("/result/{id}/contrast", handleContrast)

	http.Handle("/", r)

	http.ListenAndServe(":"+config.Port, nil)
}
