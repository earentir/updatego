package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"updatego/config"
)

const (
	metadataTimeout = 30 * time.Second
	downloadTimeout = 30 * time.Minute
)

var metadataClient = &http.Client{Timeout: metadataTimeout}
var downloadClient = &http.Client{Timeout: downloadTimeout}

// RefuseWindows returns an error when zip-based installs are not supported.
func RefuseWindows() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows zip-based Go installs are not supported by updatego")
	}
	return nil
}

func httpGet(url string) (*http.Response, error) {
	return downloadClient.Get(url)
}

func closeBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		fmt.Println("Error closing response body:", err)
	}
}

func archiveFilenameFromURL(url string) string {
	return regexp.MustCompile(`[^/]+$`).FindString(url)
}

func configTempDir() string {
	if config.GlobalConfig != nil && config.GlobalConfig.TempDir != "" {
		return config.GlobalConfig.TempDir
	}
	return os.TempDir()
}

type dlRelease struct {
	Version string   `json:"version"`
	Stable  bool     `json:"stable"`
	Files   []dlFile `json:"files"`
}

type dlFile struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
	Kind     string `json:"kind"`
}

// sha256FromDLJSON finds the SHA256 for filename in go.dev/?mode=json payload.
func sha256FromDLJSON(data []byte, filename string) (string, error) {
	var releases []dlRelease
	if err := json.Unmarshal(data, &releases); err != nil {
		return "", fmt.Errorf("invalid go.dev download metadata: %w", err)
	}
	for _, r := range releases {
		for _, f := range r.Files {
			if f.Filename == filename && f.SHA256 != "" {
				return f.SHA256, nil
			}
		}
	}
	return "", fmt.Errorf("no SHA256 found for %s in go.dev download metadata", filename)
}

// fetchSHA256ForFilename looks up the archive checksum via go.dev/?mode=json.
// Sidecar URLs like *.tar.gz.sha256 are not published; they return HTML.
func fetchSHA256ForFilename(filename string) (string, error) {
	resp, err := metadataClient.Get(GoDownloadURL + "?mode=json")
	if err != nil {
		return "", err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for download metadata", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sum, err := sha256FromDLJSON(body, filename)
	if err != nil {
		return "", err
	}
	if len(sum) != 64 || strings.Trim(sum, "0123456789abcdefABCDEF") != "" {
		return "", fmt.Errorf("invalid SHA256 for %s: %q", filename, sum)
	}
	return sum, nil
}

// DownloadArchive downloads an archive to a temp file and verifies its SHA256 checksum.
func DownloadArchive(fileURL string) (string, error) {
	if err := RefuseWindows(); err != nil {
		return "", err
	}

	resp, err := httpGet(fileURL)
	if err != nil {
		return "", err
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, fileURL)
	}

	filename := archiveFilenameFromURL(fileURL)
	tempfile := filepath.Join(configTempDir(), filename)
	fmt.Println("URL:", fileURL)
	fmt.Println("Writing to:", tempfile)

	out, err := os.Create(tempfile)
	if err != nil {
		return "", err
	}

	var downloadedSize int64
	buffer := make([]byte, 32*1024)
	hasher := sha256.New()
	mw := io.MultiWriter(out, hasher)

	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := mw.Write(buffer[:n]); writeErr != nil {
				out.Close()
				return "", writeErr
			}
			downloadedSize += int64(n)
			if downloadedSize%(500*1024) < int64(n) {
				fmt.Print("#")
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			out.Close()
			return "", readErr
		}
	}
	fmt.Println()
	if err := out.Close(); err != nil {
		return "", err
	}

	if downloadedSize == 0 {
		return "", fmt.Errorf("downloaded file is empty")
	}

	expected, err := fetchSHA256ForFilename(filename)
	if err != nil {
		return "", err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return "", fmt.Errorf("checksum mismatch for %s: got %s, want %s", filename, actual, expected)
	}

	return tempfile, nil
}
