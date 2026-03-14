package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	JavaPath        string
	MinMemory       int
	MaxMemory       int
	NewGenMemory    int
	Username        string
	UUID            string
	AccessToken     string
	Width           int
	Height          int
	FullScreen      bool
	GameDir         string
	WorkingDir      string
	LogDir          string
	DownloadDir     string
	LauncherVersion string
}

func NewConfig() *Config {
	// 获取可执行文件所在目录
	execDir, err := getExecutableDir()
	if err != nil {
		execDir = "."
	}

	// 游戏目录放在可执行文件同目录下的.minecraft文件夹
	gameDir := filepath.Join(execDir, ".minecraft")
	workingDir := filepath.Join(execDir, ".mclauncher")
	logDir := filepath.Join(workingDir, "logs")
	downloadDir := filepath.Join(workingDir, "downloads")

	return &Config{
		JavaPath:        "java",
		MinMemory:       1024,
		MaxMemory:       4096,
		NewGenMemory:    256,
		Username:        "Player",
		UUID:            "00000000-0000-0000-0000-000000000000",
		AccessToken:     "offline",
		Width:           1920,
		Height:          1080,
		FullScreen:      false,
		GameDir:         gameDir,
		WorkingDir:      workingDir,
		LogDir:          logDir,
		DownloadDir:     downloadDir,
		LauncherVersion: "1.0.0",
	}
}

func getExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// 获取可执行文件所在目录
	execDir := filepath.Dir(exePath)

	// 如果是符号链接，解析到实际路径
	realPath, err := filepath.EvalSymlinks(exePath)
	if err == nil {
		execDir = filepath.Dir(realPath)
	}

	return execDir, nil
}

func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.GameDir,
		c.WorkingDir,
		c.LogDir,
		c.DownloadDir,
		filepath.Join(c.GameDir, "versions"),
		filepath.Join(c.GameDir, "libraries"),
		filepath.Join(c.GameDir, "assets"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (c *Config) GetVersionDir(version string) string {
	return filepath.Join(c.GameDir, "versions", version)
}

func (c *Config) GetVersionJarPath(version string) string {
	return filepath.Join(c.GetVersionDir(version), version+".jar")
}

func (c *Config) GetVersionJsonPath(version string) string {
	return filepath.Join(c.GetVersionDir(version), version+".json")
}

func (c *Config) GetNativesDir(version string) string {
	return filepath.Join(c.GetVersionDir(version), "natives")
}

func (c *Config) GetLibrariesDir() string {
	return filepath.Join(c.GameDir, "libraries")
}

func (c *Config) GetAssetsDir() string {
	return filepath.Join(c.GameDir, "assets")
}
