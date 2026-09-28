package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func safeTarTarget(extractPath, headerName string) (string, error) {
	extractPath = filepath.Clean(extractPath)
	rel := headerName
	if strings.HasPrefix(headerName, "go/") {
		rel = strings.TrimPrefix(headerName, "go/")
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid tar path: %s", headerName)
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid tar path: %s", headerName)
	}
	target := filepath.Join(extractPath, rel)
	absExtract, err := filepath.Abs(extractPath)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if absTarget != absExtract && !strings.HasPrefix(absTarget, absExtract+string(os.PathSeparator)) {
		return "", fmt.Errorf("tar path escapes destination: %s", headerName)
	}
	return target, nil
}
