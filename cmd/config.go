package cmd

import (
	"fmt"
	"strings"

	"github.com/ChinsaaWei/HiraaLib/config"
	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/util"
)

func loadConfig() *config.Config {
	cfg := config.NewConfig()

	if javaPath != "" {
		cfg.JavaPath = javaPath
	} else {
		if bestJava, err := util.FindBestJava(17); err == nil {
			cfg.JavaPath = bestJava
		}
	}

	if gameDir != "" {
		cfg.GameDir = gameDir
	}

	cfg.Username = username
	cfg.Width = width
	cfg.Height = height
	cfg.FullScreen = fullScreen

	if memory != "" {
		if strings.HasSuffix(memory, "G") || strings.HasSuffix(memory, "g") {
			var mem int
			fmt.Sscanf(memory, "%d", &mem)
			cfg.MaxMemory = mem * 1024
			cfg.MinMemory = mem / 2 * 1024
		} else if strings.HasSuffix(memory, "M") || strings.HasSuffix(memory, "m") {
			var mem int
			fmt.Sscanf(memory, "%d", &mem)
			cfg.MaxMemory = mem
			cfg.MinMemory = mem / 2
		}
	}

	if err := cfg.EnsureDirectories(); err != nil {
		logger.Warn("Failed to ensure directories: %v", err)
	}

	return cfg
}
