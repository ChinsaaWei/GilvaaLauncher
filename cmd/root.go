package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"GilvaaLauncher/tui"

	"github.com/ChinsaaWei/HiraaLib/config"
	"github.com/ChinsaaWei/HiraaLib/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	verbose    bool
	logLevel   string
	javaPath   string
	memory     string
	gameDir    string
	username   string
	width      int
	height     int
	fullScreen bool
	serverAddr string
	serverPort int
)

var loggerInitialized bool

var rootCmd = &cobra.Command{
	Use:   "GilvaaLauncher",
	Short: "Minecraft Launcher - Command Line Tool",
	Long:  `A command-line Minecraft launcher that supports version management, downloading, and launching Minecraft.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&javaPath, "java", "", "Java executable path")
	rootCmd.PersistentFlags().StringVar(&memory, "memory", "4G", "memory allocation (e.g., 2G, 4G)")
	rootCmd.PersistentFlags().StringVar(&gameDir, "game-dir", "", "Minecraft game directory")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "Player", "username")
	rootCmd.PersistentFlags().IntVar(&width, "width", 1920, "window width (min: 800, max: 3840)")
	rootCmd.PersistentFlags().IntVar(&height, "height", 1080, "window height (min: 600, max: 2160)")
	rootCmd.PersistentFlags().BoolVar(&fullScreen, "fullscreen", false, "fullscreen mode")
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "", "server address")
	rootCmd.PersistentFlags().IntVar(&serverPort, "port", 25565, "server port")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(javaCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(commandCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func StartTUI() {
	tui.Start()
}

func LoggerEnabled() bool {
	return loggerInitialized
}

func CloseLogger() {
	if loggerInitialized {
		logger.Close()
	}
}

func FatalLog(format string, args ...interface{}) {
	if loggerInitialized {
		logger.Fatal(format, args...)
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		os.Exit(1)
	}
}

func InitLogger() {
	level := logger.INFO
	if verbose {
		level = logger.DEBUG
	}

	if logLevel != "" {
		switch strings.ToLower(logLevel) {
		case "debug":
			level = logger.DEBUG
		case "info":
			level = logger.INFO
		case "warn":
			level = logger.WARN
		case "error":
			level = logger.ERROR
		}
	}

	cfg := config.NewConfig()
	if err := logger.Init(level, cfg.LogDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	loggerInitialized = true

	logger.Info("Minecraft Launcher v%s starting...", cfg.LauncherVersion)
	logger.Info("OS: %s %s", runtime.GOOS, runtime.GOARCH)
}
