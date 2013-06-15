package wind

import (
	"time"

	"github.com/woodycarl/wind-go/logger"
)

const (
	DATE_FORMAT_MY   = "200601"
	DATE_FORMAT_YMHM = "2006010215"
)

var (
	Info  = logger.Info
	Warn  = logger.Warn
	Debug = logger.Debug
	Error = logger.Error

	LOCATION, _ = time.LoadLocation("Local")
)
