package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"updatego/config"
	"updatego/utils"
)

// ManagedGoVersion returns the semver of the managed install, if readable.
func ManagedGoVersion() (string, error) {
	goFullPath := config.GlobalConfig.GoFullPath
	out, err := utils.CheckGoVersion(goFullPath)
	if err != nil {
		return "", err
	}
	version, _ := utils.ParseGoVersion(out)
	if version == "" || version == "Unknown version" {
		return "", fmt.Errorf("could not read Go version at %s", goFullPath)
	}
	return version, nil
}

// RunCurrent verifies the managed Go installation and shell environment.
// With brief, it prints only the version string.
func RunCurrent(brief bool) error {
	if brief {
		version, err := ManagedGoVersion()
		if err != nil {
			return err
		}
		fmt.Println(version)
		return nil
	}

	goFullPath := config.GlobalConfig.GoFullPath
	rows := make([]checkRow, 0, 16)

	version, verr := ManagedGoVersion()
	if verr != nil {
		fmt.Println("Current Go version: unknown")
	} else {
		fmt.Printf("Current Go version: %s\n", version)
	}
	fmt.Println()

	if err := utils.RefuseWindows(); err != nil {
		rows = append(rows, checkRow{test: "Operating system is supported", kind: statusFail, detail: err.Error()})
		printCheckTable(rows)
		printCheckSummary(rows, "checks")
		return err
	}
	rows = append(rows, checkRow{test: "Operating system is supported", kind: statusOK, detail: "yes"})

	rows = append(rows, checkRow{
		test:   "Configuration load",
		kind:   statusOK,
		detail: "extract root " + config.GlobalConfig.GoExtractPathRoot,
	})

	addCheckInfo(&rows, "Install layout", utils.DetermineInstallType(goFullPath))

	if !utils.IsDirExists(goFullPath) {
		addCheck(&rows, "Managed GOROOT directory exists", false, "missing "+goFullPath)
	} else {
		addCheck(&rows, "Managed GOROOT directory exists", true, goFullPath)

		if verr != nil {
			addCheck(&rows, "Go version is readable", false, verr.Error())
		} else {
			out, _ := utils.CheckGoVersion(goFullPath)
			_, osArch := utils.ParseGoVersion(out)
			addCheck(&rows, "Go version is readable", true, version)
			if utils.IsHostPlatform(osArch) {
				addCheck(&rows, "Go binary matches host platform", true, osArch)
			} else {
				addCheck(&rows, "Go binary matches host platform", false,
					osArch+" (host "+utils.HostOSArch()+")")
			}
		}

		if utils.GoTreeReady(goFullPath) {
			addCheck(&rows, "Managed Go tree is usable", true, "bin/go ok")
		} else {
			addCheck(&rows, "Managed Go tree is usable", false, "bin/go missing or invalid")
		}
	}

	if utils.IsGoInPath(goFullPath) {
		addCheck(&rows, "Managed go binary is on PATH", true, filepath.Join(goFullPath, "bin", "go"))
	} else {
		addCheck(&rows, "Managed go binary is on PATH", false, "add "+filepath.Join(goFullPath, "bin")+" to PATH")
	}

	if goroot := os.Getenv("GOROOT"); goroot == "" {
		addCheck(&rows, "GOROOT environment variable", true, "unset (derived from binary)")
	} else if goroot == goFullPath {
		addCheck(&rows, "GOROOT environment variable", true, goroot)
	} else {
		addCheck(&rows, "GOROOT environment variable", false, goroot+" (want "+goFullPath+")")
	}

	toolchain := GOTOOLCHAINForStatus()
	if strings.HasPrefix(toolchain, "unavailable") {
		addCheckWarn(&rows, "GOTOOLCHAIN readable from managed go", toolchain)
	} else {
		addCheck(&rows, "GOTOOLCHAIN readable from managed go", true, toolchain)
	}

	expectedGOPATH := filepath.Join(os.Getenv("HOME"), "go")
	if gopath := os.Getenv("GOPATH"); gopath == "" {
		addCheck(&rows, "GOPATH environment variable", true, "unset (defaults to "+expectedGOPATH+")")
	} else if gopath == expectedGOPATH {
		addCheck(&rows, "GOPATH environment variable", true, gopath)
	} else {
		addCheckWarn(&rows, "GOPATH environment variable", gopath+" (default would be "+expectedGOPATH+")")
	}

	return finishReport(rows, "checks")
}
