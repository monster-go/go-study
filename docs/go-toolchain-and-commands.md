# Go 工具链与命令：从 go doc 到 go test 的完整指南

> 知识点总结：系统掌握 Go 安装后附带的 `go` 命令及其子工具，能查文档、编译运行、管理依赖、格式化与静态检查，并知道各命令在什么场景下使用。

---

## 1. 为什么需要了解 Go 工具链

安装 Go 后得到的不仅是编译器，而是一套完整的**工具链（toolchain）**。日常开发中几乎所有操作——查 API、编译、测试、拉依赖、格式化——都通过 `go` 命令完成，而不是单独记忆十几个独立程序的名字。

不了解工具链容易导致：

- 只会 `go run`，遇到 `go mod tidy`、`go vet` 不知道干什么
- 查内置函数文档失败（如 `go doc builtin.delete` 报错）
- 把 `go get` 和 `go install` 混用，或误以为必须放在 GOPATH 下

本篇在 [安装与入门](go-install-and-getting-started.md) 和 [构建与依赖演进](go-build-and-dependency-evolution.md) 的基础上，**按使用场景罗列全部常用命令**，形成可查阅的教程。

---

## 2. 核心概念

### 2.1 三层工具结构

```
┌─────────────────────────────────────────────────────────┐
│  go 命令（入口）                                          │
│  go build / run / test / doc / mod / get / ...          │
├─────────────────────────────────────────────────────────┤
│  go tool（底层工具，一般间接调用）                         │
│  compile / link / vet / cover / pprof / trace / ...     │
├─────────────────────────────────────────────────────────┤
│  独立可执行文件（部分通过 go install 安装到 GOBIN）        │
│  gopls / dlv / staticcheck / ...（非 Go 发行版自带）     │
└─────────────────────────────────────────────────────────┘
```

| 层级 | 典型命令 | 日常是否直接调用 |
|------|----------|------------------|
| `go` 子命令 | `go build`、`go doc` | **是**，主力 |
| `go tool` | `go tool pprof`、`go tool cover` | 偶尔，排查性能/覆盖率时用 |
| 第三方工具 | `gopls`、`dlv` | IDE / 调试时自动调用 |

### 2.2 查看帮助的统一方式

```bash
go help                    # 列出所有 go 子命令
go help <command>          # 某个子命令的详细说明
go help mod tidy           # 子命令还可以有子命令
go help environment        # 帮助主题（非子命令）
```

### 2.3 命令与包模式（package pattern）

许多命令接受**包路径**或**模式**，例如：

| 写法 | 含义 |
|------|------|
| `.` | 当前目录的包 |
| `./...` | 当前目录及所有子目录的包 |
| `example.com/foo` | 模块内某 import 路径 |
| `std` | 标准库所有包 |
| `cmd` | GOROOT 下 cmd 目录的包 |

详见 `go help packages`。

### 2.4 命令总览

执行 `go help` 可看到全部子命令（Go 1.26 示例）：

| 分类 | 命令 | 一句话说明 |
|------|------|------------|
| **环境与信息** | `version` | 查看 Go 版本 |
| | `env` | 查看/设置环境变量 |
| | `doc` | 查看包或符号的文档 |
| | `list` | 列出包信息 |
| | `bug` | 在浏览器打开 bug 报告页 |
| **编译与运行** | `build` | 编译，生成可执行文件或检查能否编译 |
| | `run` | 编译并立即运行 |
| | `install` | 编译并安装到 GOBIN |
| | `clean` | 清理编译产物与缓存 |
| **代码质量** | `fmt` | 格式化（调用 gofmt） |
| | `vet` | 静态检查可疑写法 |
| | `fix` | 自动修复旧版 API 用法 |
| | `test` | 运行测试、基准、模糊测试 |
| **依赖与模块** | `get` | 添加/升级/移除依赖 |
| | `mod` | 维护 go.mod（init、tidy、vendor 等） |
| | `work` | 维护 go.work 工作区 |
| **其他** | `generate` | 执行源码中的 `//go:generate` 指令 |
| | `tool` | 调用底层工具（compile、link、vet 等） |
| | `telemetry` | 管理遥测数据与设置 |

---

## 3. 环境与信息类命令

### 3.1 `go version` — 查看版本

```bash
go version                          # 当前 go 命令自身的版本
go version ./bin/myapp              # 查看某个二进制是用哪个 Go 版本编译的
go version -m ./bin/myapp           # 同时显示嵌入的模块信息
```

### 3.2 `go env` — 环境变量

```bash
go env                              # 以 shell 脚本形式输出全部变量
go env GOPATH GOROOT GOPROXY        # 只查看指定变量
go env -json                        # JSON 格式，便于脚本解析
go env -w GOPROXY=https://goproxy.cn,direct   # 持久化写入用户配置
go env -u GOPROXY                   # 取消 -w 设置的值，恢复默认
```

常用变量速查：

| 变量 | 含义 |
|------|------|
| `GOROOT` | Go 安装目录 |
| `GOPATH` | 工作区根（缓存、install 输出等） |
| `GOMODCACHE` | 模块下载缓存目录 |
| `GOBIN` | `go install` 可执行文件输出目录 |
| `GOPROXY` | 模块代理 |
| `GOOS` / `GOARCH` | 交叉编译目标平台 |

### 3.3 `go doc` — 查文档

`go doc` 是日常查 API 最常用的命令，但有几个细节需要理解。

**基本用法：**

```bash
cd example/maps
go doc fmt.Println                  # 查标准库函数
go doc fmt                          # 查整个包（含子项摘要）
go doc net/http Handler             # 查类型
go doc net/http Handler.ServeHTTP   # 查方法
go doc -all fmt                     # 显示包内所有符号的完整文档
go doc -src fmt.Println             # 显示源码（含注释）
```

**查内置函数（预声明标识符）：**

`delete`、`len`、`make`、`append` 等是**语言内置**的，不是普通导出符号，必须加 `-u`：

```bash
go doc builtin.delete               # 报错：no symbol delete in package builtin
go doc -u builtin.delete            # 正确：显示 delete 的文档
go doc -u -all builtin              # 查看所有内置函数文档
```

**当前包：**

```bash
go doc                              # 无参数时，显示当前目录包的文档
go doc -cmd                         # 对 main 包也显示导出符号
```

**原理说明：** `builtin` 包是文档用的虚拟包，内置标识符在源码里以小写形式声明（如 `func delete(...)`），`go doc` 默认只显示**导出（大写开头）**符号；`-u` 表示也显示未导出符号，从而能看到内置函数。

### 3.4 `go list` — 列出包信息

```bash
go list .                           # 当前包 import 路径
go list ./...                       # 当前模块所有包
go list -m all                      # 列出所有依赖模块及版本
go list -json .                     # JSON 格式，含 Dir、Deps 等字段
go list -f '{{.ImportPath}} -> {{.Dir}}' ./...
```

适合在脚本里获取包路径、源码目录、依赖关系。

### 3.5 `go bug` — 报告问题

```bash
go bug                              # 打开浏览器，预填系统与 Go 版本信息
```

---

## 4. 编译与运行类命令

### 4.1 `go build` — 编译

```bash
go build .                          # 编译当前 main 包，输出与目录同名的可执行文件
go build -o bin/app .               # 指定输出文件名
go build ./cmd/server               # 编译子目录中的 main 包
go build -tags debug .              # 按 build tag 条件编译
GOOS=linux GOARCH=amd64 go build .  # 交叉编译
```

编译**非 main 包**时只做语法检查，不产生可执行文件。

### 4.2 `go run` — 编译并运行

```bash
go run .                            # 编译并运行当前 main 包（产物在临时目录）
go run main.go                      # 指定单个或多个 .go 文件
go run . -mode basic                # 传给程序的参数写在包路径之后
go run golang.org/x/example/hello@latest   # 忽略当前 go.mod，临时运行远程模块
```

与 `go build` 的区别：`go run` 不保留可执行文件（除非加 `-exec` 等高级用法）。

### 4.3 `go install` — 安装到 GOBIN

```bash
go install .                        # 安装当前 main 包到 $GOBIN 或 $GOPATH/bin
go install golang.org/x/tools/cmd/goimports@latest   # 安装第三方工具
```

| 对比 | `go build` | `go install` |
|------|------------|--------------|
| 输出位置 | 当前目录或 `-o` 指定 | `GOBIN` / `$GOPATH/bin` |
| 典型场景 | 项目本地构建 | 安装全局 CLI 工具 |

### 4.4 `go clean` — 清理

```bash
go clean                            # 清理当前包目录下的旧产物
go clean -cache                     # 清理整个构建缓存（下次编译变慢）
go clean -testcache                 # 清理测试结果缓存
go clean -modcache                  # 清理模块下载缓存（慎用，需重新下载）
go clean -i -r ./...                # 递归清理并移除 go install 的文件
```

---

## 5. 代码质量类命令

### 5.1 `go fmt` — 格式化

```bash
go fmt ./...                        # 格式化当前模块所有包（改写文件）
go fmt -n ./...                     # 只打印会执行什么，不实际改文件
```

底层调用 `gofmt`。也可直接使用：

```bash
gofmt -w main.go                    # 格式化单个文件
gofmt -l .                          # 列出需要格式化的文件（不改写）
```

Go 官方风格：**Tab 缩进、goimports 管理 import**。IDE 保存时通常自动执行。

### 5.2 `go vet` — 静态检查

```bash
go vet ./...                        # 检查整个模块
go vet -json ./...                  # JSON 输出
go vet -fix ./...                   # 自动应用第一个建议修复
```

能发现编译器不报错的可疑写法，例如：

- `Printf` 格式串与参数不匹配
- `context.Background()` 未取消
- 结构体字面量中错误的字段名

CI 中常与 `go test` 一起运行。

### 5.3 `go fix` — 自动升级旧写法

```bash
go fix ./...                        # 将旧版 API 用法改写为新版本
go fix -diff ./...                  # 只显示 diff，不写入
```

升级 Go 大版本后，可用此命令批量更新代码（如 `io/ioutil` → `io`）。

### 5.4 `go test` — 测试

```bash
go test .                           # 运行当前包测试
go test ./...                       # 运行所有子包测试
go test -v .                        # 详细输出每个测试名
go test -run TestAdd .              # 只运行匹配的测试函数
go test -bench .                    # 运行基准测试
go test -cover ./...                # 显示覆盖率
go test -coverprofile=cover.out .   # 输出覆盖率文件
go test -race ./...                 # 开启竞态检测
go test -count=1 ./...              # 禁用测试缓存（排错时用）
```

测试相关约定见 `go help testfunc`（`*_test.go`、`TestXxx`、`BenchmarkXxx` 等）。

---

## 6. 依赖与模块类命令

模块的日常概念详见 [构建与依赖演进](go-build-and-dependency-evolution.md)，此处聚焦命令速查。

### 6.1 `go mod` 子命令

```bash
go mod init example.com/myapp       # 初始化 go.mod
go mod tidy                         # 增删依赖，同步 go.sum（最常用）
go mod download                     # 下载依赖到本地缓存
go mod verify                       # 校验缓存模块未被篡改
go mod graph                        # 打印依赖图
go mod why golang.org/x/net         # 解释为何需要某依赖
go mod edit -go=1.22                # 修改 go.mod 中的 Go 版本
go mod vendor                       # 将依赖复制到 vendor/ 目录
```

### 6.2 `go get` — 管理依赖版本

```bash
go get example.com/pkg              # 添加或升级到最新版本
go get example.com/pkg@v1.2.3       # 指定版本
go get example.com/pkg@none         # 移除该依赖
go get -u ./...                     # 升级所有直接依赖
go get go@latest                    # 升级 go.mod 中的 Go 版本声明
go get toolchain@patch              # 升级工具链补丁版本
```

**注意：** Go Modules 时代，`go get` 主要负责**改 go.mod**；安装可执行工具更推荐 `go install pkg@version`。

### 6.3 `go work` — 多模块工作区

```bash
go work init                        # 创建 go.work
go work use ./module-a ./module-b   # 将本地模块加入工作区
```

适合同时开发多个互相依赖的本地模块。详见 [官方 Workspace 教程](https://go.dev/doc/tutorial/workspaces)。

---

## 7. 其他命令

### 7.1 `go generate` — 代码生成

```bash
go generate ./...                   # 执行源码中 //go:generate 指令
```

`//go:generate` 是**注释里的指令**，不会自动运行，必须显式执行 `go generate`。常用于 `stringer`、`mockgen`、`protoc` 等。

### 7.2 `go tool` — 底层工具

```bash
go tool                             # 列出所有内置 tool
go tool vet help                    # 查看 vet 工具帮助
go tool cover -func=cover.out       # 从覆盖率文件生成报告
go tool pprof cpu.prof              # 交互式分析 CPU profile
go tool trace trace.out             # 查看执行跟踪
```

一般通过 `go` 子命令间接调用；直接调用 `go tool` 多在性能分析、覆盖率报告等场景。

| 工具 | 作用 |
|------|------|
| `compile` | 编译单个包为 .o |
| `link` | 链接为可执行文件 |
| `asm` | 汇编 |
| `cgo` | 生成 C 互操作代码 |
| `vet` | 静态分析（`go vet` 调用它） |
| `cover` | 覆盖率分析 |
| `pprof` | CPU/内存 profile 分析 |
| `trace` | 执行跟踪可视化 |

### 7.3 `go telemetry` — 遥测

```bash
go telemetry                        # 查看遥测状态
```

Go 工具链可选地上传匿名使用数据，可用此命令查看或调整。

---

## 8. 动手实践

在仓库示例目录中按顺序执行以下命令，熟悉工具链（需已安装 Go 1.21+）。

### 8.1 环境与文档

```bash
cd example/maps

go version                          # 确认版本
go env GOMOD GOPATH                 # 查看模块与路径

go doc fmt Println                  # 查标准库
go doc -u builtin.delete            # 查内置 delete（须 -u）
go doc -u map                       # 查内置 map 类型说明
```

### 8.2 编译与运行

```bash
go run .                            # 运行全部 demo
go run . -mode delete               # 只运行 delete 相关 demo
go build -o /tmp/maps-demo .        # 编译到指定路径
/tmp/maps-demo -mode basic          # 运行编译产物
```

### 8.3 格式化与检查

```bash
go fmt ./...                        # 格式化（maps 目录单文件）
go vet ./...                        # 静态检查，应无输出表示通过
```

### 8.4 模块信息

```bash
cd ../..                            # 回到仓库根（若有 go.mod）或 example/maps 的模块根
go mod tidy                         # 在含 go.mod 的目录执行
go list -m all                      # 列出依赖（若有）
```

### 8.5 测试（其他示例）

```bash
cd example/slices
go test -v .                        # 若包内有 _test.go 则运行；无则显示 [no test files]
```

### 8.6 自检清单

- [ ] `go help` 能列出子命令，并会用 `go help doc` 查细节
- [ ] `go doc -u builtin.delete` 能显示文档，理解为何需要 `-u`
- [ ] `go run .`、`go build` 能在 `example/maps` 下成功执行
- [ ] 能说出 `go build` 与 `go install` 的输出位置区别
- [ ] 能说出 `go get` 与 `go install` 的典型使用场景

---

## 9. 常见坑与排查

### 9.1 `go doc builtin.xxx` 报 no symbol

**现象：**

```text
go doc builtin.delete
doc: no symbol delete in package builtin
```

**原因：** 内置函数在 `builtin` 包中以小写声明，属于未导出符号。

**修复：** 加 `-u`：`go doc -u builtin.delete`。

### 9.2 `go: cannot find main module`

**现象：** 在无 `go.mod` 的目录执行 `go run .` 或 `go test`。

**修复：** 在该目录执行 `go mod init <module-path>`，或 `cd` 到已有模块根目录。单文件可 `go run main.go` 绕过模块。

### 9.3 `go install` 后命令找不到

**现象：** `go install xxx@latest` 成功，但 shell 里输入命令报 `command not found`。

**原因：** `GOBIN` 或 `$GOPATH/bin` 不在 `PATH` 中。

**修复：**

```bash
go env GOBIN GOPATH
export PATH="$PATH:$(go env GOPATH)/bin"   # 写入 ~/.zshrc 或 ~/.bashrc
```

### 9.4 `go get` 与 `go install` 混用

| 场景 | 应用命令 |
|------|----------|
| 给当前项目添加库依赖 | `go get example.com/lib@v1.0.0` |
| 安装全局 CLI 工具 | `go install golang.org/x/tools/cmd/goimports@latest` |
| 本地编译项目 | `go build` / `go run` |

Go 1.17+ 起，`go install` 安装带 `@version` 的程序时**不会**修改当前项目的 `go.mod`。

### 9.5 `go clean -modcache` 后编译变慢

**原因：** 删除了所有已下载模块，下次需重新从代理拉取。

**建议：** 日常排错优先 `go clean -cache` 或 `go clean -testcache`；`-modcache` 仅在缓存损坏时考虑。

### 9.6 测试结果被缓存

**现象：** 改了代码但 `go test` 仍显示旧的 `ok`，或未重新运行。

**原因：** Go 1.10+ 默认缓存测试结果。

**修复：** `go test -count=1 ./...` 禁用缓存。

---

## 10. 命令速查表（按使用频率）

| 频率 | 命令 | 典型场景 |
|------|------|----------|
| 每天 | `go run`、`go build` | 开发调试 |
| 每天 | `go doc`、`go doc -u` | 查 API / 内置函数 |
| 每天 | `go fmt` | 提交前格式化 |
| 经常 | `go test ./...` | 跑测试 |
| 经常 | `go mod tidy` | 改完 import 后整理依赖 |
| 经常 | `go get pkg@ver` | 加/升/删依赖 |
| 每周 | `go vet ./...` | CI / 提交前检查 |
| 偶尔 | `go install tool@latest` | 安装 goimports、dlv 等 |
| 偶尔 | `go tool pprof` | 性能分析 |
| 偶尔 | `go clean -cache` | 构建异常时清缓存 |
| 很少 | `go fix`、`go work` | 大版本升级、多模块联调 |

---

## 11. 小结与延伸阅读

### 11.1 要点回顾

1. **`go` 是唯一入口**：子命令覆盖编译、测试、文档、依赖全流程；细节用 `go help <cmd>` 查。
2. **查文档记得 `-u`**：内置函数 `delete`、`len`、`make` 等要用 `go doc -u builtin.xxx`。
3. **build / run / install 分工明确**：本地产物用 `build`，临时运行用 `run`，全局工具用 `install`。
4. **模块依赖靠 get + mod**：`go get` 改版本，`go mod tidy` 同步 `go.sum`；详见依赖演进文档。
5. **质量三板斧**：`fmt` 统一风格，`vet` 查可疑代码，`test` 验证行为。

### 11.2 官方链接

- Go 命令文档：[https://pkg.go.dev/cmd/go](https://pkg.go.dev/cmd/go)
- 内置标识符（builtin）：[https://pkg.go.dev/builtin](https://pkg.go.dev/builtin)
- 测试指南：[https://go.dev/doc/tutorial/add-a-test](https://go.dev/doc/tutorial/add-a-test)
- 模块参考：[https://go.dev/ref/mod](https://go.dev/ref/mod)
- 工具链说明：[https://go.dev/doc/install/source](https://go.dev/doc/install/source)

### 11.3 本仓库相关文档

- [Go 安装与入门](go-install-and-getting-started.md) — 安装与环境变量
- [Go 构建与依赖管理演进](go-build-and-dependency-evolution.md) — GOPATH / Modules / Workspace 深入
- [example/maps/](../example/maps/) — `delete`、map 操作示例，配合 `go doc -u builtin.delete` 使用
