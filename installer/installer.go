// Package installer provides functions to install Go versions
package installer

import (
	"fmt"
	"path/filepath"

	"updatego/config"
	"updatego/local"
	"updatego/utils"
)

// InstallGo installs the specified Go version.
func InstallGo(version string, force, global, user bool, customPath string) error {
	if err := utils.RefuseWindows(); err != nil {
		return err
	}

	extractRoot, err := utils.ResolveInstallRoot(global, user, customPath)
	if err != nil {
		return err
	}
	config.ApplyExtractRoot(extractRoot)

	if !utils.IsWritable(config.GlobalConfig.GoExtractPathRoot) {
		return fmt.Errorf("installation directory %s is not writable", config.GlobalConfig.GoExtractPathRoot)
	}

	if version == "" {
		version, err = utils.GetVersionToInstall("")
		if err != nil {
			return err
		}
	} else {
		version, err = utils.GetVersionToInstall(version)
		if err != nil {
			return err
		}
	}

	if utils.DirNotEmpty(config.GlobalConfig.GoFullPath) {
		if err := handleExistingInstallation(version, force); err != nil {
			return err
		}
	} else {
		if err := installNewVersion(version); err != nil {
			return err
		}
	}

	if err := config.SaveExtractRoot(config.GlobalConfig.GoExtractPathRoot); err != nil {
		return fmt.Errorf("install succeeded but could not save config: %w", err)
	}

	utils.PrintShellHints(config.GlobalConfig.GoFullPath)
	return nil
}

func handleExistingInstallation(version string, force bool) error {
	currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err != nil {
		fmt.Printf("Error checking current Go version: %v\n", err)
		fmt.Println("Existing installation is unusable; replacing it.")
		return replaceOrInstall(version)
	}

	parsedCurrentVersion, osArch := utils.ParseGoVersion(currentVersion)
	fmt.Printf("Go is already installed. Current version: %s (%s)\n", parsedCurrentVersion, osArch)

	wrongPlatform := !utils.IsHostPlatform(osArch)
	if wrongPlatform {
		fmt.Printf("Installed Go is %s, this machine is %s. Replacing it.\n", osArch, utils.HostOSArch())
	}

	if parsedCurrentVersion == version && !force && !wrongPlatform {
		fmt.Printf("Go version %s is already installed. Use the --force flag to reinstall it.\n", version)
		return nil
	}

	return replaceOrInstall(version)
}

func replaceOrInstall(version string) error {
	localPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)
	if utils.GoTreeReady(localPath) {
		fmt.Printf("Go version %s is already available locally. Switching to this version.\n", version)
		return local.SwitchGoVersion(version)
	}
	return installNewVersion(version)
}

func installNewVersion(version string) error {
	targetPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)
	fmt.Printf("Installing Go version: %s\n", version)
	if err := installGoVersion(version, targetPath); err != nil {
		return fmt.Errorf("error installing Go version %s: %w", version, err)
	}

	fmt.Printf("Switching to the newly installed Go version: %s\n", version)
	return local.SwitchGoVersion(version)
}

func installGoVersion(version, installPath string) error {
	filename := utils.BuildFilename(version)
	fmt.Printf("Downloading %s\n", filename)
	filePath, err := utils.DownloadArchive(utils.GoDownloadURL + filename)
	if err != nil {
		return err
	}

	fmt.Println("Extracting the new Go version...")
	if err := utils.ExtractTarGz(filePath, installPath, false); err != nil {
		return err
	}
	if !utils.GoTreeReady(installPath) {
		return fmt.Errorf("extracted tree at %s is not a usable Go installation", installPath)
	}
	return nil
}
