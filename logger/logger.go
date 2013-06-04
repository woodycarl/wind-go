package logger

import (
	"log"
	"os"
)

// https://github.com/astaxie/beego/blob/master/log.go

// logger references the used application logger.
var Log = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// SetLogger sets a new logger.
func SetLogger(l *log.Logger) {
	Log = l
}

func Debug(v ...interface{}) {
	Log.Printf("[D] %v\n", v)
}

func Info(v ...interface{}) {
	Log.Printf("[I] %v\n", v)
}

func Warn(v ...interface{}) {
	Log.Printf("[W] %v\n", v)
}

func Error(v ...interface{}) {
	Log.Printf("[E] %v\n", v)
}
