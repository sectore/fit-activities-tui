package fit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsFitFile(fileName string) bool {
	return !strings.HasPrefix(fileName, ".") &&
		strings.HasSuffix(fileName, ".fit")
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
			if !file.IsDir() && IsFitFile(file.Name()) {
				fitFiles = append(fitFiles, filepath.Join(path, file.Name()))
			}
		}
	} else {
		fitFiles = append(fitFiles, path)
	}

	return fitFiles, nil
}
