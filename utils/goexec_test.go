package utils

import (
	"slices"
	"testing"
)

func TestGoTreeEnvOffline(t *testing.T) {
	t.Setenv("GOTOOLCHAIN", "auto")
	t.Setenv("GOROOT", "/wrong")
	t.Setenv("GOPROXY", "https://proxy.golang.org")

	env := GoTreeEnv("/usr/local/go")
	for _, tc := range []struct{ k, v string }{
		{"GOTOOLCHAIN", "local"},
		{"GOROOT", "/usr/local/go"},
		{"GOPROXY", "off"},
		{"GOSUMDB", "off"},
	} {
		got := envValue(env, tc.k)
		if got != tc.v {
			t.Fatalf("%s=%q, want %q", tc.k, got, tc.v)
		}
		if n := countEnvKey(env, tc.k); n != 1 {
			t.Fatalf("%s appears %d times, want 1", tc.k, n)
		}
	}
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if stringsHasPrefix(e, prefix) {
			return e[len(prefix):]
		}
	}
	return ""
}

func countEnvKey(env []string, key string) int {
	n := 0
	for _, e := range env {
		if stringsHasPrefix(e, key+"=") || e == key {
			n++
		}
	}
	return n
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func TestWithoutEnvKey(t *testing.T) {
	base := []string{"HOME=/tmp", "GOTOOLCHAIN=auto", "PATH=/bin"}
	out := withoutEnvKey(base, "GOTOOLCHAIN")
	if slices.Contains(out, "GOTOOLCHAIN=auto") {
		t.Fatal("GOTOOLCHAIN should be removed")
	}
	if !slices.Contains(out, "HOME=/tmp") {
		t.Fatal("HOME should remain")
	}
}

func TestRunGoTreeRejectsEmptyRoot(t *testing.T) {
	if _, err := RunGoTree("", "version"); err == nil {
		t.Fatal("expected error for empty goroot")
	}
}
