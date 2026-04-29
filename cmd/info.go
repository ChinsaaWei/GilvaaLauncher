package cmd

import (
	"fmt"

	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/version"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <version>",
	Short: "Show information about a Minecraft version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		versionID := args[0]

		cfg := loadConfig()
		vm := version.NewManager(cfg.GameDir, nil)

		installed := vm.IsVersionInstalled(versionID)
		fmt.Printf("Version: %s\n", versionID)
		fmt.Printf("Installed: %v\n", installed)

		if installed {
			versionInfo, err := vm.GetVersionInfo(versionID)
			if err != nil {
				logger.Fatal("Failed to get version info: %v", err)
			}

			fmt.Printf("Type: %s\n", versionInfo.Type)
			fmt.Printf("Main Class: %s\n", versionInfo.MainClass)
			fmt.Printf("Release Time: %s\n", versionInfo.ReleaseTime)
			fmt.Printf("Libraries: %d\n", len(versionInfo.Libraries))
			fmt.Printf("Assets: %s\n", versionInfo.Assets)

			if versionInfo.JavaVersion != nil {
				fmt.Printf("Required Java: %d+\n", versionInfo.JavaVersion.MajorVersion)
			}
		}
	},
}