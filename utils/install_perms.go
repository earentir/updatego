package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PermCheck is one install/update preflight result.
type PermCheck struct {
	Name   string // short label for errors
	Test   string // description shown by updatego test
	Pass   bool
	Detail string
}

// InspectInstallPermissions runs filesystem preflight checks without mutating anything.
func InspectInstallPermissions(extractRoot, goFullPath, tempDir string) []PermCheck {
	extractRoot = filepath.Clean(extractRoot)
	goFullPath = filepath.Clean(goFullPath)
	tempDir = filepath.Clean(tempDir)

	checks := []PermCheck{
		tempPermCheck(tempDir),
		extractRootPermCheck(extractRoot),
	}
	if IsDirExists(goFullPath) {
		parent := filepath.Dir(goFullPath)
		ok := IsWritable(parent)
		detail := "ok"
		if !ok {
			detail = fmt.Sprintf("parent %s is not writable", parent)
		}
		checks = append(checks, PermCheck{
			Name:   "Move active install",
			Test:   "Active Go install can be renamed aside for backup",
			Pass:   ok,
			Detail: detail,
		})
	} else {
		checks = append(checks, PermCheck{
			Name:   "Move active install",
			Test:   "Active Go install can be renamed aside for backup",
			Pass:   true,
			Detail: "no active tree at " + goFullPath,
		})
	}
	return checks
}

func tempPermCheck(tempDir string) PermCheck {
	ok := IsWritable(tempDir)
	detail := tempDir
	if !ok {
		detail = tempDir + " (not writable)"
	}
	return PermCheck{
		Name: "Temp directory",
		Test: "Temp directory is writable (needed for downloads)",
		Pass: ok, Detail: detail,
	}
}

func extractRootPermCheck(extractRoot string) PermCheck {
	ok := IsWritable(extractRoot)
	detail := extractRoot
	if !ok {
		detail = extractRoot + " (not writable)"
	}
	return PermCheck{
		Name: "Extract root",
		Test: "Extract root is writable (needed for install/update/switch)",
		Pass: ok, Detail: detail,
	}
}

// CheckInstallPermissions verifies temp, extract root, and (when present) the active
// go tree can be moved aside before install, update, or switch run downloads or extracts.
func CheckInstallPermissions(extractRoot, goFullPath, tempDir string) error {
	var problems []string
	for _, c := range InspectInstallPermissions(extractRoot, goFullPath, tempDir) {
		if !c.Pass {
			problems = append(problems, fmt.Sprintf("%s: %s", c.Name, c.Detail))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	extractRoot = filepath.Clean(extractRoot)
	return fmt.Errorf("%s%s", strings.Join(problems, "; "), permissionHint(extractRoot))
}

func permissionHint(extractRoot string) string {
	if extractRoot == "/usr/local" || strings.HasPrefix(extractRoot, "/usr/local/") {
		return ", re-run with elevated permissions"
	}
	return ""
}
