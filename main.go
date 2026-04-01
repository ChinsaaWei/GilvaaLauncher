package main

import (
	"GilvaaLauncher/cmd"
	"GilvaaLauncher/logger"
)

func main() {
	cmd.InitLogger()
	defer logger.Close()

	if err := cmd.Execute(); err != nil {
		logger.Fatal("%v", err)
	}
}
