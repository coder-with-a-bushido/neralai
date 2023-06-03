package utils

import (
	"fmt"
	"os"
)

var outputDir string

// Return `output` directory path.
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

// Check and create a directory if it doesn't exist.
func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Delete all contents of a directory.
func DeleteDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

// Create a new file and return it.
func NewFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Create new file and write the `content` to it.
func CreateAndWriteToFile(path, content string) error {
	file, err := NewFile(path)
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
