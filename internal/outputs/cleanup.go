package outputs

import "os"

func CleanUp() {
	// delete `output` directory
	if err := os.RemoveAll(outputDir); err != nil {
		panic(err)
	}
}
