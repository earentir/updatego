package utils

import (
	"bytes"
	"debug/buildinfo"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
)

var goVersionLineRE = regexp.MustCompile(`^go(\d+\.\d+(?:\.\d+)?(?:\.\d+)?)`)

// CheckGoVersion reports the toolchain version for an install tree.
// It reads build metadata embedded in bin/go; if that fails, it runs
// go version with offline env (no toolchain or module network use).
func CheckGoVersion(goroot string) (string, error) {
	if version, platform, err := goTreeVersionFromBuildInfo(goroot); err == nil {
		return fmt.Sprintf("go version go%s %s\n", version, platform), nil
	}
	out, err := RunGoTree(goroot, "version")
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(out))
	if !bytes.HasPrefix([]byte(line), []byte("go version go")) {
		return "", fmt.Errorf("unexpected go version output: %s", line)
	}
	return line + "\n", nil
}

func goTreeVersionFromBuildInfo(goroot string) (version, platform string, err error) {
	if goroot == "" {
		return "", "", fmt.Errorf("empty Go install root")
	}
	goBin := filepath.Join(goroot, "bin", "go")

	info, err := buildinfo.ReadFile(goBin)
	if err != nil {
		return "", "", fmt.Errorf("cannot read build info from %s: %w", goBin, err)
	}
	if info.GoVersion == "" {
		return "", "", fmt.Errorf("no Go version in build info for %s", goBin)
	}
	version, err = normalizeGoRelease(info.GoVersion)
	if err != nil {
		return "", "", err
	}
	goos := buildSetting(info.Settings, "GOOS")
	goarch := buildSetting(info.Settings, "GOARCH")
	if goos == "" || goarch == "" {
		return "", "", fmt.Errorf("no GOOS/GOARCH in build info for %s", goBin)
	}
	return version, goos + "/" + goarch, nil
}

func buildSetting(settings []debug.BuildSetting, key string) string {
	for _, s := range settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

func normalizeGoRelease(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "go")
	if m := goVersionLineRE.FindStringSubmatch("go" + raw); len(m) >= 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("unrecognized Go version %q", raw)
}
