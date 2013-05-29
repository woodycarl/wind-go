package handle

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var genID = generateID()

func generateID() func() string {
	num := 0
	ds, _ := ioutil.ReadDir(OUTPUT_DIR)

	for _, v := range ds {
		if v.IsDir() {
			xs := strings.Split(v.Name(), "-")
			x, _ := strconv.Atoi(xs[0])
			if x > num {
				num = x
			}
		}
	}

	return func() string {
		num++
		return strconv.Itoa(num)
	}
}

func writeLines(lines []string, path string) (err error) {
	var (
		file *os.File
	)

	if file, err = os.Create(path); err != nil {
		return
	}
	defer file.Close()

	for _, item := range lines {
		_, err := file.WriteString(strings.TrimSpace(item) + "\r\n")
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	return
}
