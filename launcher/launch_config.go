package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LaunchConfig struct {
	JavaPath         string
	MinMemory        int
	MaxMemory        int
	NewGenMemory     int
	MainClass        string
	Classpath        []string
	GameDir          string
	AssetsDir        string
	AssetsIndex      string
	Version          string
	VersionType      string
	Username         string
	UUID             string
	AccessToken      string
	UserType         string
	Width            int
	Height           int
	FullScreen       bool
	JVMArgs          []string
	GameArgs         []string
	NativesDir       string
	ServerAddress    string
	ServerPort       int
	Demo             bool
	Log4jConfig      string
	LauncherBrand    string
	LauncherVersion  string
	SystemProperties map[string]string
}

func NewLaunchConfig() *LaunchConfig {
	return &LaunchConfig{
		JavaPath:         "java",
		MinMemory:        512,
		MaxMemory:        2048,
		NewGenMemory:     128,
		MainClass:        "net.minecraft.client.main.Main",
		VersionType:      "release",
		UserType:         "mojang",
		Width:            854,
		Height:           480,
		LauncherBrand:    "custom-launcher",
		LauncherVersion:  "1.0.0",
		SystemProperties: make(map[string]string),
	}
}

func (c *LaunchConfig) BuildClasspath() string {
	paths := make([]string, len(c.Classpath))
	for i, p := range c.Classpath {
		paths[i] = filepath.ToSlash(p)
	}
	return strings.Join(paths, string(filepath.ListSeparator))
}

func (c *LaunchConfig) BuildJVMArgs() []string {
	args := []string{}

	args = append(args, "-XX:HeapDumpPath=MojangTricksIntelDriversForPerformance_javaw.exe_minecraft.exe.heapdump")

	osName := runtime.GOOS
	if osName == "windows" {
		osName = "Windows"
	} else if osName == "darwin" {
		osName = "Mac OS X"
	} else if osName == "linux" {
		osName = "Linux"
	}

	osVersion := os.Getenv("OSVERSION")
	if osVersion == "" {
		osVersion = "10.0"
	}

	args = append(args, fmt.Sprintf("-Dos.name=%s", osName))
	args = append(args, fmt.Sprintf("-Dos.version=%s", osVersion))

	if c.MinMemory > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", c.MinMemory))
	}
	if c.MaxMemory > 0 {
		args = append(args, fmt.Sprintf("-Xmx%dM", c.MaxMemory))
	}
	if c.NewGenMemory > 0 {
		args = append(args, fmt.Sprintf("-Xmn%dM", c.NewGenMemory))
	}

	if c.NativesDir != "" {
		args = append(args, fmt.Sprintf("-Djava.library.path=%s", filepath.ToSlash(c.NativesDir)))
	}

	args = append(args, fmt.Sprintf("-Dminecraft.launcher.brand=%s", c.LauncherBrand))
	args = append(args, fmt.Sprintf("-Dminecraft.launcher.version=%s", c.LauncherVersion))

	if c.Log4jConfig != "" {
		args = append(args, fmt.Sprintf("-Dlog4j.configurationFile=%s", filepath.ToSlash(c.Log4jConfig)))
	}

	for key, value := range c.SystemProperties {
		args = append(args, fmt.Sprintf("-D%s=%s", key, value))
	}

	args = append(args, "-XX:+UseG1GC")
	args = append(args, "-XX:+UnlockExperimentalVMOptions")
	args = append(args, "-XX:G1NewSizePercent=20")
	args = append(args, "-XX:G1ReservePercent=20")
	args = append(args, "-XX:MaxGCPauseMillis=50")
	args = append(args, "-XX:G1HeapRegionSize=32M")
	args = append(args, "-XX:-UseAdaptiveSizePolicy")
	args = append(args, "-XX:-OmitStackTraceInFastThrow")

	args = append(args, c.JVMArgs...)

	args = append(args, "-cp", c.BuildClasspath())

	return args
}

func (c *LaunchConfig) BuildGameArgs() []string {
	args := []string{}

	if c.Username != "" {
		args = append(args, "--username", c.Username)
	}
	if c.UUID != "" {
		args = append(args, "--uuid", c.UUID)
	}
	if c.AccessToken != "" {
		args = append(args, "--accessToken", c.AccessToken)
	}
	if c.Version != "" {
		args = append(args, "--version", c.Version)
	}
	if c.GameDir != "" {
		args = append(args, "--gameDir", filepath.ToSlash(c.GameDir))
	}
	if c.AssetsDir != "" {
		args = append(args, "--assetsDir", filepath.ToSlash(c.AssetsDir))
	}
	if c.AssetsIndex != "" {
		args = append(args, "--assetIndex", c.AssetsIndex)
	}
	if c.UserType != "" {
		args = append(args, "--userType", c.UserType)
	}
	if c.VersionType != "" {
		args = append(args, "--versionType", c.VersionType)
	}

	if c.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", c.Width))
	}
	if c.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", c.Height))
	}

	if c.FullScreen {
		args = append(args, "--fullscreen")
	}

	if c.ServerAddress != "" {
		args = append(args, "--server", c.ServerAddress)
		if c.ServerPort > 0 {
			args = append(args, "--port", fmt.Sprintf("%d", c.ServerPort))
		}
	}

	if c.Demo {
		args = append(args, "--demo")
	}

	args = append(args, c.GameArgs...)

	return args
}

func (c *LaunchConfig) BuildCommand() []string {
	cmd := []string{c.JavaPath}

	cmd = append(cmd, c.BuildJVMArgs()...)

	cmd = append(cmd, c.MainClass)

	cmd = append(cmd, c.BuildGameArgs()...)

	return cmd
}

func (c *LaunchConfig) BuildCommandString() string {
	parts := c.BuildCommand()
	quoted := make([]string, len(parts))
	for i, part := range parts {
		if strings.Contains(part, " ") || strings.Contains(part, "\"") {
			quoted[i] = fmt.Sprintf("\"%s\"", strings.ReplaceAll(part, "\"", "\\\""))
		} else {
			quoted[i] = part
		}
	}
	return strings.Join(quoted, " ")
}

func (c *LaunchConfig) AddLibrary(path string) {
	c.Classpath = append(c.Classpath, path)
}

func (c *LaunchConfig) AddJVMArg(arg string) {
	c.JVMArgs = append(c.JVMArgs, arg)
}

func (c *LaunchConfig) AddGameArg(arg string) {
	c.GameArgs = append(c.GameArgs, arg)
}

func (c *LaunchConfig) SetSystemProperty(key, value string) {
	c.SystemProperties[key] = value
}

func (c *LaunchConfig) GetCommandLineForWMIC() string {
	return c.BuildCommandString()
}

func (c *LaunchConfig) ExportToBatchFile(filename string) error {
	cmdStr := c.BuildCommandString()
	return os.WriteFile(filename, []byte(cmdStr), 0644)
}
