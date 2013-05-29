package handle

import (
	"github.com/gorilla/mux"
	"math"
	"net/http"

	"github.com/woodycarl/wind-go/wind"
)

type WindshearData struct {
	Points [][]float64
	Line   [][]float64
	A      float64
	B      float64
	R      float64
	Data   []wind.Ws
}

func handleWindshear(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	data := getData(id)
	s := data.Station

	var points [][]float64
	for i, v := range s.Wss.Height {
		point := []float64{math.Exp(v), math.Exp(s.Wss.Wv[i])}
		points = append(points, point)
	}

	var line [][]float64
	for i := 5; i < 90; {
		x := float64(i)
		point := []float64{x, s.Wss.A * math.Pow(x, s.Wss.B)}
		line = append(line, point)
		i = i + 5
	}

	windshear := WindshearData{
		A:      s.Wss.A,
		B:      s.Wss.B,
		R:      s.Wss.R,
		Points: points,
		Line:   line,
		Data:   s.Wss.Ws,
	}

	page := Page{
		"id":        id,
		"windshear": windshear,
	}

	page.render("windshear", w)
}
