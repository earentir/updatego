package main

import (
	"fmt"
	"os"

	"updatego/config"
	"updatego/installer"
	"updatego/local"
	"updatego/update"

	cli "github.com/jawher/mow.cli"
)

var appversion = "1.3.33"

func main() {
	app := cli.App("updatego", "A simple golang version manager")
	app.Version("v version", fmt.Sprintf("updatego %s", appversion))

	verbose := app.BoolOpt("verbose", false, "Enable verbose output")

	app.Before = func() {
		config.GlobalConfig.Verbose = *verbose
	}

	app.Command("install", "Install Go", func(cmd *cli.Cmd) {
		version := cmd.StringOpt("version", "", "Specify Go version to install")
		force := cmd.BoolOpt("force f", false, "Force install by moving existing Go version")
		global := cmd.BoolOpt("global g", false, "Install globally")
		user := cmd.BoolOpt("user u", false, "Install for the user")
		customPath := cmd.StringOpt("custom-path", "", "Install to a custom path")

		cmd.Action = func() {
			installer.InstallGo(*version, *force, *global, *user, *customPath)
		}
	})

	app.Command("status", "Check Go installation status", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			local.CheckGoStatus()
		}
	})

	app.Command("latest", "Print the latest Go version available", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			local.PrintLatestGoVersion()
		}
	})

	app.Command("update", "Update Go to the latest version", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			update.Go()
		}
	})

	app.Command("list", "List all local Go versions", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			local.ListLocalVersions()
		}
	})

	app.Command("switch", "Switch to a specific Go version", func(cmd *cli.Cmd) {
		version := cmd.StringArg("VERSION", "", "Go version to switch to")
		cmd.Spec = "VERSION"
		cmd.Action = func() {
			local.SwitchGoVersion(*version)
		}
	})

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
