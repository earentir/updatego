package utils

import (
	"crypto/sha256"
	"encoding/hex"
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

const metadataTimeout = 30 * time.Second

var downloadClient = &http.Client{Timeout: metadataTimeout}

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

	sumURL := fileURL + ".sha256"
	expected, err := fetchSHA256Sum(sumURL)
	if err != nil {
		return "", err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return "", fmt.Errorf("checksum mismatch for %s", filename)
	}

	return tempfile, nil
}

func fetchSHA256Sum(sumURL string) (string, error) {
	resp, err := httpGet(sumURL)
	if err != nil {
		return "", err
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, sumURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(body))
	if len(fields) < 1 {
		return "", fmt.Errorf("invalid checksum file at %s", sumURL)
	}
	return fields[0], nil
}
