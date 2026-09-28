package main

import (
	"fmt"
	"os"

	"updatego/config"
	"updatego/installer"
	"updatego/local"
	"updatego/update"

	"github.com/spf13/cobra"
)

var appversion = "1.3.39"

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "updatego",
	Short: "A simple golang version manager",
	Long:  "A simple golang version manager",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.GlobalConfig.Verbose = verbose
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.Version = appversion
	rootCmd.SetVersionTemplate("updatego {{.Version}}\n")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(latestCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(toolchainCmd)
	rootCmd.AddCommand(syncCmd)
}

var (
	installVersion    string
	installForce      bool
	installGlobal     bool
	installUser       bool
	installCustomPath string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Go",
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().StringVar(&installVersion, "version", "", "Specify Go version to install")
	installCmd.Flags().BoolVar(&installForce, "force", false, "Force install by moving existing Go version")
	installCmd.Flags().BoolVar(&installGlobal, "global", false, "Install globally")
	installCmd.Flags().BoolVar(&installUser, "user", false, "Install for the user")
	installCmd.Flags().StringVar(&installCustomPath, "custom-path", "", "Install to a custom path")
}

func runInstall(cmd *cobra.Command, args []string) error {
	return installer.InstallGo(installVersion, installForce, installGlobal, installUser, installCustomPath)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Go installation status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		return local.CheckGoStatus()
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Print the latest Go version available",
	RunE: func(cmd *cobra.Command, args []string) error {
		return local.PrintLatestGoVersion()
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Go to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		return update.Go()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all local Go versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		return local.ListLocalVersions()
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch [VERSION]",
	Short: "Switch to a specific Go version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		version := args[0]
		if err := local.SwitchGoVersion(version); err != nil {
			return err
		}
		return local.ApplySwitchToolchain(version)
	},
}

var toolchainCmd = &cobra.Command{
	Use:   "toolchain [auto|local]",
	Short: "Show or set GOTOOLCHAIN via the managed Go toolchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		mode := ""
		if len(args) > 0 {
			mode = args[0]
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}
		return local.RunToolchain(mode)
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Set GOTOOLCHAIN to the active managed Go version (goX.Y.Z)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadIntoGlobal(); err != nil {
			return err
		}
		return local.SyncGOTOOLCHAIN()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
