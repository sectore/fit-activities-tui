package fit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsFitFile(file os.FileInfo) bool {
	return file.Mode().IsRegular() &&
		strings.HasSuffix(file.Name(), ".fit")
}

func GetFitFiles(path string) ([]string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}
	var fitFiles []string
	if fileInfo.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if fileInfo, err := file.Info(); err == nil && !file.IsDir() && IsFitFile(fileInfo) {
				fitFiles = append(fitFiles, filepath.Join(path, file.Name()))
			}
		}
	} else {
		if IsFitFile(fileInfo) {
			fitFiles = append(fitFiles, path)
		}
	}

	return fitFiles, nil
}
