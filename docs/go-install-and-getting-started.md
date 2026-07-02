# Go 安装与入门：从环境配置到第一个程序

> 知识点总结：在不同操作系统上安装 Go、理解关键环境变量、验证安装，并通过 Hello World 与简单模块示例完成第一次构建与运行。

---

## 1. 安装前需要想清楚的事

### 1.1 Go 工具链装在哪里

Go 安装后通常包含两部分：

| 部分 | 默认位置（示例） | 作用 |
|------|------------------|------|
| **GOROOT** | macOS/Linux: `/usr/local/go`；Windows: `C:\Program Files\Go` | Go 编译器、`go` 命令、标准库 |
| **GOPATH** | 默认 `~/go`（Windows: `%USERPROFILE%\go`） | 模块缓存、构建缓存、`go install` 的二进制目录等 |

现代开发以 **Go Modules** 为主，**不必**再把项目放在 `$GOPATH/src` 下；但 `GOPATH` 仍被工具链用于缓存与安装路径。

### 1.2 选哪个版本

- 官网 [https://go.dev/dl/](https://go.dev/dl/) 提供 **Stable（稳定版）** 与 **Unstable（开发版）**。
- 学习与新项目优先用 **最新稳定版**。
- 团队项目以 `go.mod` 里的 `go 1.xx` 为准，必要时安装对应或更高的小版本。

### 1.3 选哪个 CPU 架构安装包

| 系统 | 常见架构 | 安装包关键词 |
|------|----------|--------------|
| macOS（Apple Silicon） | arm64 | `darwin-arm64` |
| macOS（Intel） | amd64 | `darwin-amd64` |
| Windows | amd64 / arm64 | `windows-amd64` / `windows-arm64` |
| Linux | amd64 / arm64 | `linux-amd64` / `linux-arm64` |

不确定时：

```bash
# macOS / Linux
uname -m    # arm64 或 x86_64

# Windows PowerShell
$env:PROCESSOR_ARCHITECTURE
```

---

## 2. 各系统安装方式

### 2.1 macOS

**方式 A：官方安装包（推荐新手）**

1. 打开 [https://go.dev/dl/](https://go.dev/dl/)，下载 `.pkg`（如 `go1.24.x.darwin-arm64.pkg`）。
2. 双击安装，默认装到 `/usr/local/go`。
3. 安装程序一般会自动把 `/usr/local/go/bin` 加入 PATH。

**方式 B：Homebrew**

```bash
brew install go
```

- 路径多在 `/opt/homebrew/opt/go`（Apple Silicon）或 `/usr/local/opt/go`（Intel）。
- 升级：`brew upgrade go`。

**方式 C：asdf / mise（多版本管理）**

适合需要在本机切换多个 Go 版本时：

```bash
# mise 示例
mise install go@1.24.2
mise use -g go@1.24.2
```

### 2.2 Windows

**方式 A：官方 MSI 安装包（推荐）**

1. 下载 `go1.xx.x.windows-amd64.msi`。
2. 默认安装到 `C:\Program Files\Go`。
3. 安装器会配置用户级 `PATH`（含 `Go\bin` 与 `GOROOT\bin`）。

**方式 B：ZIP 绿色包**

1. 解压到如 `C:\Go`。
2. 手动设置环境变量（见第 3 节）。

**方式 C：包管理器**

```powershell
# winget
winget install GoLang.Go

# Chocolatey
choco install golang
```

### 2.3 Linux

**方式 A：官方 tar.gz（通用、可控）**

```bash
# 以 linux-amd64 为例，版本号请替换为当前稳定版
curl -LO https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
```

将 `/usr/local/go/bin` 加入 `PATH`（写入 `~/.bashrc` 或 `~/.zshrc`）：

```bash
export PATH=$PATH:/usr/local/go/bin
```

**方式 B：发行版包管理器**

```bash
# Debian/Ubuntu（版本可能偏旧，生产环境注意核对）
sudo apt update && sudo apt install golang-go

# Fedora
sudo dnf install golang

# Arch
sudo pacman -S go
```

若仓库版本过旧，仍建议用官方 tar.gz 或 [go.dev/dl](https://go.dev/dl/)。

---

## 3. 环境变量：装完之后要理解什么

### 3.1 必知变量

| 变量 | 是否常手动设置 | 含义 |
|------|----------------|------|
| **PATH** | 常需确认 | 必须包含 `go` 可执行文件所在目录（一般为 `$GOROOT/bin`） |
| **GOROOT** | 一般不用 | Go 安装根目录；装错或多版本时用于排查 |
| **GOPATH** | 可选 | 工作区与缓存根目录，默认 `~/go` |
| **GOBIN** | 可选 | `go install` 输出目录，默认 `$GOPATH/bin` |
| **GOPROXY** | 国内建议设置 | 模块下载代理，如 `https://goproxy.cn,direct` |
| **GOSUMDB** | 按需 | 校验和数据库；内网/offline 可设为 `off`（需知风险） |
| **GO111MODULE** | 现代 Go 可忽略 | Go 1.16+ 默认开启 Modules |

查看当前生效值：

```bash
go env
go env GOPATH GOROOT GOPROXY
```

### 3.2 PATH 配置示例

**macOS / Linux（zsh，`~/.zshrc`）**

```bash
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin   # go install 装的全局命令
export GOPROXY=https://goproxy.cn,direct # 国内加速（可选）
```

**Windows（系统环境变量）**

| 变量 | 值示例 |
|------|--------|
| PATH | 追加 `C:\Program Files\Go\bin` 与 `%USERPROFILE%\go\bin` |
| GOPROXY | `https://goproxy.cn,direct`（可选） |

修改后 **重新打开终端**，再执行 `go version`。

### 3.3 安装后自检清单

```bash
go version          # 应显示 go version go1.xx.x ...
which go            # macOS/Linux：go 的路径
where go            # Windows：go 的路径

go env GOROOT
go env GOPATH
```

期望结果：

- `go version` 能输出版本号；
- `which go` 指向你期望的安装方式（官方 `/usr/local/go/bin/go` 或 brew 路径等）；
- 不应出现「command not found」。

---

## 4. 第一个程序：Hello World

### 4.1 不用模块的最小示例（单文件）

任意目录均可（不必在 GOPATH/src 下）：

```bash
mkdir -p ~/code/hello-go && cd ~/code/hello-go
```

创建 `main.go`：

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

运行与构建：

```bash
go run main.go              # 直接运行
go build -o hello main.go   # 编译为当前目录可执行文件
./hello                     # macOS/Linux
# hello.exe                 # Windows
```

### 4.2 推荐方式：Go Module 小项目

```bash
mkdir -p ~/code/hello-module && cd ~/code/hello-module
go mod init example.com/hello
```

`go mod init` 会生成 `go.mod`，模块路径可自定义（学习用 `example.com/hello` 即可）。

`main.go` 同上，然后：

```bash
go run .
go build -o bin/hello .
```

目录结构：

```
hello-module/
├── go.mod
└── main.go
```

### 4.3 稍复杂一点：多文件与包

```bash
go mod init example.com/calc
```

`main.go`：

```go
package main

import (
	"fmt"

	"example.com/calc/mathutil"
)

func main() {
	fmt.Println("1 + 2 =", mathutil.Add(1, 2))
}
```

`mathutil/add.go`：

```go
package mathutil

func Add(a, b int) int {
	return a + b
}
```

```bash
go run .
```

说明：子目录 `mathutil` 的 **包名** 是 `mathutil`，**导入路径** 是 `example.com/calc/mathutil`（与 `go.mod` 里 module 名 + 相对路径一致）。

---

## 5. 日常会用到的几条命令

| 命令 | 作用 |
|------|------|
| `go version` | 查看版本 |
| `go env` | 查看环境配置 |
| `go mod init <module>` | 初始化模块 |
| `go mod tidy` | 整理依赖（增删 `go.mod` / `go.sum`） |
| `go get <pkg>@<version>` | 添加或升级依赖 |
| `go run .` | 编译并运行当前模块 |
| `go build` | 编译当前包 |
| `go test ./...` | 运行测试 |
| `go install <pkg>` | 编译并安装到 `GOBIN` 或 `$GOPATH/bin` |
| `go fmt ./...` | 格式化代码 |

格式化与测试示例：

```bash
go fmt ./...
go test ./...
```

---

## 6. 编辑器 / IDE 简要说明

- **VS Code / Cursor**：安装官方扩展 **Go**（golang.go），保存时会提示安装 `gopls`、`dlv` 等，选 Install 即可。
- **GoLand**：开箱即用，适合重度 Go 开发。
- 无论哪种 IDE，底层都依赖本机已正确安装的 `go` 与 `GOPATH/bin` 下的工具。

---

## 7. 常见问题排查

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| `go: command not found` | PATH 未含 Go 的 bin | 检查安装路径并写入 shell 配置 |
| 版本不对（过旧） | 系统包管理器版本旧，或多版本冲突 | `which go` 看实际用的是哪一个；优先官方包或 mise |
| `go: cannot find main module` | 目录下没有 `go.mod` | 在项目根执行 `go mod init ...` |
| 依赖下载慢或超时 | 默认走 proxy.golang.org | 设置 `GOPROXY=https://goproxy.cn,direct` |
| `go install` 的命令找不到 | `GOBIN` 或 `$GOPATH/bin` 不在 PATH | 把 `$(go env GOPATH)/bin` 加入 PATH |

---

## 8. 与本仓库其他文档的关系

- 本文：**安装 + 环境变量 + 第一个可运行项目**。
- [Go 构建与依赖管理演进：从 GOPATH 到 Modules](go-build-and-dependency-evolution.md)：**GOPATH / Modules / Workspace** 的历史与日常命令。

建议学习顺序：先完成本文的安装与 Hello World，再阅读依赖管理演进文档。

---

## 9. 参考链接

- 官方下载：[https://go.dev/dl/](https://go.dev/dl/)
- 官方安装说明：[https://go.dev/doc/install](https://go.dev/doc/install)
- 官方教程：[https://go.dev/doc/tutorial/getting-started](https://go.dev/doc/tutorial/getting-started)
- 环境变量说明：[https://pkg.go.dev/cmd/go#hdr-Environment_variables](https://pkg.go.dev/cmd/go#hdr-Environment_variables)
