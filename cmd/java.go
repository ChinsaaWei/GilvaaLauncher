package cmd

import (
	"fmt"

	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/util"
	"github.com/spf13/cobra"
)

var javaCmd = &cobra.Command{
	Use:   "java",
	Short: "Java management commands",
}

var javaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Java versions",
	Run: func(cmd *cobra.Command, args []string) {
		javaInfos, err := util.FindJava()
		if err != nil {
			logger.Fatal("Failed to find Java installations: %v", err)
		}

		if len(javaInfos) == 0 {
			fmt.Println("No Java installations found")
			return
		}

		fmt.Printf("Found %d Java installation(s):\n", len(javaInfos))
		for _, info := range javaInfos {
			marker := ""
			if info.Major >= 17 {
				marker = " (recommended)"
			}
			fmt.Printf("  Java %d%s\n", info.Major, marker)
			fmt.Printf("    Path: %s\n", info.Path)
			fmt.Printf("    Version: %s\n", info.Version)
			fmt.Printf("    Arch: %s\n", info.Arch)
			fmt.Println()
		}
	},
}

func init() {
	javaCmd.AddCommand(javaListCmd)
}