package main

import (
	"GilvaaLauncher/cmd"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		cmd.StartTUI()
		return
	}

	cmd.InitLogger()
	defer func() {
		if cmd.LoggerEnabled() {
			cmd.CloseLogger()
		}
	}()

	if err := cmd.Execute(); err != nil {
		cmd.FatalLog("%v", err)
	}
}
