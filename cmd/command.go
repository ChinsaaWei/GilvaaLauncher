package cmd

import (
	"fmt"
	"strings"

	"github.com/ChinsaaWei/HiraaLib/launch"
	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/modloader"
	"github.com/spf13/cobra"
)

var commandCmd = &cobra.Command{
	Use:   "command <version> [username]",
	Short: "Print the launch command for a Minecraft version",
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

		cmdArgs, err := l.GetLaunchCommand(versionID, user, serverAddr, serverPort)
		if err != nil {
			logger.Fatal("Failed to get launch command: %v", err)
		}

		fmt.Println("Launch Command:")
		fmt.Println(strings.Join(cmdArgs, " "))
	},
}