package cmd

import (
	"github.com/ChinsaaWei/HiraaLib/launch"
	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/modloader"
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch <version> [username]",
	Short: "Launch Minecraft",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		versionID := args[0]
		user := username
		if len(args) > 1 {
			user = args[1]
		}

		cfg := loadConfig()
		cfg.Username = user

		mlm := modloader.NewModLoaderManager()
		l := launch.NewLauncher(cfg, nil, mlm)

		if err := l.Launch(versionID, user, serverAddr, serverPort); err != nil {
			logger.Fatal("Failed to launch Minecraft: %v", err)
		}
	},
}