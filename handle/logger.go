package handle

import (
	"time"
	"fmt"
)

func logger(a ...interface{}) {
	fmt.Println(time.Now().Format("15:04:05 ")+fmt.Sprint(a))
}
