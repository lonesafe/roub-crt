# roub-crt 规格文档 V1.0

## 1. 项目概述
- **产品名称**: roub-crt
- **产品类型**: Go语言桌面终端仿真与文件传输工具
- **核心功能**: 支持多协议远程连接(SSH/Telnet/Serial)、会话管理、文件传输、端口转发、图形界面
- **目标用户**: 系统管理员、网络工程师、开发运维人员

## 2. 技术架构

### 2.1 项目结构
```
roub-crt/
├── cmd/
│   ├── root.go          # 根命令
│   ├── connect.go       # 连接命令
│   ├── session.go       # 会话管理
│   ├── transfer.go      # 文件传输
│   ├── tunnel.go        # 端口转发
│   ├── key.go          # 密钥管理
│   ├── interactive.go  # 交互式TUI
│   ├── gui.go          # GUI主程序 (fyne框架)
│   └── gui_stub.go     # GUI存根 (无GUI时)
├── internal/
│   ├── config/          # 配置管理
│   ├── session/         # 会话管理
│   ├── connection/      # 连接协议实现
│   │   ├── ssh.go       # SSH连接
│   │   ├── telnet.go    # Telnet连接
│   │   ├── serial.go    # 串口连接
│   │   └── rlogin.go    # Rlogin连接
│   ├── terminal/        # 终端仿真
│   ├── file_transfer/   # 文件传输(SFTP/SCP)
│   ├── crypto/          # 加密功能
│   └── tunnel/          # 端口转发隧道
├── pkg/
│   └── ui/              # 终端UI组件
├── sessions/            # 会话存储目录
├── config.yaml          # 配置文件
├── main.go              # 入口文件
└── go.mod
```

### 2.2 依赖包
- fyne.io/fyne - 跨平台GUI框架
- github.com/spf13/cobra - CLI框架
- github.com/spf13/viper - 配置管理
- golang.org/x/crypto - 加密算法(AES, Twofish, RSA, ECDSA)
- github.com/pkg/term - 终端操作
- github.com/AlecAivazis/survey/v2 - 交互式提示
- github.com/charmbracelet/lipgloss - 终端样式

### 2.3 构建模式
- **CLI模式**: `go build` (默认)
- **GUI模式**: `go build -tags gui` (需要OpenGL/X11开发库)

## 3. 功能规格

### 3.1 连接协议 (internal/connection/)
| 协议 | 功能 |
|------|------|
| SSH2 | 加密远程登录，支持密钥认证 |
| SSH1 | 兼容旧系统 |
| Telnet | 明文远程登录 |
| Serial | 串口通信(波特率、数据位、停止位、校验) |
| Rlogin | 简单远程登录 |
| TAPI | 电话API连接(基础支持) |

### 3.2 会话管理 (internal/session/)
- 保存连接配置: 主机、端口、用户名、协议、编码
- 文件夹分类: 支持树形目录组织会话
- 快速连接: 一键连接已保存会话
- 导入/导出: JSON格式会话文件

### 3.3 终端仿真 (internal/terminal/)
- VT100/VT220/xterm兼容
- 256色支持
- Unicode显示
- 滚动缓冲
- 自定义配色方案(内置方案 + 自定义)

### 3.4 文件传输 (internal/file_transfer/)
- SFTP协议支持
- SCP协议支持
- 双窗格布局(本地/远程)
- 文件操作: 上传、下载、删除、重命名、创建目录
- 隐藏文件显示
- 文件排序(名称、大小、日期)

### 3.5 端口转发 (internal/tunnel/)
- 本地端口转发: `ssh -L`
- 远程端口转发: `ssh -R`
- 动态转发(SOCKS5代理): `ssh -D`

### 3.6 安全功能 (internal/crypto/)
- 加密算法: AES-256, Twofish-256
- 密钥类型: RSA(2048/4096), DSA, ECDSA
- 密钥管理: 生成、导入、导出

## 4. CLI命令设计

### 4.1 主命令
```bash
roub-crt [command]
```

### 4.2 子命令
| 命令 | 说明 | 示例 |
|------|------|------|
| connect | 连接到主机 | `roub-crt connect 192.168.1.100` |
| session | 管理会话 | `roub-crt session list` |
| session add | 添加会话 | `roub-crt session add` |
| session edit | 编辑会话 | `roub-crt session edit <name>` |
| session del | 删除会话 | `roub-crt session del <name>` |
| transfer | 文件传输 | `roub-crt transfer <session>` |
| tunnel | 端口转发 | `roub-crt tunnel local 8080:remote:80` |
| keygen | 生成密钥 | `roub-crt keygen -t rsa` |
| config | 配置管理 | `roub-crt config set colorscheme` |
| interactive | 交互模式 | `roub-crt interactive` |

### 4.3 连接参数
```bash
roub-crt connect <host> [flags]
  -p, --port int        端口 (默认: 22 for SSH, 23 for Telnet)
  -u, --user string     用户名
  -P, --password string 密码
  -k, --key string      密钥文件路径
  -t, --type string     协议类型 (ssh/telnet/serial/rlogin)
  --serial-baud int     串口波特率 (默认: 115200)
  --serial-data int     数据位 (默认: 8)
  --serial-stop int     停止位 (默认: 1)
  --serial-parity int   校验 (0=None, 1=Odd, 2=Even)
```

## 5. 配置文件 (config.yaml)

```yaml
app:
  name: "roub-crt"
  version: "1.0.0"

terminal:
  font: "Monospace"
  font_size: 14
  scrollback: 10000
  cursor_shape: block  # block/underline/beam

colorschemes:
  default:
    background: "#1E1E1E"
    foreground: "#D4D4D4"
    cursor: "#FFFFFF"
  monokai:
    background: "#272822"
    foreground: "#F8F8F2"
    cursor: "#F8F8F0"
  solarized_dark:
    background: "#002B36"
    foreground: "#839496"
    cursor: "#D33682"

sessions:
  default_folder: "./sessions"
  auto_save: true

connection:
  timeout: 30
  keepalive: 10
  default_encoding: "UTF-8"

security:
  encrypt_transfers: true
  strict_host_key_check: true
```

## 6. 交互式界面

### 6.1 主菜单
```
╔══════════════════════════════════════════╗
║           roub-crt v1.0.0                 ║
║     Terminal Emulation & File Transfer   ║
╠══════════════════════════════════════════╣
║  [1] Quick Connect                        ║
║  [2] Session Manager                      ║
║  [3] File Transfer                       ║
║  [4] Port Tunnel                         ║
║  [5] Key Manager                          ║
║  [6] Settings                            ║
║  [0] Exit                                ║
╚══════════════════════════════════════════╝
```

### 6.2 文件传输双窗格
```
┌─────────────────────┬─────────────────────┐
│  LOCAL (./)         │  REMOTE (/)         │
├─────────────────────┼─────────────────────┤
│  ..                 │  ..                 │
│  ├── projects/      │  ├── home/          │
│  ├── downloads/     │  ├── var/           │
│  └── config.yaml    │  └── etc/           │
├─────────────────────┼─────────────────────┤
│  [F5] Upload │ [F6] Download │ [F10] Quit │
└─────────────────────┴─────────────────────┘
```

## 7. 验收标准

### 7.1 连接功能
- [ ] SSH2连接成功，支持密码和密钥认证
- [ ] Telnet连接成功
- [ ] 串口连接成功，配置参数可调
- [ ] Rlogin连接成功
- [ ] 支持会话保存和加载

### 7.2 终端功能
- [ ] 终端输出正常，支持滚屏
- [ ] 256色彩色显示
- [ ] 自定义配色方案可切换

### 7.3 文件传输
- [ ] SFTP上传/下载功能
- [ ] 双窗格显示
- [ ] 隐藏文件显示切换
- [ ] 文件排序功能

### 7.4 端口转发
- [ ] 本地端口转发工作
- [ ] 远程端口转发工作
- [ ] 动态转发工作

### 7.5 安全功能
- [ ] AES/Twofish加密传输
- [ ] RSA/DSA/ECDSA密钥生成
- [ ] 密钥导入/导出

## 8. 数据流

### 8.1 连接流程
```
用户输入 → 协议选择 → 认证处理 → 会话建立 → 终端仿真 → 数据交互
```

### 8.2 文件传输流程
```
选择文件 → SFTP/SCP通道 → 加密传输(可选) → 进度显示 → 完成确认
```

### 8.3 端口转发流程
```
监听本地端口 → 加密隧道 → 远程转发 → 目标服务
```

## 9. CI/CD 自动构建

### 9.1 GitHub Actions 工作流

文件: `.github/workflows/build.yml`

| 触发条件 | 构建产物 |
|---------|---------|
| Push tags `v*` | CLI + GUI 多平台二进制 |
| Push to `main` | CLI + GUI 多平台二进制 |
| PR to `main` | 测试 + 构建验证 |
| Manual trigger | CLI + GUI 多平台二进制 |

### 9.2 构建矩阵

| 平台 | CLI | GUI | 输出文件 |
|------|-----|-----|---------|
| Linux amd64 | ✓ | ✓ | `roub-crt`, `roub-crt-gui` |
| Linux arm64 | ✓ | ✓ | `roub-crt`, `roub-crt-gui` |
| Windows amd64 | ✓ | ✓ | `roub-crt.exe`, `roub-crt-gui.exe` |
| macOS amd64 | ✓ | ✓ | `roub-crt`, `roub-crt-gui` |
| macOS arm64 | ✓ | ✓ | `roub-crt`, `roub-crt-gui` |

### 9.3 发布流程

1. **版本标签**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **自动创建 Release**:
   - GitHub Actions 自动构建所有平台
   - 生成 SHA256 校验和
   - 创建 GitHub Release 并上传 artifacts

### 9.4 GoReleaser (可选)

使用 `goreleaser` 替代 GitHub Actions:
```bash
# 安装 goreleaser
brew install goreleaser

# 快速测试
goreleaser check
goreleaser snapshot --clean

# 正式发布
goreleaser release --clean
```

配置: `.goreleaser.yml`

