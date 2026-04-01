package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"GilvaaLauncher/config"
	"GilvaaLauncher/downloader"
	"GilvaaLauncher/launcher"
	"GilvaaLauncher/logger"
	"GilvaaLauncher/modloader"
	"GilvaaLauncher/util"
	"GilvaaLauncher/version"

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

var rootCmd = &cobra.Command{
	Use:   "GilvaaLauncher",
	Short: "Minecraft Launcher - Command Line Tool",
	Long:  `A command-line Minecraft launcher that supports version management, downloading, and launching Minecraft.`,
}

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
		dl := downloader.NewDownloader()
		vd := downloader.NewVersionDownloader(dl, cfg.DownloadDir)
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
		dl := downloader.NewDownloader()
		vd := downloader.NewVersionDownloader(dl, cfg.DownloadDir)
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
		dl := downloader.NewDownloader()
		vd := downloader.NewVersionDownloader(dl, cfg.DownloadDir)
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
		l := launcher.NewLauncher(cfg, nil, mlm)

		if err := l.Launch(versionID, user, serverAddr, serverPort); err != nil {
			logger.Fatal("Failed to launch Minecraft: %v", err)
		}
	},
}

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
		l := launcher.NewLauncher(cfg, nil, mlm)

		cmdArgs, err := l.GetLaunchCommand(versionID, user, serverAddr, serverPort)
		if err != nil {
			logger.Fatal("Failed to get launch command: %v", err)
		}

		fmt.Println("Launch Command:")
		fmt.Println(strings.Join(cmdArgs, " "))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&javaPath, "java", "", "Java executable path")
	rootCmd.PersistentFlags().StringVar(&memory, "memory", "4G", "memory allocation (e.g., 2G, 4G)")
	rootCmd.PersistentFlags().StringVar(&gameDir, "game-dir", "", "Minecraft game directory")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "Player", "username")
	rootCmd.PersistentFlags().IntVar(&width, "width", 1920, "window width")
	rootCmd.PersistentFlags().IntVar(&height, "height", 1080, "window height")
	rootCmd.PersistentFlags().BoolVar(&fullScreen, "fullscreen", false, "fullscreen mode")
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "", "server address")
	rootCmd.PersistentFlags().IntVar(&serverPort, "port", 25565, "server port")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(launchCmd)

	rootCmd.AddCommand(javaCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(commandCmd)

	versionCmd.AddCommand(listCmd)
	versionCmd.AddCommand(installedCmd)
	versionCmd.AddCommand(installCmd)
	versionCmd.AddCommand(uninstallCmd)
	versionCmd.AddCommand(searchCmd)

	javaCmd.AddCommand(javaListCmd)
}

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

func Execute() error {
	return rootCmd.Execute()
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

	logger.Info("Minecraft Launcher v%s starting...", cfg.LauncherVersion)
	logger.Info("OS: %s %s", runtime.GOOS, runtime.GOARCH)
}
