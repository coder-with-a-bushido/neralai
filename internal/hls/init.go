package hls

import (
	"coder-with-a-bushido.in/neralai/internal/utils"
)

func Init() {
	// Create `output` dir
	if err := utils.CreateDir(utils.GetOutputDir()); err != nil {
		panic(err)
	}
}

func CleanUp() {
	// Delete `output` directory
	if err := utils.DeleteDir(utils.GetOutputDir()); err != nil {
		panic(err)
	}
}
