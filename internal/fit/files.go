package fit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isFitFile(file os.FileInfo) bool {
	return file.Mode().IsRegular() &&
		strings.HasSuffix(file.Name(), ".fit")
}

func isValidFitFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && isFitFile(info)
}

func filesOrError(files []string, path string) ([]string, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("Given path does not include FIT files: %s", path)
	}
	return files, nil
}

func GetFitFilePaths(path string) ([]string, error) {
	var fitFiles []string

	// 1. Try glob pattern first
	if strings.ContainsAny(path, "*?[") {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %v", err)
		}
		for _, p := range matches {
			if isValidFitFile(p) {
				fitFiles = append(fitFiles, p)
			}
		}
		return filesOrError(fitFiles, path)
	}

	// 2. Try as directory
	if fileInfo, err := os.Stat(path); err == nil && fileInfo.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			if info, err := entry.Info(); err == nil && !entry.IsDir() && isFitFile(info) {
				fitFiles = append(fitFiles, filepath.Join(path, entry.Name()))
			}
		}
		return filesOrError(fitFiles, path)
	}

	// 3. Try as single file
	if isValidFitFile(path) {
		fitFiles = append(fitFiles, path)
	}

	return filesOrError(fitFiles, path)
}
