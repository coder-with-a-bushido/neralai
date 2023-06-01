package hls

import "os"

func CleanUp() {
	// delete `output` directory
	if err := os.RemoveAll(OutputDir); err != nil {
		panic(err)
	}
}
