package handle

import (
	"text/template"
	"encoding/json"
	"fmt"
)

var funcMap = template.FuncMap{
	"equal": equal,
	"addInt": addInt,
	"toJson": toJson,
}

func equal(a, b interface{}) bool {
	if a==b {
		return true
	}
	return false
}
func addInt(a ...int) (r int) {
	for _, v := range a {
		r = r + v
	}
	return
}

func toJson(a interface{}) string {
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println("error:", err)
	}

	return string(b)
}

