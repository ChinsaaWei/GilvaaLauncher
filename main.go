package main

import (
	"mclauncher/cmd"
	"mclauncher/logger"
)

func main() {
	cmd.InitLogger()
	defer logger.Close()

	if err := cmd.Execute(); err != nil {
		logger.Fatal("%v", err)
	}
}
