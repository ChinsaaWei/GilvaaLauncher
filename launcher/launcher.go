package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"mclauncher/config"
	"mclauncher/downloader"
	"mclauncher/logger"
	"mclauncher/modloader"
	"mclauncher/util"
	"mclauncher/version"
)

type Launcher struct {
	config       *config.Config
	modLoaderMgr *modloader.ModLoaderManager
}

func NewLauncher(cfg *config.Config, versionMgr *downloader.VersionDownloader, modLoaderMgr *modloader.ModLoaderManager) *Launcher {
	return &Launcher{
		config:       cfg,
		modLoaderMgr: modLoaderMgr,
	}
}

func (l *Launcher) Launch(versionID string, username string, serverAddress string, serverPort int) error {
	logger.Info("Launching Minecraft %s as %s", versionID, username)

	dl := downloader.NewDownloader()
	vd := downloader.NewVersionDownloader(dl, l.config.DownloadDir)
	vm := version.NewManager(l.config.GameDir, vd)

	if _, err := vm.InstallVersion(versionID); err != nil {
		return fmt.Errorf("failed to download version: %w", err)
	}

	versionInfo, err := vm.GetVersionInfo(versionID)
	if err != nil {
		return fmt.Errorf("failed to get version info: %w", err)
	}

	versionGameDir := l.config.GetVersionGameDir(versionID)
	if err := os.MkdirAll(versionGameDir, 0755); err != nil {
		return fmt.Errorf("failed to create version game directory: %w", err)
	}
	logger.Info("Using version game directory: %s", versionGameDir)

	javaVersion := downloader.GetJavaVersion(versionInfo)
	if err := util.ValidateJavaVersion(l.config.JavaPath, javaVersion); err != nil {
		logger.Warn("Java validation warning: %v", err)
	}

	nativesDir := l.config.GetNativesDir(versionID)
	librariesDir := l.config.GetLibrariesDir()

	assetsDownloader := downloader.NewAssetsDownloader(dl, l.config.GetAssetsDir())
	librariesDownloader := downloader.NewLibrariesDownloader(dl, librariesDir)

	if err := assetsDownloader.DownloadAssets(versionInfo.AssetIndex, true); err != nil {
		return fmt.Errorf("failed to download assets: %w", err)
	}

	libraries := downloader.GetLibraries(versionInfo)
	classpath, err := librariesDownloader.DownloadLibraries(libraries, nativesDir)
	if err != nil {
		return fmt.Errorf("failed to download libraries: %w", err)
	}

	clientJarPath := l.config.GetVersionJarPath(versionID)
	if _, err := os.Stat(clientJarPath); os.IsNotExist(err) {
		return fmt.Errorf("client JAR not found: %s", clientJarPath)
	}
	classpath = append(classpath, clientJarPath)

	launchConfig := l.buildLaunchConfig(versionInfo, classpath, nativesDir, username, serverAddress, serverPort)

	cmd := l.buildCommand(launchConfig)
	logger.Info("Starting Minecraft process...")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Minecraft: %w", err)
	}

	logger.Info("Minecraft started successfully (PID: %d)", cmd.Process.Pid)
	return nil
}

func (l *Launcher) buildLaunchConfig(versionInfo *downloader.VersionInfo, classpath []string, nativesDir, username, serverAddress string, serverPort int) *LaunchConfig {
	cfg := NewLaunchConfig()

	cfg.JavaPath = l.config.JavaPath
	cfg.MinMemory = l.config.MinMemory
	cfg.MaxMemory = l.config.MaxMemory
	cfg.NewGenMemory = l.config.NewGenMemory
	cfg.MainClass = downloader.GetMainClass(versionInfo)
	cfg.Classpath = classpath

	cfg.GameDir = l.config.GetVersionGameDir(versionInfo.ID)
	cfg.AssetsDir = l.config.GetAssetsDir()
	cfg.AssetsIndex = downloader.GetAssetIndex(versionInfo)
	cfg.Version = versionInfo.ID
	cfg.VersionType = versionInfo.Type

	cfg.Username = username
	cfg.UUID = l.config.UUID
	cfg.AccessToken = l.config.AccessToken
	cfg.UserType = "legacy"

	cfg.Width = l.config.Width
	cfg.Height = l.config.Height
	cfg.FullScreen = l.config.FullScreen

	cfg.NativesDir = nativesDir
	cfg.ServerAddress = serverAddress
	cfg.ServerPort = serverPort

	cfg.LauncherBrand = "mclauncher"
	cfg.LauncherVersion = l.config.LauncherVersion

	cfg.SetSystemProperty("fml.ignoreInvalidMinecraftCertificates", "true")
	cfg.SetSystemProperty("fml.ignorePatchDiscrepancies", "true")

	return cfg
}

func (l *Launcher) buildCommand(cfg *LaunchConfig) *exec.Cmd {
	args := cfg.BuildCommand()

	cmd := exec.Command(args[0], args[1:]...)

	return cmd
}



func (l *Launcher) GetLaunchCommand(versionID, username string, serverAddress string, serverPort int) ([]string, error) {
	dl := downloader.NewDownloader()
	vd := downloader.NewVersionDownloader(dl, l.config.DownloadDir)
	vm := version.NewManager(l.config.GameDir, vd)

	versionInfo, err := vm.GetVersionInfo(versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version info: %w", err)
	}

	nativesDir := l.config.GetNativesDir(versionID)

	classpath := []string{}

	clientJarPath := l.config.GetVersionJarPath(versionID)
	classpath = append(classpath, clientJarPath)

	launchConfig := l.buildLaunchConfig(versionInfo, classpath, nativesDir, username, serverAddress, serverPort)
	return launchConfig.BuildCommand(), nil
}
