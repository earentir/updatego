package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeSwitchVersion(t *testing.T) {
	got, err := NormalizeSwitchVersion("1.25.0")
	if err != nil || got != "1.25.0" {
		t.Fatalf("got %q err %v", got, err)
	}
	got, err = NormalizeSwitchVersion("go1.24")
	if err != nil || got != "1.24.0" {
		t.Fatalf("got %q err %v", got, err)
	}
	_, err = NormalizeSwitchVersion("not-a-version")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetGoModToolchain_replace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	const before = `module example.com/m

go 1.22.0

toolchain go1.21.0
`
	if err := os.WriteFile(path, []byte(before), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SetGoModToolchain(dir, "1.25.0"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "toolchain go1.25.0") {
		t.Fatalf("unexpected go.mod:\n%s", data)
	}
	if strings.Contains(string(data), "go1.21.0") {
		t.Fatalf("old toolchain still present:\n%s", data)
	}
}

func TestSetGoModToolchain_insert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	const before = `module example.com/m

go 1.22.0

require example.com/foo v1.0.0
`
	if err := os.WriteFile(path, []byte(before), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SetGoModToolchain(dir, "1.25"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.Contains(s, "go 1.22.0") {
		t.Fatalf("go directive changed:\n%s", data)
	}
	if !strings.Contains(s, "toolchain go1.25.0") {
		t.Fatalf("toolchain not inserted after go line:\n%s", data)
	}
}

func TestSetGoModToolchain_missingFile(t *testing.T) {
	if err := SetGoModToolchain(t.TempDir(), "1.25.0"); err != nil {
		t.Fatal(err)
	}
}
