package handle

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/woodycarl/wind-go/wind"
)

const (
	CONFIG_FILE_PATH      = "config.json"
	CONFIG_MAX_NUM_IN_MEM = 5
)

var (
	config Config
)

type Config struct {
	Port        string
	Result      string // dir|mem 结果页 显示输出目录中的结果|显示内存中的结果
	MaxNumInMem int
	Config      wind.Config
}

func init() {
	config = getJsonConfig()
}

func getJsonConfig() (config Config) {
	configFile, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	if config.MaxNumInMem == 0 {
		config.MaxNumInMem = CONFIG_MAX_NUM_IN_MEM
	}

	re := regexp.MustCompile(`[dD][iI][rR]`)
	if re.MatchString(config.Result) {
		config.Result = "dir"
	} else {
		config.Result = "mem"
	}

	return
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("data_revise") {
	case "false":
		config.Config.AutoRevise = false
		Info("Config: AutoRevise false")
	case "true":
		config.Config.AutoRevise = true
		Info("Config: AutoRevise true")
	}

	switch r.FormValue("data_result") {
	case "dir":
		config.Result = "dir"
		Info("Config: Result dir")
	case "mem":
		config.Result = "mem"
		Info("Config: Result mem")
	}

	switch r.FormValue("data_separate") {
	case "false":
		config.Config.Separate = false
		Info("Config: Separate false")
	case "true":
		config.Config.Separate = true
		Info("Config: Separate true")
	}

	if maxNumS := r.FormValue("data_max_num"); maxNumS != "" {
		max, err := strconv.Atoi(maxNumS)
		if err == nil && max > 0 {
			config.MaxNumInMem = max
			Info("Config: MaxNumInMem", max)
		}
	}
}
