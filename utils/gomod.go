package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var switchVersionRE = regexp.MustCompile(`^(\d+)\.(\d+)(?:\.(\d+))?$`)

// NormalizeSwitchVersion accepts X.Y.Z or X.Y (optional go prefix) for switch/go.mod.
func NormalizeSwitchVersion(version string) (string, error) {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "go")
	if !switchVersionRE.MatchString(v) {
		return "", fmt.Errorf("invalid version %q: expected X.Y.Z or X.Y", version)
	}
	parts := strings.Split(v, ".")
	if len(parts) == 2 {
		v = v + ".0"
	}
	return v, nil
}

// ToolchainDirective returns the go.mod toolchain line body, e.g. go1.25.0.
func ToolchainDirective(version string) (string, error) {
	norm, err := NormalizeSwitchVersion(version)
	if err != nil {
		return "", err
	}
	return "go" + norm, nil
}

// SetGoModToolchain updates or inserts the toolchain line in dir/go.mod.
func SetGoModToolchain(dir, version string) error {
	path := filepath.Join(dir, "go.mod")
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}

	directive, err := ToolchainDirective(version)
	if err != nil {
		return err
	}
	newLine := "toolchain " + directive

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.SplitAfter(string(data), "\n")
	out, err := rewriteGoModToolchain(lines, newLine)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(out), info.Mode())
}

func rewriteGoModToolchain(lines []string, toolchainLine string) (string, error) {
	hasGo := false
	toolchainIdx := -1
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "go ") {
			hasGo = true
		}
		if strings.HasPrefix(trim, "toolchain ") {
			toolchainIdx = i
		}
	}
	if !hasGo {
		return "", fmt.Errorf("go.mod has no go directive")
	}

	if toolchainIdx >= 0 {
		lines[toolchainIdx] = toolchainLine + "\n"
		return strings.Join(lines, ""), nil
	}

	var out []string
	inserted := false
	for _, line := range lines {
		out = append(out, line)
		if !inserted && strings.HasPrefix(strings.TrimSpace(line), "go ") {
			out = append(out, toolchainLine+"\n")
			inserted = true
		}
	}
	if !inserted {
		return "", fmt.Errorf("go.mod has no go directive")
	}
	return strings.Join(out, ""), nil
}
