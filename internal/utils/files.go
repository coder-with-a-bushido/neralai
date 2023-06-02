package utils

import (
	"fmt"
	"os"
)

var outputDir string

func GetOutputDir() string {
	if outputDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		outputDir = fmt.Sprintf("%s/output", dir)

	}
	return outputDir
}

func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func DeleteDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func CreateAndWriteToFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	return nil
}

func NewLogFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
