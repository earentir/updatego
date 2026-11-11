package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckInstallPermissionsOk(t *testing.T) {
	root := t.TempDir()
	goDir := filepath.Join(root, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := CheckInstallPermissions(root, goDir, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCheckInstallPermissionsExtractRoot(t *testing.T) {
	root := filepath.Join(os.TempDir(), "updatego-perm-ro-"+t.Name())
	if err := os.Mkdir(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755); _ = os.RemoveAll(root) })

	tmp := t.TempDir()
	goDir := filepath.Join(root, "go")
	err := CheckInstallPermissions(root, goDir, tmp)
	if err == nil {
		t.Fatal("expected error for non-writable extract root")
	}
	if !strings.Contains(err.Error(), "Extract root") {
		t.Fatalf("unexpected error: %v", err)
	}
}
