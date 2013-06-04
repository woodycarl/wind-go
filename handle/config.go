package handle

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/woodycarl/wind-go/wind"
)

const (
	CONFIG_FILE_PATH = "config.json"
)

var (
	config Config
)

type Config struct {
	Port   string
	Config wind.Config
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

	return
}
