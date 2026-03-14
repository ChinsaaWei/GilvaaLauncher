# Minecraft Launcher - Command Line Tool

一个功能完整的Minecraft命令行启动器，支持版本管理、下载和启动Minecraft。

## 功能特性

- **版本管理**: 列出、搜索、安装、卸载Minecraft版本
- **资源下载**: 自动下载assets、libraries和natives
- **启动游戏**: 支持自定义Java路径、内存分配、窗口大小等
- **版本隔离**: 每个版本独立管理，互不干扰
- **Java检测**: 自动检测系统中的Java安装
- **模组加载器接口**: 预留Forge、Fabric等模组加载器接口

## 安装

### 编译

```bash
go build -o mclauncher.exe
```

### 依赖

- Go 1.25.0+
- Java 17+ (用于运行Minecraft 1.18+)

## 使用方法

### 基本命令

#### 查看帮助

```bash
mclauncher --help
```

#### 列出可用版本

```bash
# 列出所有版本
mclauncher version list

# 只列出release版本
mclauncher version list release

# 只列出snapshot版本
mclauncher version list snapshot
```

#### 搜索版本

```bash
mclauncher version search 1.20
```

#### 安装版本

```bash
mclauncher version install 1.20.4
```

#### 查看已安装版本

```bash
mclauncher version installed
```

#### 查看版本信息

```bash
mclauncher info 1.20.4
```

#### 卸载版本

```bash
mclauncher version uninstall 1.20.4
```

#### 准备版本（下载资源但不启动）

```bash
mclauncher prepare 1.20.4
```

#### 启动游戏

```bash
# 基本启动
mclauncher launch 1.20.4

# 指定用户名
mclauncher launch 1.20.4 MyUsername

# 连接到服务器
mclauncher launch 1.20.4 --server mc.hypixel.net --port 25565
```

#### Java管理

```bash
# 列出已安装的Java版本
mclauncher java list
```

### 高级选项

#### 全局参数

- `--java`: 指定Java可执行文件路径
- `--memory`: 设置内存分配（如 2G, 4G）
- `--game-dir`: 指定Minecraft游戏目录
- `--username`: 设置用户名
- `--width`: 设置窗口宽度
- `--height`: 设置窗口高度
- `--fullscreen`: 全屏模式
- `--verbose`: 详细输出
- `--log-level`: 日志级别（debug, info, warn, error）

#### 示例

```bash
# 使用4GB内存启动
mclauncher launch 1.20.4 --memory 4G

# 使用特定Java版本
mclauncher launch 1.20.4 --java "C:\Program Files\Java\jdk-17\bin\java.exe"

# 自定义游戏目录
mclauncher launch 1.20.4 --game-dir "D:\Minecraft"

# 全屏模式
mclauncher launch 1.20.4 --fullscreen

# 自定义窗口大小
mclauncher launch 1.20.4 --width 1280 --height 720
```

## 项目结构

```
Launcher/
├── cmd/              # 命令行接口
│   └── root.go       # CLI命令定义
├── config/           # 配置管理
│   └── config.go     # 配置结构体
├── downloader/       # 下载器
│   ├── downloader.go # 通用下载器
│   ├── version.go    # 版本下载
│   ├── assets.go     # 资源下载
│   └── libraries.go  # 库文件下载
├── launcher/         # 启动器核心
│   ├── launcher.go   # 启动逻辑
│   └── launch_config.go # 启动配置
├── logger/           # 日志系统
│   └── logger.go     # 日志实现
├── modloader/        # 模组加载器接口
│   └── interface.go  # 模组加载器定义
├── util/             # 工具函数
│   ├── hash.go       # 哈希计算
│   └── java.go       # Java检测
├── version/          # 版本管理
│   └── manager.go    # 版本管理器
├── go.mod            # Go模块定义
└── main.go           # 程序入口
```

## 模组加载器接口

项目预留了模组加载器接口，支持以下类型：

- Forge
- Fabric
- Quilt
- NeoForge

可以通过实现 `ModLoader` 接口来添加对特定模组加载器的支持。

## 配置文件

启动器会在以下位置创建配置和日志：

- Windows: `%APPDATA%\.minecraft`
- 工作目录: `~/.mclauncher`
- 日志文件: `~/.mclauncher/logs/launcher.log`

## 注意事项

1. 首次运行需要下载Minecraft资源，可能需要较长时间
2. 确保有足够的磁盘空间（通常需要2-5GB）
3. 某些版本需要特定版本的Java
4. 离线模式使用默认的访问令牌

## 许可证

本项目仅供学习和个人使用。

## 贡献

欢迎提交问题和拉取请求。
