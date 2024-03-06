package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

const (
	goDownloadURL = "https://go.dev/dl/"
	// goDownloadPath    = "/tmp/"
	goExtractPathRoot = "/usr/local/"
	goFullPath        = goExtractPathRoot + "go"
)

var (
	tempDir = ""
)

func main() {
	tempDir = os.TempDir()

	fmt.Println("Downloading Go Data to get the latest version...")
	htmlContent, err := downloadHTML(goDownloadURL)
	if err != nil {
		panic(err)
	}

	fmt.Println("Finding the latest version...")
	version, err := findVersion(htmlContent)
	if err != nil {
		panic(err)
	}
	fmt.Println("Latest version found:", version)

	// Check write permission by attempting to create a temporary file
	tempFile, err := os.CreateTemp(tempDir, "test_write_")
	if err != nil {
		fmt.Printf("No write permission in temp directory: %s\n", tempDir)
		panic(err)
	} else {
		fmt.Printf("Write permission confirmed in temp directory: %s\n", tempDir)
		// Clean up after the test by removing the temporary file
		tempFile.Close()
		os.Remove(tempFile.Name())
	}

	filename := buildFilename(version)

	fmt.Println("Downloading the latest version:", filename, " From: ", goDownloadURL+filename)
	fileURL := goDownloadURL + filename
	filePath, err := downloadFile(fileURL)
	if err != nil {
		panic(err)
	}

	fmt.Println("Removing the old Go folder...")
	if err := removeGoFolder(goFullPath); err != nil {
		panic(err)
	}

	fmt.Println("Extracting the new Go version...")
	if err := extractTarGz(filePath, goExtractPathRoot); err != nil {
		panic(err)
	}
}

func findVersion(htmlContent string) (string, error) {
	// Updated regex to capture just the version part
	regex := regexp.MustCompile(`go(\d+\.\d+\.\d+)\.linux-amd64\.tar\.gz`)
	matches := regex.FindStringSubmatch(htmlContent)
	if len(matches) < 2 { // matches[0] is the full match, matches[1] should be the version
		err := errors.New("No version found")
		return "", err
	}
	// Return just the version part
	return matches[1], nil
}

func buildFilename(version string) string {
	// Construct the filename using the version
	return "go" + version + ".linux-amd64.tar.gz"
}

func removeGoFolder(path string) error {
	return os.RemoveAll(path)
}
