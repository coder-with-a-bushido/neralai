package hls

import (
	"coder-with-a-bushido.in/neralai/internal/utils"
)

func Init() {
	// check and create `output` dir if it doesn't exist
	if err := utils.CreateDir(utils.GetOutputDir()); err != nil {
		panic(err)
	}
}

func CleanUp() {
	// delete `output` directory
	if err := utils.DeleteDir(utils.GetOutputDir()); err != nil {
		panic(err)
	}
}
