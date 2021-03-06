package handle

import (
	"net/http"
	"runtime"

	"github.com/gorilla/mux"
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
	r.HandleFunc("/config", handleConfig)

	r.HandleFunc("/result/{id}/{cat:info-site|info-logger|info-sensors}", handleInfo)
	r.HandleFunc("/result/{id}/{cat:integrity|integrity-all}", handleInfo)

	r.HandleFunc("/result/{id}/linest/{cat:wv|wd}", handleLinest)
	r.HandleFunc("/result/{id}/linest-fig/{cat:wv|wd}/{ch}", handleLinestFig)

	r.HandleFunc("/result/{id}/wvp-avg/{cat:all|turb|raw|raw-a}", handleWvpAvg)
	r.HandleFunc("/result/{id}/turbs", handleTurbs)
	r.HandleFunc("/result/{id}/windshear", handleWindshear)

	r.HandleFunc("/result/{id}/wvp-freq", handleWvpFreq)
	r.HandleFunc("/result/{id}/wvp-freq/turb", handleWvpFreqTurbine)
	r.HandleFunc("/result/{id}/windrose", handleWdvpWindRose)
	r.HandleFunc("/result/{id}/weibull", handleWeibull)

	r.HandleFunc("/result/{id}/contrast", handleContrast)

	http.Handle("/", r)

	http.ListenAndServe(":"+config.Port, nil)
}
