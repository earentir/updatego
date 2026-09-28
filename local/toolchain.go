package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"updatego/config"
	"updatego/utils"
)

// ManagedGoBinary returns the path to the managed go executable.
func ManagedGoBinary() (string, error) {
	path := filepath.Join(config.GlobalConfig.GoFullPath, "bin", "go")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("managed Go binary not found at %s: %w", path, err)
	}
	return path, nil
}

func runManagedGo(args ...string) (string, error) {
	bin, err := ManagedGoBinary()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", bin, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// PrintGOTOOLCHAIN prints GOTOOLCHAIN from the managed Go installation.
func PrintGOTOOLCHAIN() error {
	val, err := runManagedGo("env", "GOTOOLCHAIN")
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}

// SetGOTOOLCHAIN writes GOTOOLCHAIN via go env -w on the managed binary.
func SetGOTOOLCHAIN(value string) error {
	_, err := runManagedGo("env", "-w", "GOTOOLCHAIN="+value)
	return err
}

// RunToolchain manages GOTOOLCHAIN: print, auto, or local.
func RunToolchain(mode string) error {
	if err := utils.RefuseWindows(); err != nil {
		return err
	}
	switch mode {
	case "":
		return PrintGOTOOLCHAIN()
	case "auto", "local":
		if err := SetGOTOOLCHAIN(mode); err != nil {
			return err
		}
		fmt.Printf("GOTOOLCHAIN=%s\n", mode)
		return nil
	default:
		return fmt.Errorf("unknown toolchain mode %q (use auto or local)", mode)
	}
}

// SyncGOTOOLCHAIN sets GOTOOLCHAIN to goX.Y.Z matching the active managed Go.
func SyncGOTOOLCHAIN() error {
	if err := utils.RefuseWindows(); err != nil {
		return err
	}
	out, err := runManagedGo("version")
	if err != nil {
		return err
	}
	version, _ := utils.ParseGoVersion(out + "\n")
	if version == "" || version == "Unknown version" {
		return fmt.Errorf("could not parse Go version from: %s", out)
	}
	value := "go" + version
	if err := SetGOTOOLCHAIN(value); err != nil {
		return err
	}
	fmt.Printf("GOTOOLCHAIN=%s\n", value)
	return nil
}

// ApplySwitchToolchain sets GOTOOLCHAIN=local and updates ./go.mod when present.
func ApplySwitchToolchain(version string) error {
	if err := SetGOTOOLCHAIN("local"); err != nil {
		return err
	}
	fmt.Println("GOTOOLCHAIN=local")
	if err := utils.SetGoModToolchain(".", version); err != nil {
		return err
	}
	directive, err := utils.ToolchainDirective(version)
	if err != nil {
		return err
	}
	if _, err := os.Stat("go.mod"); err == nil {
		fmt.Printf("Updated go.mod toolchain to %s\n", directive)
	}
	return nil
}

// GOTOOLCHAINForStatus returns the current GOTOOLCHAIN or an error message.
func GOTOOLCHAINForStatus() string {
	val, err := runManagedGo("env", "GOTOOLCHAIN")
	if err != nil {
		return fmt.Sprintf("unavailable (%v)", err)
	}
	return val
}
