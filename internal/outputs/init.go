package outputs

import (
	"fmt"
	"os"
)

var outputDir string

func Init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	outputDir = fmt.Sprintf("%s/output", dir)

	// check and create `output` dir if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err = os.Mkdir(outputDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
}
