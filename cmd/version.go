package cmd

import (
	"fmt"

	"github.com/ChinsaaWei/HiraaLib/download"
	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage Minecraft versions",
}

var listCmd = &cobra.Command{
	Use:   "list [type]",
	Short: "List available Minecraft versions",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		versionType := ""
		if len(args) > 0 {
			versionType = args[0]
		}

		cfg := loadConfig()
		dl := download.NewDownloader()
		vd := download.NewVersionDownloader(dl, cfg.DownloadDir)
		vm := version.NewManager(cfg.GameDir, vd)

		versions, err := vm.ListAvailableVersions(versionType)
		if err != nil {
			logger.Fatal("Failed to list versions: %v", err)
		}

		fmt.Printf("Available Minecraft versions (%d):\n", len(versions))
		for _, v := range versions {
			installed := "  "
			if vm.IsVersionInstalled(v.ID) {
				installed = "* "
			}
			fmt.Printf("%s%-15s %s (released: %s)\n", installed, v.ID, v.Type, v.ReleaseTime[:10])
		}
	},
}

var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List installed Minecraft versions",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		vm := version.NewManager(cfg.GameDir, nil)

		versions, err := vm.ListInstalledVersions()
		if err != nil {
			logger.Fatal("Failed to list installed versions: %v", err)
		}

		if len(versions) == 0 {
			fmt.Println("No versions installed")
			return
		}

		fmt.Printf("Installed versions (%d):\n", len(versions))
		for _, v := range versions {
			fmt.Printf("  %s\n", v)
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Minecraft version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		versionID := args[0]

		cfg := loadConfig()
		dl := download.NewDownloader()
		vd := download.NewVersionDownloader(dl, cfg.DownloadDir)
		vm := version.NewManager(cfg.GameDir, vd)

		if _, err := vm.InstallVersion(versionID); err != nil {
			logger.Fatal("Failed to install version %s: %v", versionID, err)
		}

		fmt.Printf("Version %s installed successfully\n", versionID)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Uninstall a Minecraft version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		versionID := args[0]

		cfg := loadConfig()
		vm := version.NewManager(cfg.GameDir, nil)

		if err := vm.UninstallVersion(versionID); err != nil {
			logger.Fatal("Failed to uninstall version %s: %v", versionID, err)
		}

		fmt.Printf("Version %s uninstalled successfully\n", versionID)
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for Minecraft versions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		cfg := loadConfig()
		dl := download.NewDownloader()
		vd := download.NewVersionDownloader(dl, cfg.DownloadDir)
		vm := version.NewManager(cfg.GameDir, vd)

		versions, err := vm.SearchVersions(query)
		if err != nil {
			logger.Fatal("Failed to search versions: %v", err)
		}

		if len(versions) == 0 {
			fmt.Printf("No versions found matching '%s'\n", query)
			return
		}

		fmt.Printf("Found %d versions matching '%s':\n", len(versions), query)
		for _, v := range versions {
			fmt.Printf("  %-15s %s (released: %s)\n", v.ID, v.Type, v.ReleaseTime[:10])
		}
	},
}

func init() {
	versionCmd.AddCommand(listCmd)
	versionCmd.AddCommand(installedCmd)
	versionCmd.AddCommand(installCmd)
	versionCmd.AddCommand(uninstallCmd)
	versionCmd.AddCommand(searchCmd)
}
