package utils

import (
	"runtime"
	"strings"
	"testing"
)

func TestCheckGoVersionReadsCurrentGOROOT(t *testing.T) {
	goroot := runtime.GOROOT()
	out, err := CheckGoVersion(goroot)
	if err != nil {
		t.Fatal(err)
	}
	version, platform := ParseGoVersion(out)
	if version == "" || version == "Unknown version" {
		t.Fatalf("unexpected version from %q", out)
	}
	if platform == "" || platform == "Unknown OS/Arch" {
		t.Fatalf("unexpected platform from %q", out)
	}
	if want := runtime.GOOS + "/" + runtime.GOARCH; !strings.EqualFold(platform, want) {
		// buildinfo reports the binary's target; on cross installs this may differ.
		t.Logf("platform %q (host %q)", platform, want)
	}
}

func TestNormalizeGoRelease(t *testing.T) {
	got, err := normalizeGoRelease("go1.26.4")
	if err != nil || got != "1.26.4" {
		t.Fatalf("got %q err=%v", got, err)
	}
}
