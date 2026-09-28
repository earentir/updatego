package utils

import (
	"runtime"
	"strings"
	"testing"
)

func TestPlatformParts(t *testing.T) {
	tests := []struct {
		goos, goarch string
		wantOS       string
		wantArch     string
		wantExt      string
	}{
		{"darwin", "arm64", "darwin", "arm64", ".tar.gz"},
		{"darwin", "amd64", "darwin", "amd64", ".tar.gz"},
		{"linux", "amd64", "linux", "amd64", ".tar.gz"},
		{"linux", "arm64", "linux", "arm64", ".tar.gz"},
		{"linux", "arm", "linux", "armv6l", ".tar.gz"},
		{"freebsd", "amd64", "freebsd", "amd64", ".tar.gz"},
		{"windows", "amd64", "windows", "amd64", ".zip"},
		{"windows", "arm64", "windows", "arm64", ".zip"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"-"+tt.goarch, func(t *testing.T) {
			osName, arch, ext := platformParts(tt.goos, tt.goarch)
			if osName != tt.wantOS || arch != tt.wantArch || ext != tt.wantExt {
				t.Fatalf("platformParts(%q, %q) = %q, %q, %q; want %q, %q, %q",
					tt.goos, tt.goarch, osName, arch, ext, tt.wantOS, tt.wantArch, tt.wantExt)
			}
		})
	}
}

func TestBuildFilenameFor(t *testing.T) {
	tests := []struct {
		version, goos, goarch, want string
	}{
		{"1.25.0", "darwin", "arm64", "go1.25.0.darwin-arm64.tar.gz"},
		{"1.25.0", "darwin", "amd64", "go1.25.0.darwin-amd64.tar.gz"},
		{"1.25.0", "linux", "amd64", "go1.25.0.linux-amd64.tar.gz"},
		{"1.24.5", "linux", "arm", "go1.24.5.linux-armv6l.tar.gz"},
		{"1.23.4", "windows", "amd64", "go1.23.4.windows-amd64.zip"},
	}

	for _, tt := range tests {
		got := buildFilenameFor(tt.version, tt.goos, tt.goarch)
		if got != tt.want {
			t.Errorf("buildFilenameFor(%q, %q, %q) = %q; want %q",
				tt.version, tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestBuildFilenameMatchesCurrentRuntime(t *testing.T) {
	got := BuildFilename("1.25.0")
	wantOS := runtime.GOOS
	wantArch := runtime.GOARCH
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm" {
		wantArch = "armv6l"
	}
	wantExt := ".tar.gz"
	if runtime.GOOS == "windows" {
		wantExt = ".zip"
	}
	want := "go1.25.0." + wantOS + "-" + wantArch + wantExt
	if got != want {
		t.Fatalf("BuildFilename() = %q; want %q for this host", got, want)
	}
	if strings.Contains(got, "linux-amd64") && (runtime.GOOS != "linux" || runtime.GOARCH != "amd64") {
		t.Fatalf("BuildFilename() hardcoded linux-amd64 on %s/%s: %s", runtime.GOOS, runtime.GOARCH, got)
	}
}

func TestFindVersionForCurrentPlatform(t *testing.T) {
	html := `
		<a href="/dl/go1.24.9.linux-amd64.tar.gz">go1.24.9.linux-amd64.tar.gz</a>
		<a href="/dl/go1.25.0.linux-amd64.tar.gz">go1.25.0.linux-amd64.tar.gz</a>
		<a href="/dl/go1.25.0.linux-arm64.tar.gz">go1.25.0.linux-arm64.tar.gz</a>
		<a href="/dl/go1.25.0.darwin-amd64.tar.gz">go1.25.0.darwin-amd64.tar.gz</a>
		<a href="/dl/go1.25.0.darwin-arm64.tar.gz">go1.25.0.darwin-arm64.tar.gz</a>
		<a href="/dl/go1.25.0.windows-amd64.zip">go1.25.0.windows-amd64.zip</a>
	`

	tests := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "1.25.0"},
		{"darwin", "amd64", "1.25.0"},
		{"linux", "amd64", "1.25.0"},
		{"linux", "arm64", "1.25.0"},
		{"windows", "amd64", "1.25.0"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"-"+tt.goarch, func(t *testing.T) {
			got, err := findVersionFor(html, tt.goos, tt.goarch)
			if err != nil {
				t.Fatalf("findVersionFor(..., %q, %q) unexpected error: %v", tt.goos, tt.goarch, err)
			}
			if got != tt.want {
				t.Fatalf("findVersionFor(..., %q, %q) = %q; want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestFindVersionForMissingPlatform(t *testing.T) {
	html := `<a href="/dl/go1.25.0.linux-amd64.tar.gz">go1.25.0.linux-amd64.tar.gz</a>`
	_, err := findVersionFor(html, "darwin", "arm64")
	if err == nil {
		t.Fatal("expected error when darwin-arm64 archive is missing")
	}
}

func TestIsHostPlatform(t *testing.T) {
	host := runtime.GOOS + "/" + runtime.GOARCH
	if !IsHostPlatform(host) {
		t.Fatalf("IsHostPlatform(%q) = false; want true", host)
	}
	if IsHostPlatform("linux/amd64") && host != "linux/amd64" {
		t.Fatalf("IsHostPlatform(linux/amd64) = true on %s", host)
	}
	if HostOSArch() != host {
		t.Fatalf("HostOSArch() = %q; want %q", HostOSArch(), host)
	}
}
