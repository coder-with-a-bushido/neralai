package hls

import (
	"fmt"
	"os"
)

var OutputDir string

func Init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	OutputDir = fmt.Sprintf("%s/output", dir)

	// check and create `output` dir if it doesn't exist
	if _, err := os.Stat(OutputDir); os.IsNotExist(err) {
		if err = os.Mkdir(OutputDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
}
