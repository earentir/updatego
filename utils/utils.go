// Package utils provides utility functions for the Go installer
package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// GoDownloadURL is the URL to download Go
	GoDownloadURL = "https://go.dev/dl/"
)

// DownloadHTML downloads the HTML content from the provided URL
func DownloadHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DownloadFile downloads a file from the provided URL
func DownloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tempDir := os.TempDir()
	tempfile := filepath.Join(tempDir, regexp.MustCompile(`[^/]+$`).FindString(url))

	fmt.Println("Writing to:", tempfile)

	out, err := os.Create(tempfile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return out.Name(), err
}

// DownloadFileWithProgress downloads a file with a progress indicator
func DownloadFileWithProgress(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tempDir := os.TempDir()
	tempfile := filepath.Join(tempDir, regexp.MustCompile(`[^/]+$`).FindString(url))

	fmt.Println("Writing to:", tempfile)

	out, err := os.Create(tempfile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	var downloadedSize int64
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return "", writeErr
			}
			downloadedSize += int64(n)
			if downloadedSize%(500*1024) < int64(n) {
				fmt.Print("#")
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}
	fmt.Println()
	return out.Name(), nil
}

// ExtractTarGz extracts a tarball to a target directory
func ExtractTarGz(filePath, extractPath string, isMainGoDir bool) error {
	gzFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var targetPath string
		if strings.HasPrefix(header.Name, "go/") {
			targetPath = filepath.Join(extractPath, strings.TrimPrefix(header.Name, "go/"))
		} else {
			targetPath = filepath.Join(extractPath, header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), os.FileMode(0755)); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %b in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

// FindVersion finds the Go version in the HTML content
func FindVersion(htmlContent string) (string, error) {
	regex := regexp.MustCompile(`go(\d+\.\d+\.\d+)\.linux-amd64\.tar\.gz`)
	matches := regex.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		err := errors.New("No version found")
		return "", err
	}
	return matches[1], nil
}

// BuildFilename builds the filename for the Go version
func BuildFilename(version string) string {
	return "go" + version + ".linux-amd64.tar.gz"
}

// RemoveGoFolder removes the Go folder
func RemoveGoFolder(path string) error {
	return os.RemoveAll(path)
}

// IsWritable checks if a path is writable
func IsWritable(path string) bool {
	tmpFilePath := filepath.Join(path, ".tmp-check")
	defer os.Remove(tmpFilePath)

	file, err := os.Create(tmpFilePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// DirNotEmpty checks if a directory is not empty
func DirNotEmpty(path string) bool {
	files, err := os.ReadDir(path)
	return err == nil && len(files) > 0
}

// CheckGoVersion checks the Go version in the provided path
func CheckGoVersion(path string) (string, error) {
	goVersionPath := filepath.Join(path, "bin", "go")
	output, err := exec.Command(goVersionPath, "version").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ParseGoVersion parses the Go version and OS/Arch from the output
func ParseGoVersion(output string) (string, string) {
	regex := regexp.MustCompile(`go version go(\d+\.\d+\.\d+) (.+/.+)`)
	matches := regex.FindStringSubmatch(output)
	if len(matches) < 3 {
		return "Unknown version", "Unknown OS/Arch"
	}
	return matches[1], matches[2]
}

// VersionExists checks if the version exists in the HTML content
func VersionExists(htmlContent, filename string) bool {
	return strings.Contains(htmlContent, filename)
}

// DetermineInstallPath determines the installation path
func DetermineInstallPath(global, user bool, customPath string) string {
	if global {
		return "/usr/local/"
	} else if user {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user home directory:", err)
			os.Exit(1)
		}
		return homeDir
	} else if customPath != "" {
		return customPath
	}
	return "/usr/local/"
}

// BackupOldGo backs up the old Go folder
func BackupOldGo(backupPath, goFullPath string) {
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		os.RemoveAll(backupPath)
	}
	if err := os.Rename(goFullPath, backupPath); err != nil {
		fmt.Println("Error renaming the old Go folder:", err)
		os.Exit(1)
	}
}

// GetVersionToInstall returns the version to install
func GetVersionToInstall(version string) string {
	htmlContent, err := DownloadHTML(GoDownloadURL)
	if err != nil {
		fmt.Println("Error downloading HTML content:", err)
		os.Exit(1)
	}

	latestVersion, err := FindVersion(htmlContent)
	if err != nil {
		fmt.Println("Error finding the latest version:", err)
		os.Exit(1)
	}

	if version == "" {
		version = latestVersion
	} else if version != latestVersion {
		filename := BuildFilename(version)
		if !VersionExists(htmlContent, filename) {
			fmt.Printf("Requested version %s is not available. Latest version is %s.\n", version, latestVersion)
			os.Exit(1)
		}
	}

	return version
}

// DownloadAndVerifyFile downloads a file and verifies the write permission in the temp directory
func DownloadAndVerifyFile(url string) (string, error) {
	filePath, err := DownloadFile(url)
	if err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp(os.TempDir(), "test_write_")
	if err != nil {
		fmt.Printf("No write permission in temp directory: %s\n", os.TempDir())
		os.Exit(1)
	}

	fmt.Printf("Write permission confirmed in temp directory: %s\n", os.TempDir())
	tempFile.Close()
	os.Remove(tempFile.Name())

	return filePath, nil
}

// SetEnvironmentVariables sets the GOROOT and GOPATH environment variables
func SetEnvironmentVariables(goFullPath string) {
	fmt.Println("Setting up environment variables...")
	os.Setenv("GOROOT", goFullPath)
	os.Setenv("GOPATH", filepath.Join(os.Getenv("HOME"), "go"))
	fmt.Println("GOROOT set to:", goFullPath)
	fmt.Println("GOPATH set to:", filepath.Join(os.Getenv("HOME"), "go"))
}

// DetermineInstallType determines the type of Go installation
func DetermineInstallType(goFullPath string) string {
	if strings.Contains(goFullPath, os.TempDir()) {
		return "User"
	} else if strings.Contains(goFullPath, "/usr/local/") {
		return "Global"
	}
	return "Custom"
}

// IsGoInPath checks if the `go` binary is in PATH
func IsGoInPath(goFullPath string) bool {
	pathDirs := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range pathDirs {
		if dir == filepath.Join(goFullPath, "bin") {
			return true
		}
	}
	return false
}

// IsDirExists checks if a directory exists
func IsDirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetLatestVersion returns the latest Go version available
func GetLatestVersion() (string, error) {
	htmlContent, err := DownloadHTML(GoDownloadURL)
	if err != nil {
		return "", err
	}

	return FindVersion(htmlContent)
}

// CopyDir copies a directory recursively
func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, src)
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return err
			}
		} else {
			if err := copyFile(path, targetPath); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, sourceFileInfo.Mode()); err != nil {
		return err
	}

	return nil
}
