package fit

import (
	"fmt"
	"log"
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

func GetFitFiles(path string) ([]string, error) {
	var fitFiles []string

	log.Printf("param %v", path)

	// 1. Try glob pattern first
	if strings.ContainsAny(path, "*?[") {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %v", err)
		}
		log.Printf("matches %v", matches)
		for _, p := range matches {
			log.Printf("match %v", p)
			if isValidFitFile(p) {
				fitFiles = append(fitFiles, p)
			}
		}
		return fitFiles, nil
	}

	// 2. Try as directory
	if fileInfo, err := os.Stat(path); err == nil && fileInfo.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			if info, err := entry.Info(); err == nil && !entry.IsDir() && isFitFile(info) {
				fitFiles = append(fitFiles, filepath.Join(path, entry.Name()))
			}
		}
		return fitFiles, nil
	}

	// 3. Try as single file
	if isValidFitFile(path) {
		fitFiles = append(fitFiles, path)
	}

	return fitFiles, nil
}
