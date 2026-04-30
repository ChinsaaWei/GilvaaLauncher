package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ChinsaaWei/HiraaLib/config"
	"github.com/ChinsaaWei/HiraaLib/download"
	"github.com/ChinsaaWei/HiraaLib/launch"
	"github.com/ChinsaaWei/HiraaLib/logger"
	"github.com/ChinsaaWei/HiraaLib/modloader"
	"github.com/ChinsaaWei/HiraaLib/util"
	"github.com/ChinsaaWei/HiraaLib/version"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewType int

const (
	LaunchView ViewType = iota
	VersionView
	ConfigView
	JavaView
	CommandView
	InfoView
)

var viewNames = []string{
	"🚀 启动",
	"📦 版本管理",
	"⚙️ 配置",
	"☕ Java管理",
	"📋 命令预览",
	"ℹ️ 版本信息",
}

type model struct {
	currentView ViewType

	launchState  launchState
	versionState versionState
	configState  configState
	javaState    javaState
	commandState commandState
	infoState    infoState

	focusedField  int
	statusMessage string
	statusType    string
	showHelp      bool
	loading       bool
	err           error
	width         int
	height        int
}

type launchState struct {
	versions        []string
	selectedVersion string
	username        string
	memory          string
	javaPath        string
	gameDir         string
	width           int
	height          int
	fullScreen      bool
	serverAddr      string
	serverPort      int
}

type versionState struct {
	availableVersions []versionInfo
	installedVersions []string
	selectedIndex     int
	filterType        string
	details           versionDetail
	showDetails       bool
}

type versionInfo struct {
	ID          string
	Type        string
	ReleaseTime string
	Installed   bool
}

type versionDetail struct {
	MainClass   string
	Type        string
	ReleaseTime string
	Libraries   int
	Assets      string
	JavaVersion int
}

type configState struct {
	configPath string
	logLevel   string
	verbose    bool
	gameDir    string
	javaPath   string
	memory     string
	width      int
	height     int
	fullScreen bool
	modified   bool
}

type javaState struct {
	javaList      []javaInfo
	selectedIndex int
}

type javaInfo struct {
	Path    string
	Version string
	Major   int
	Arch    string
}

type commandState struct {
	command string
}

type infoState struct {
	versionID string
	details   versionDetail
	loading   bool
}

func NewModel() *model {
	cfg := loadConfig()
	m := &model{
		currentView: LaunchView,
		launchState: launchState{
			versions:        []string{},
			selectedVersion: "1.20.1",
			username:        cfg.Username,
			memory:          fmt.Sprintf("%dM", cfg.MaxMemory),
			javaPath:        cfg.JavaPath,
			gameDir:         cfg.GameDir,
			width:           cfg.Width,
			height:          cfg.Height,
			fullScreen:      cfg.FullScreen,
			serverPort:      25565,
		},
		configState: configState{
			configPath: cfg.WorkingDir,
			logLevel:   "info",
			verbose:    false,
			gameDir:    cfg.GameDir,
			javaPath:   cfg.JavaPath,
			memory:     fmt.Sprintf("%dM", cfg.MaxMemory),
			width:      cfg.Width,
			height:     cfg.Height,
			fullScreen: cfg.FullScreen,
		},
		javaState: javaState{
			javaList: []javaInfo{},
		},
		statusMessage: "就绪",
	}

	m.loadVersions()
	m.loadJavaList()
	return m
}

func loadConfig() *config.Config {
	cfg := config.NewConfig()
	return cfg
}

var (
	borderStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	headerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("227")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	greenStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	redStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	yellowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Background(lipgloss.Color("235"))
	buttonStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("76")).Padding(0, 2)
	buttonDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Padding(0, 2)
	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("235"))
)

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.showHelp {
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "ctrl+c" {
				if msg.String() == "q" || msg.String() == "ctrl+c" {
					m.quit()
				}
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+q", "ctrl+c":
			m.quit()
			return m, tea.Quit

		case "ctrl+s":
			m.saveConfig()
			return m, nil

		case "f1":
			m.showHelp = true
			return m, nil

		case "tab":
			m.nextView()
			return m, nil

		case "shift+tab":
			m.prevView()
			return m, nil

		case "1", "2", "3", "4", "5", "6":
			viewNum := int(msg.String()[0] - '1')
			if viewNum >= 0 && viewNum <= 5 {
				m.currentView = ViewType(viewNum)
				m.focusedField = 0
			}
			return m, nil
		}

		m.handleViewInput(msg)
	}
	return m, nil
}

func (m *model) handleViewInput(msg tea.KeyMsg) {
	key := msg.String()

	switch m.currentView {
	case LaunchView:
		m.handleLaunchInput(key)
	case VersionView:
		m.handleVersionInput(key)
	case ConfigView:
		m.handleConfigInput(key)
	case JavaView:
		m.handleJavaInput(key)
	case CommandView:
		m.handleCommandInput(key)
	case InfoView:
		m.handleInfoInput(key)
	}
}

func (m *model) handleLaunchInput(key string) {
	switch key {
	case "up", "k":
		if m.focusedField > 0 {
			m.focusedField--
		}
	case "down", "j":
		if m.focusedField < 8 {
			m.focusedField++
		}
	case "left", "h":
		m.handleListNav(-1)
	case "right", "l":
		m.handleListNav(1)
	case "enter":
		if m.focusedField == 8 {
			m.launchGame()
		}
	case " ":
		m.handleToggle()
	case "backspace":
		m.handleBackspace()
	default:
		if len(key) == 1 && m.focusedField >= 2 && m.focusedField <= 6 {
			m.handleCharInput(key)
		}
	}
}

func (m *model) handleVersionInput(key string) {
	switch key {
	case "up", "k":
		if m.versionState.selectedIndex > 0 {
			m.versionState.selectedIndex--
			m.updateVersionDetails()
		}
	case "down", "j":
		if m.versionState.selectedIndex < len(m.versionState.availableVersions)-1 {
			m.versionState.selectedIndex++
			m.updateVersionDetails()
		}
	case "enter":
		m.versionState.showDetails = !m.versionState.showDetails
	case "i":
		m.installVersion()
	case "d":
		m.deleteVersion()
	case "s":
		m.setDefaultVersion()
	}
}

func (m *model) handleConfigInput(key string) {
	switch key {
	case "up", "k":
		if m.focusedField > 0 {
			m.focusedField--
		}
	case "down", "j":
		if m.focusedField < 7 {
			m.focusedField++
		}
	case "enter":
		if m.focusedField == 6 {
			m.configState.fullScreen = !m.configState.fullScreen
		} else if m.focusedField == 2 {
			m.configState.verbose = !m.configState.verbose
		}
	case " ":
		m.handleToggle()
	default:
		if len(key) == 1 && m.focusedField >= 3 {
			m.handleCharInput(key)
		}
	}
}

func (m *model) handleJavaInput(key string) {
	switch key {
	case "up", "k":
		if m.javaState.selectedIndex > 0 {
			m.javaState.selectedIndex--
		}
	case "down", "j":
		if m.javaState.selectedIndex < len(m.javaState.javaList)-1 {
			m.javaState.selectedIndex++
		}
	case "enter":
		m.setDefaultJava()
	}
}

func (m *model) handleCommandInput(key string) {
	switch key {
	case "enter":
		m.launchGame()
	case "c":
		m.copyCommand()
	}
}

func (m *model) handleInfoInput(key string) {
	switch key {
	case "enter":
		m.updateInfoDetails()
	}
}

func (m *model) handleListNav(dir int) {
	versions := m.launchState.versions
	if len(versions) == 0 {
		return
	}

	currentIdx := -1
	for i, v := range versions {
		if v == m.launchState.selectedVersion {
			currentIdx = i
			break
		}
	}

	newIdx := currentIdx + dir
	if newIdx < 0 {
		newIdx = 0
	} else if newIdx >= len(versions) {
		newIdx = len(versions) - 1
	}

	m.launchState.selectedVersion = versions[newIdx]
	m.updateLaunchCommand()
}

func (m *model) handleToggle() {
	switch m.currentView {
	case LaunchView:
		if m.focusedField == 7 {
			m.launchState.fullScreen = !m.launchState.fullScreen
		}
	}
}

func (m *model) handleBackspace() {
	switch m.currentView {
	case LaunchView:
		m.handleFieldBackspace()
	case ConfigView:
		m.handleConfigBackspace()
	}
}

func (m *model) handleFieldBackspace() {
	switch m.focusedField {
	case 2:
		if len(m.launchState.username) > 0 {
			m.launchState.username = m.launchState.username[:len(m.launchState.username)-1]
		}
	case 3:
		if len(m.launchState.memory) > 0 {
			m.launchState.memory = m.launchState.memory[:len(m.launchState.memory)-1]
		}
	case 4:
		if len(m.launchState.javaPath) > 0 {
			m.launchState.javaPath = m.launchState.javaPath[:len(m.launchState.javaPath)-1]
		}
	case 5:
		if len(m.launchState.gameDir) > 0 {
			m.launchState.gameDir = m.launchState.gameDir[:len(m.launchState.gameDir)-1]
		}
	case 6:
		if len(m.launchState.serverAddr) > 0 {
			m.launchState.serverAddr = m.launchState.serverAddr[:len(m.launchState.serverAddr)-1]
		}
	}
}

func (m *model) handleConfigBackspace() {
	switch m.focusedField {
	case 3:
		if len(m.configState.gameDir) > 0 {
			m.configState.gameDir = m.configState.gameDir[:len(m.configState.gameDir)-1]
		}
	case 4:
		if len(m.configState.javaPath) > 0 {
			m.configState.javaPath = m.configState.javaPath[:len(m.configState.javaPath)-1]
		}
	case 5:
		if len(m.configState.memory) > 0 {
			m.configState.memory = m.configState.memory[:len(m.configState.memory)-1]
		}
	}
}

func (m *model) handleCharInput(char string) {
	switch m.currentView {
	case LaunchView:
		m.handleFieldCharInput(char)
	case ConfigView:
		m.handleConfigCharInput(char)
	}
}

func (m *model) handleFieldCharInput(char string) {
	switch m.focusedField {
	case 2:
		m.launchState.username += char
	case 3:
		m.launchState.memory += char
	case 4:
		m.launchState.javaPath += char
	case 5:
		m.launchState.gameDir += char
	}
}

func (m *model) handleConfigCharInput(char string) {
	switch m.focusedField {
	case 3:
		m.configState.gameDir += char
	case 4:
		m.configState.javaPath += char
	case 5:
		m.configState.memory += char
	}
}

func (m *model) nextView() {
	m.currentView = (m.currentView + 1) % 6
	m.focusedField = 0
}

func (m *model) prevView() {
	m.currentView = (m.currentView - 1 + 6) % 6
	m.focusedField = 0
}

func (m *model) loadVersions() {
	cfg := loadConfig()
	dl := download.NewDownloader()
	vd := download.NewVersionDownloader(dl, cfg.DownloadDir)
	vm := version.NewManager(cfg.GameDir, vd)

	versions, err := vm.ListAvailableVersions("")
	if err != nil {
		m.statusMessage = "加载版本失败"
		return
	}

	m.launchState.versions = []string{}
	m.versionState.availableVersions = []versionInfo{}

	for _, v := range versions {
		info := versionInfo{
			ID:          v.ID,
			Type:        v.Type,
			ReleaseTime: v.ReleaseTime,
			Installed:   vm.IsVersionInstalled(v.ID),
		}
		m.versionState.availableVersions = append(m.versionState.availableVersions, info)
		if info.Installed {
			m.launchState.versions = append(m.launchState.versions, v.ID)
		}
	}

	installedVersions, _ := vm.ListInstalledVersions()
	m.versionState.installedVersions = installedVersions

	if len(m.launchState.versions) > 0 && m.launchState.selectedVersion == "" {
		m.launchState.selectedVersion = m.launchState.versions[0]
	}

	m.updateLaunchCommand()
}

func (m *model) loadJavaList() {
	javaInfos, err := util.FindJava()
	if err != nil {
		return
	}

	m.javaState.javaList = []javaInfo{}
	for _, info := range javaInfos {
		java := javaInfo{
			Path:    info.Path,
			Version: info.Version,
			Major:   info.Major,
			Arch:    info.Arch,
		}
		m.javaState.javaList = append(m.javaState.javaList, java)
	}
}

func (m *model) updateVersionDetails() {
	if m.versionState.selectedIndex >= 0 && m.versionState.selectedIndex < len(m.versionState.availableVersions) {
		versionID := m.versionState.availableVersions[m.versionState.selectedIndex].ID
		m.fetchVersionDetails(versionID)
	}
}

func (m *model) fetchVersionDetails(versionID string) {
	cfg := loadConfig()
	vm := version.NewManager(cfg.GameDir, nil)

	info, err := vm.GetVersionInfo(versionID)
	if err != nil {
		return
	}

	detail := versionDetail{
		MainClass:   info.MainClass,
		Type:        info.Type,
		ReleaseTime: info.ReleaseTime,
		Libraries:   len(info.Libraries),
		Assets:      info.Assets,
	}
	if info.JavaVersion != nil {
		detail.JavaVersion = info.JavaVersion.MajorVersion
	}

	if m.currentView == VersionView {
		m.versionState.details = detail
	} else if m.currentView == InfoView {
		m.infoState.details = detail
	}
}

func (m *model) updateInfoDetails() {
	if m.infoState.versionID != "" {
		m.fetchVersionDetails(m.infoState.versionID)
	}
}

func (m *model) updateLaunchCommand() {
	cfg := loadConfig()
	cfg.Username = m.launchState.username
	cfg.JavaPath = m.launchState.javaPath
	cfg.GameDir = m.launchState.gameDir

	mlm := modloader.NewModLoaderManager()
	l := launch.NewLauncher(cfg, nil, mlm)

	cmdArgs, err := l.GetLaunchCommand(
		m.launchState.selectedVersion,
		m.launchState.username,
		m.launchState.serverAddr,
		m.launchState.serverPort,
	)
	if err != nil {
		m.commandState.command = ""
		return
	}

	m.commandState.command = strings.Join(cmdArgs, " \\\n")
}

func (m *model) installVersion() {
	if m.versionState.selectedIndex >= 0 && m.versionState.selectedIndex < len(m.versionState.availableVersions) {
		versionID := m.versionState.availableVersions[m.versionState.selectedIndex].ID
		if !m.versionState.availableVersions[m.versionState.selectedIndex].Installed {
			m.statusMessage = fmt.Sprintf("正在安装版本 %s...", versionID)
			go m.doInstall(versionID)
		}
	}
}

func (m *model) doInstall(versionID string) {
	cfg := loadConfig()
	dl := download.NewDownloader()
	vd := download.NewVersionDownloader(dl, cfg.DownloadDir)
	vm := version.NewManager(cfg.GameDir, vd)

	_, err := vm.InstallVersion(versionID)
	if err != nil {
		m.statusMessage = fmt.Sprintf("安装失败: %v", err)
		m.statusType = "error"
	} else {
		m.statusMessage = fmt.Sprintf("版本 %s 安装成功", versionID)
		m.statusType = "success"
		m.loadVersions()
	}
}

func (m *model) deleteVersion() {
	if m.versionState.selectedIndex >= 0 && m.versionState.selectedIndex < len(m.versionState.availableVersions) {
		versionID := m.versionState.availableVersions[m.versionState.selectedIndex].ID
		if m.versionState.availableVersions[m.versionState.selectedIndex].Installed {
			cfg := loadConfig()
			versionDir := cfg.GetVersionDir(versionID)
			if err := os.RemoveAll(versionDir); err != nil {
				m.statusMessage = fmt.Sprintf("删除失败: %v", err)
				m.statusType = "error"
			} else {
				m.statusMessage = fmt.Sprintf("版本 %s 已删除", versionID)
				m.statusType = "success"
				m.loadVersions()
			}
		}
	}
}

func (m *model) setDefaultVersion() {
	if m.versionState.selectedIndex >= 0 && m.versionState.selectedIndex < len(m.versionState.availableVersions) {
		versionID := m.versionState.availableVersions[m.versionState.selectedIndex].ID
		m.launchState.selectedVersion = versionID
		m.statusMessage = fmt.Sprintf("已设为启动版本: %s", versionID)
		m.currentView = LaunchView
	}
}

func (m *model) setDefaultJava() {
	if m.javaState.selectedIndex >= 0 && m.javaState.selectedIndex < len(m.javaState.javaList) {
		javaPath := m.javaState.javaList[m.javaState.selectedIndex].Path
		m.launchState.javaPath = javaPath
		m.configState.javaPath = javaPath
		m.statusMessage = fmt.Sprintf("已设为默认Java: %s", javaPath)
	}
}

func (m *model) launchGame() {
	m.statusMessage = "正在启动游戏..."
	m.loading = true

	go m.doLaunch()
}

func parseMemory(memStr string) int {
	memStr = strings.TrimSpace(memStr)
	if strings.HasSuffix(memStr, "G") || strings.HasSuffix(memStr, "g") {
		val := strings.TrimSuffix(memStr, "G")
		val = strings.TrimSuffix(val, "g")
		var mb int
		fmt.Sscanf(val, "%d", &mb)
		return mb * 1024
	}
	if strings.HasSuffix(memStr, "M") || strings.HasSuffix(memStr, "m") {
		val := strings.TrimSuffix(memStr, "M")
		val = strings.TrimSuffix(val, "m")
		var mb int
		fmt.Sscanf(val, "%d", &mb)
		return mb
	}
	var mb int
	fmt.Sscanf(memStr, "%d", &mb)
	return mb
}

func (m *model) doLaunch() {
	cfg := loadConfig()
	cfg.Username = m.launchState.username
	cfg.MaxMemory = parseMemory(m.launchState.memory)
	cfg.JavaPath = m.launchState.javaPath
	cfg.GameDir = m.launchState.gameDir
	cfg.Width = m.launchState.width
	cfg.Height = m.launchState.height
	cfg.FullScreen = m.launchState.fullScreen

	mlm := modloader.NewModLoaderManager()
	l := launch.NewLauncher(cfg, nil, mlm)

	if err := l.Launch(
		m.launchState.selectedVersion,
		m.launchState.username,
		m.launchState.serverAddr,
		m.launchState.serverPort,
	); err != nil {
		m.statusMessage = fmt.Sprintf("启动失败: %v", err)
		m.statusType = "error"
		m.loading = false
	} else {
		m.statusMessage = "游戏已启动"
		m.statusType = "success"
		m.loading = false
	}
}

func (m *model) copyCommand() {
	if m.commandState.command != "" {
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(m.commandState.command)
		cmd.Run()
		m.statusMessage = "命令已复制到剪贴板"
	}
}

func (m *model) saveConfig() {
	cfg := loadConfig()
	cfg.GameDir = m.configState.gameDir
	cfg.JavaPath = m.configState.javaPath
	cfg.Width = m.configState.width
	cfg.Height = m.configState.height
	cfg.FullScreen = m.configState.fullScreen

	m.launchState.gameDir = m.configState.gameDir
	m.launchState.javaPath = m.configState.javaPath
	m.launchState.width = m.configState.width
	m.launchState.height = m.configState.height
	m.launchState.fullScreen = m.configState.fullScreen

	m.statusMessage = "配置已保存 (仅当前会话有效)"
	m.statusType = "success"
}

func (m *model) quit() {
	logger.Info("TUI 退出")
}

func (m *model) View() string {
	if m.width < 20 {
		m.width = 80
	}
	if m.height < 10 {
		m.height = 24
	}

	if m.showHelp {
		return m.renderHelp()
	}

	var content string

	header := m.renderHeader()
	nav := m.renderNavigation()
	main := m.renderMainView()
	status := m.renderStatusBar()

	content = header + "\n" + nav + "\n" + main + "\n" + status

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *model) renderHeader() string {
	title := "🎮 Gilvaa Minecraft 启动器"
	width := m.width - 4
	if width < 0 {
		width = 0
	}
	padding := width - len(title)
	if padding < 0 {
		padding = 0
		if width > 0 && width < len(title) {
			title = title[:width]
		} else if width <= 0 {
			title = ""
		}
	}
	return headerStyle.Render(title + strings.Repeat(" ", padding))
}

func (m *model) renderNavigation() string {
	var navItems []string
	for i, name := range viewNames {
		if ViewType(i) == m.currentView {
			navItems = append(navItems, selectedStyle.Render("["+name+"]"))
		} else {
			navItems = append(navItems, normalStyle.Render(name))
		}
	}
	nav := strings.Join(navItems, "  ")
	width := m.width - 4
	padding := width - len(nav) + len(navItems)*6
	if padding < 0 {
		padding = 0
	}
	return lipgloss.NewStyle().PaddingLeft(padding / 2).Render(nav)
}

func (m *model) renderMainView() string {
	switch m.currentView {
	case LaunchView:
		return m.renderLaunchView()
	case VersionView:
		return m.renderVersionView()
	case ConfigView:
		return m.renderConfigView()
	case JavaView:
		return m.renderJavaView()
	case CommandView:
		return m.renderCommandView()
	case InfoView:
		return m.renderInfoView()
	}
	return ""
}

func (m *model) renderLaunchView() string {
	content := borderStyle.Render(fmt.Sprintf(` %s | %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 游戏版本: [ %s %s ]  (下拉列表，显示已安装版本)
 用户名:   [ %s%s ]
 内存:     [ %s%s ]  (-Xmx)
 Java:     [ %s%s ] (可浏览/选择)
 游戏目录: [ %s%s ] (可修改)
 窗口:     宽 [%d%s] 高 [%d%s]  [%s] 全屏
 服务器:   [ %s%s ] 端口 [ %d%s ]

              %s`,
		dimStyle.Render("🚀 启动视图"),
		dimStyle.Render("目标：一步完成游戏启动"),
		m.launchState.selectedVersion,
		dimStyle.Render("▼"),
		m.launchState.username,
		m.getCursor(2),
		m.launchState.memory,
		m.getCursor(3),
		m.launchState.javaPath,
		m.getCursor(4),
		m.launchState.gameDir,
		m.getCursor(5),
		m.launchState.width,
		m.getCursor(6),
		m.launchState.height,
		m.getCursor(7),
		m.getCheckBox(m.launchState.fullScreen),
		m.launchState.serverAddr,
		m.getCursor(8),
		m.launchState.serverPort,
		m.getCursor(9),
		buttonStyle.Render(" 🟢 启动游戏 "),
	))

	help := dimStyle.Render("  ↑↓导航  ←→选择版本  Enter启动  Tab切换视图")
	return content + "\n" + help
}

func (m *model) getCursor(field int) string {
	if m.focusedField == field {
		return focusedStyle.Render("◀")
	}
	return dimStyle.Render(" ")
}

func (m *model) getCheckBox(checked bool) string {
	if checked {
		return greenStyle.Render("☑")
	}
	return normalStyle.Render("☐")
}

func (m *model) renderVersionView() string {
	leftItems := []string{}
	for i, v := range m.versionState.availableVersions {
		marker := "  "
		if v.Installed {
			marker = dimStyle.Render("✓ ")
		}
		item := fmt.Sprintf("%s%s  %s (%s)", marker, v.ID, v.Type, v.ReleaseTime[:10])
		if i == m.versionState.selectedIndex {
			item = selectedStyle.Render(item)
		} else {
			item = normalStyle.Render(item)
		}
		leftItems = append(leftItems, item)
	}

	leftContent := strings.Join(leftItems, "\n")
	rightItems := []string{}
	for _, v := range m.versionState.installedVersions {
		marker := dimStyle.Render("• ")
		if v == m.launchState.selectedVersion {
			marker = greenStyle.Render("▶ ")
		}
		rightItems = append(rightItems, marker+normalStyle.Render(v))
	}
	rightContent := strings.Join(rightItems, "\n")

	leftPanel := lipgloss.NewStyle().Width(m.width/2 - 2).Render(
		borderStyle.Render(fmt.Sprintf(" %s \n%s", dimStyle.Render("版本列表 (可用)"), leftContent)),
	)
	rightPanel := lipgloss.NewStyle().Width(m.width/2 - 2).Render(
		borderStyle.Render(fmt.Sprintf(" %s \n%s", dimStyle.Render("已安装版本"), rightContent)),
	)

	details := m.versionState.details
	detailsContent := fmt.Sprintf(` 版本: %s | 类型: %s | 需要Java: %d+
 发布时间: %s | 依赖库数: %d`,
		m.versionState.availableVersions[m.versionState.selectedIndex].ID,
		details.Type,
		details.JavaVersion,
		details.ReleaseTime,
		details.Libraries,
	)

	detailsPanel := borderStyle.Width(m.width - 4).Render(
		fmt.Sprintf(" %s\n%s", dimStyle.Render("详情"), detailsContent),
	)

	actions := fmt.Sprintf(" %s %s %s %s",
		buttonStyle.Render("[I 安装]"),
		buttonDimStyle.Render("[D 删除]"),
		buttonDimStyle.Render("[S 设为启动]"),
		buttonDimStyle.Render("[Enter 查看详情]"),
	)

	help := dimStyle.Render("  ↑↓选择  I安装 D删除 S设默认  Tab切换视图")

	return leftPanel + " " + rightPanel + "\n" + detailsPanel + "\n" + actions + "\n" + help
}

func (m *model) renderConfigView() string {
	content := borderStyle.Render(fmt.Sprintf(` %s | %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 配置文件路径: [ %s ] (只读显示)
 日志级别:    [ %s%s ]  (debug/info/warn/error)
 详细输出:    %s  --verbose

 默认游戏目录: [ %s%s ]
 默认Java路径: [ %s%s ]  (自动检测)
 默认内存:     [ %s%s ]
 默认窗口宽:   [ %d%s ]  高: [ %d%s ]
 默认全屏:     %s

              %s  %s`,
		dimStyle.Render("⚙️ 配置视图"),
		dimStyle.Render("全局设置"),
		m.configState.configPath,
		m.configState.logLevel,
		m.getCursor(1),
		m.getCheckBox(m.configState.verbose),
		m.configState.gameDir,
		m.getCursor(3),
		m.configState.javaPath,
		m.getCursor(4),
		m.configState.memory,
		m.getCursor(5),
		m.configState.width,
		m.getCursor(6),
		m.configState.height,
		m.getCursor(7),
		m.getCheckBox(m.configState.fullScreen),
		buttonStyle.Render("[保存配置]"),
		buttonDimStyle.Render("[重置为默认]"),
	))

	help := dimStyle.Render("  ↑↓导航  Enter切换  Ctrl+S保存  Tab切换视图")
	return content + "\n" + help
}

func (m *model) renderJavaView() string {
	var items []string
	for i, java := range m.javaState.javaList {
		marker := dimStyle.Render("  ")
		if java.Path == m.launchState.javaPath {
			marker = greenStyle.Render("▶ ")
		}
		recMarker := ""
		if java.Major >= 17 {
			recMarker = yellowStyle.Render(" (recommended)")
		}
		item := fmt.Sprintf("%sJava %d%s\n    Path: %s\n    Version: %s\n    Arch: %s",
			marker, java.Major, recMarker, java.Path, java.Version, java.Arch)
		if i == m.javaState.selectedIndex {
			item = selectedStyle.Render(item)
		} else {
			item = normalStyle.Render(item)
		}
		items = append(items, item)
	}

	content := strings.Join(items, "\n")

	panel := borderStyle.Width(m.width - 4).Render(fmt.Sprintf(` %s
%s`, dimStyle.Render("系统已安装的 Java 版本:"), content))

	actions := fmt.Sprintf(" %s  %s",
		buttonStyle.Render("[设为默认Java]"),
		buttonDimStyle.Render("[测试所选Java]"),
	)

	help := dimStyle.Render("  ↑↓选择  Enter设为默认  Tab切换视图")
	return panel + "\n" + actions + "\n" + help
}

func (m *model) renderCommandView() string {
	header := dimStyle.Render("完整启动命令 (基于当前配置):")
	cmd := m.commandState.command
	if cmd == "" {
		cmd = dimStyle.Render("< 请先在启动视图中选择版本 >")
	}

	panel := borderStyle.Width(m.width - 4).Height(10).Render(fmt.Sprintf(` %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s`, header, cmd))

	actions := fmt.Sprintf(" %s  %s",
		buttonDimStyle.Render("[复制到剪贴板]"),
		buttonStyle.Render("[启动游戏]"),
	)

	help := dimStyle.Render("  Enter启动  Tab切换视图")
	return panel + "\n" + actions + "\n" + help
}

func (m *model) renderInfoView() string {
	content := borderStyle.Render(fmt.Sprintf(` %s | %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 版本: [ %s ]  (可输入或从列表选择)

 详细信息:
   类型: %s
   主类: %s
   发布时间: %s
   依赖库数量: %d
   所需Java版本: %d+
   资产索引: %s`,
		dimStyle.Render("ℹ️ 版本信息视图"),
		dimStyle.Render("查看任意版本详情"),
		m.infoState.versionID,
		m.infoState.details.Type,
		m.infoState.details.MainClass,
		m.infoState.details.ReleaseTime,
		m.infoState.details.Libraries,
		m.infoState.details.JavaVersion,
		m.infoState.details.Assets,
	))

	help := dimStyle.Render("  Enter查看详情  Tab切换视图")
	return content + "\n" + help
}

func (m *model) renderStatusBar() string {
	statusColor := normalStyle
	if m.statusType == "error" {
		statusColor = redStyle
	} else if m.statusType == "success" {
		statusColor = greenStyle
	}

	statusText := statusColor.Render(m.statusMessage)
	versionText := dimStyle.Render(fmt.Sprintf("版本: %s", m.launchState.selectedVersion))
	javaText := dimStyle.Render(fmt.Sprintf("Java: %s", m.launchState.javaPath))
	memoryText := dimStyle.Render(fmt.Sprintf("内存: %s", m.launchState.memory))

	width := m.width - 4
	left := statusText
	right := fmt.Sprintf("%s | %s | %s | Ctrl+S保存 | F1帮助 | Ctrl+Q退出", versionText, javaText, memoryText)
	rightLen := len(right) + 4
	leftLen := width - rightLen

	if leftLen < 0 {
		leftLen = 0
	}

	if len(left) > leftLen {
		if leftLen > 3 {
			left = left[:leftLen-3] + "..."
		} else {
			left = ""
		}
	} else {
		left = left + strings.Repeat(" ", leftLen-len(left))
	}

	return statusBarStyle.Render(" " + left + right)
}

func (m *model) renderHelp() string {
	help := fmt.Sprintf(` %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 全局热键 (在任何视图中有效)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Ctrl+S    保存当前配置到文件
  Ctrl+Q    退出 TUI
  F1        显示/关闭帮助弹窗
  Tab       切换到下一个视图
  Shift+Tab 切换到上一个视图
  1-6       直接跳转到对应视图

   1: 启动视图
   2: 版本管理视图
   3: 配置视图
   4: Java管理视图
   5: 命令预览视图
   6: 版本信息视图

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 启动视图热键
   ↑↓       切换表单焦点
   ←→       切换版本列表
   Enter    启动游戏 (焦点在按钮时)

 版本管理视图热键
   ↑↓       选择版本
   I        安装选中的版本
   D        删除选中的版本
   S        将选中版本设为启动版本
   Enter    查看版本详情

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 按 Esc 或 q 关闭此帮助
`, headerStyle.Render("帮助"))

	return lipgloss.NewStyle().
		Width(m.width - 10).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(help)
}
