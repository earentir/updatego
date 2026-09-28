// Package utils provides utility functions for the Go installer
package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"updatego/config"
)

const (
	// GoDownloadURL is the URL to download Go
	GoDownloadURL = "https://go.dev/dl/"
)

// DownloadHTML downloads the HTML content from the provided URL.
func DownloadHTML(url string) (string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ExtractTarGz extracts a tarball to a target directory
func ExtractTarGz(filePath, extractPath string, isMainGoDir bool) error {
	gzFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzFile.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath, err := safeTarTarget(extractPath, header.Name)
		if err != nil {
			return err
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
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, copyErr := io.Copy(outFile, tarReader); copyErr != nil {
				outFile.Close()
				return copyErr
			}
			if err := outFile.Close(); err != nil {
				return err
			}
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

// GoPlatform returns the OS, architecture, and archive extension used in
// official Go download filenames for the current machine.
func GoPlatform() (osName, arch, ext string) {
	return platformParts(runtime.GOOS, runtime.GOARCH)
}

func platformParts(goos, goarch string) (osName, arch, ext string) {
	osName = goos
	arch = goarch
	if goos == "linux" && goarch == "arm" {
		arch = "armv6l"
	}
	ext = ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return osName, arch, ext
}

// FindVersion finds the latest Go version for the current OS and architecture
// in the HTML content from go.dev/dl.
func FindVersion(htmlContent string) (string, error) {
	return findVersionFor(htmlContent, runtime.GOOS, runtime.GOARCH)
}

// BuildFilename builds the official Go archive filename for the current OS and architecture.
func BuildFilename(version string) string {
	return buildFilenameFor(version, runtime.GOOS, runtime.GOARCH)
}

func buildFilenameFor(version, goos, goarch string) string {
	osName, arch, ext := platformParts(goos, goarch)
	return "go" + version + "." + osName + "-" + arch + ext
}

// HostOSArch returns the current machine as GOOS/GOARCH, matching `go version` output.
func HostOSArch() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// IsHostPlatform reports whether osArch (for example "darwin/arm64") matches this machine.
func IsHostPlatform(osArch string) bool {
	return osArch == HostOSArch()
}

// RemoveGoFolder removes the Go folder
func RemoveGoFolder(path string) error {
	return os.RemoveAll(path)
}

// IsWritable checks if a path is writable
func IsWritable(path string) bool {
	tmpFilePath := filepath.Join(path, ".tmp-check")
	defer func() {
		if err := os.Remove(tmpFilePath); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	file, err := os.Create(tmpFilePath)
	if err != nil {
		return false
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()
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

// GetVersionToInstall resolves the version to install.
func GetVersionToInstall(version string) (string, error) {
	htmlContent, err := DownloadHTML(GoDownloadURL)
	if err != nil {
		return "", fmt.Errorf("error downloading HTML content: %w", err)
	}

	latestVersion, err := FindVersion(htmlContent)
	if err != nil {
		return "", fmt.Errorf("error finding the latest version: %w", err)
	}

	if version == "" {
		return latestVersion, nil
	}
	if version != latestVersion {
		filename := BuildFilename(version)
		if !VersionExists(htmlContent, filename) {
			return "", fmt.Errorf("requested version %s is not available; latest stable is %s", version, latestVersion)
		}
	}
	return version, nil
}

// GoTreeReady reports whether path looks like a usable Go tree.
func GoTreeReady(path string) bool {
	if !IsDirExists(path) {
		return false
	}
	_, err := CheckGoVersion(path)
	return err == nil
}

// DetermineInstallType determines the type of Go installation
func DetermineInstallType(goFullPath string) string {
	if strings.Contains(goFullPath, os.TempDir()) {
		return "User"
	}
	if home, err := os.UserHomeDir(); err == nil && strings.Contains(goFullPath, filepath.Join(home, ".local")) {
		return "User"
	}
	if strings.Contains(goFullPath, "/usr/local/") || goFullPath == filepath.Join("/usr/local", "go") {
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
	defer func() {
		if err := sourceFile.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

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

func Log(message string) {
	if config.GlobalConfig.Verbose {
		fmt.Println(message)
	}
}

func VerifyDownloadedFile(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error checking downloaded file: %v", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty")
	}
	return nil
}
