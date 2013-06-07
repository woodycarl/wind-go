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

func handleConfig(r *http.Request) {
	if r.FormValue("data_revise") == "false" {
		config.Config.AutoRevise = false
	} else {
		config.Config.AutoRevise = true
	}

	if r.FormValue("data_result") == "dir" {
		config.Result = "dir"
	} else {
		config.Result = "mem"
	}

	max, err := strconv.Atoi(r.FormValue("data_max_num"))
	if err != nil {
		log.Fatal(err)
	}
	if max > 0 {
		config.MaxNumInMem = max
	}
}
