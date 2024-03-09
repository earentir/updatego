package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	goDownloadURL = "https://go.dev/dl/"
)

var (
	tempDir           = ""
	goExtractPathRoot = "/usr/local/"
	goFullPath        = ""
)

func main() {
	tempDir = os.TempDir()

	// Define a string flag
	// Check if at least one argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: program [user|global|<custom path>]")
		os.Exit(1)
	}

	// The first parameter after the program name
	installType := os.Args[1]

	// Determine the install path based on the installType flag
	switch installType {
	case "user":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user home directory:", err)
			os.Exit(1)
		}
		goExtractPathRoot = homeDir
	case "global":
		goExtractPathRoot = "/usr/local/"
	default:
		if installType == "" {
			fmt.Println("Usage: program --installType=[user|global|<custom path>]")
			os.Exit(1)
		}
		goExtractPathRoot = installType
	}

	// Check if the installPath is a valid, writable path
	if isWritable(goExtractPathRoot) {
		fmt.Printf("Install path is set to: %s\n", goExtractPathRoot)
	} else {
		fmt.Println("The provided install path is not valid or not writable.")
		os.Exit(1)
	}

	goFullPath = goExtractPathRoot + "go"

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

// isWritable checks if the path is writable
func isWritable(path string) bool {
	// Try creating a temporary file at the path to check for write permission
	tmpFilePath := filepath.Join(path, ".tmp-check")
	defer os.Remove(tmpFilePath) // Clean up after the check

	file, err := os.Create(tmpFilePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}
