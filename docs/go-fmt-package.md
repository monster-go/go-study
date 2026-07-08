# Go fmt 包：格式化输入输出

> 知识点总结：掌握 `fmt` 的打印、格式化字符串、扫描输入与 `Errorf` 错误包装；理解 **stdin / stdout / stderr** 与**缓冲区**（`bytes.Buffer`、`bufio`），能根据场景选用 `Print`/`Printf`/`Sprintf`/`Fprintf` 及常用动词（`%v`、`%d`、`%s`、`%w` 等）。

---

## 1. 为什么需要了解这个

几乎每个 Go 程序都会用到 `fmt`：调试输出、拼日志、读用户输入、构造错误信息。新人常见困惑包括：

- 分不清 `Print`、`Println`、`Printf` 何时用哪个
- 以为 `fmt.Println(a, b)` 和 `fmt.Printf("%v %v", a, b)` 完全一样（空格与换行规则不同）
- 格式化动词记不全，或 `%v` / `%+v` / `%#v` 混用
- 用 `+` 拼接错误字符串，而不是 `fmt.Errorf` 与 `%w` 包装
- 搞不清 stdout 和 stderr，导致管道或重定向时「日志进结果、结果进日志」
- 不知道 stdin/stdout 有缓冲，疑惑「为什么 printf 没立刻显示」或 `Scan` 读不到下一行

本篇建立在 [变量与常量](go-variables-types-and-constants.md) 之后——你已会基本类型与 `import`，现在把「把数据变成可读文本」以及「数据从哪来、写到哪去」这一层补上。后续错误处理、日志、HTTP 响应都会反复用到 `fmt`。

---

## 2. 核心概念

### 2.1 fmt 是什么

`fmt` 是标准库中的**格式化 I/O（formatted I/O）**包，实现 C 语言 `printf` 家族类似的 API，但更符合 Go 习惯：

| 能力 | 典型函数 | 输出目标 |
|------|----------|----------|
| 打印到标准输出 | `Print` / `Println` / `Printf` | `os.Stdout` |
| 格式化成字符串 | `Sprint` / `Sprintln` / `Sprintf` | 返回 `string` |
| 写入任意 Writer | `Fprint` / `Fprintln` / `Fprintf` | `io.Writer`（文件、缓冲区等） |
| 从 Reader 扫描 | `Scan` / `Scanln` / `Scanf` | 从 `os.Stdin` 或 `Scan*` 从字符串读 |
| 构造错误 | `Errorf` | 返回 `error` |

**命名规律：** 前缀 `Print` / `Sprint` / `Fprint` 表示目的地；后缀 `ln` 表示自动换行；后缀 `f` 表示带格式串。

### 2.2 打印三兄弟：Print / Println / Printf

```go
fmt.Print("a", "b")       // ab（无空格、无换行）
fmt.Println("a", "b")     // a b\n（参数间加空格，末尾换行）
fmt.Printf("%s-%s\n", "a", "b") // a-b\n（完全由格式串控制）
```

| 函数 | 参数间分隔 | 末尾换行 | 典型场景 |
|------|------------|----------|----------|
| `Print` | 无 | 无 | 连续输出片段 |
| `Println` | 空格 | 有 | 快速调试、多值一行 |
| `Printf` | 由格式串决定 | 由格式串决定 | 对齐、精度、自定义布局 |

`Println` 对非字符串类型会按**默认格式**打印（类似 `%v`），并保证每个参数之间有一个空格。

### 2.3 Sprint 与 Fprint：同一套逻辑，换输出目标

```go
msg := fmt.Sprintf("user=%s id=%d", name, id)  // 得到 string，常用于拼 SQL、JSON 片段、测试断言

fmt.Fprintf(os.Stderr, "warn: %v\n", err)    // 写到 stderr，不污染 stdout
```

- **`Sprint*`**：不打印，返回字符串——适合「先格式化再交给别的库」
- **`Fprint*`**：第一个参数是 `io.Writer`——适合写文件、`bytes.Buffer`、HTTP `ResponseWriter`；写诊断信息时用 `os.Stderr`（见 **2.9**）

### 2.4 格式动词（Format Verbs）速查

`Printf` / `Sprintf` / `Fprintf` / `Scanf` 使用 `%` 开头的动词。最常用如下：

| 动词 | 含义 | 示例 |
|------|------|------|
| `%v` | 默认格式 | `fmt.Printf("%v", x)` |
| `%+v` | 结构体字段名+值 | `fmt.Printf("%+v", user{Name:"a"})` |
| `%#v` | Go 语法风格表示 | 指针、结构体类型名等更明显 |
| `%T` | 类型 | `fmt.Printf("%T", 42)` → `int` |
| `%t` | 布尔 | `true` / `false` |
| `%d` | 十进制整数 | `42` |
| `%b` / `%o` / `%x` | 二/八/十六进制 | 见下方推导 |
| `%f` / `%e` / `%g` | 浮点 | 默认、科学计数、自动选短格式 |
| `%s` | 字符串或 `[]byte` | 原文 |
| `%q` | 带引号的字符串 | `"hello"`，不可见字符会转义 |
| `%p` | 指针地址 | `0xc0000140a0` |
| `%w` | **仅 Errorf**：包装底层 error | 见 2.7 |
| `%%` | 字面量 `%` | |

**同一数字 `42` 用不同进制动词（可复算）：**

先把 42 写成二进制（32+8+2 = 42）：

```
42₁₀ = 101010₂ = 52₈ = 2a₁₆
       ↑ 32+8+2    ↑ 5×8+2   ↑ 2×16+10
```

| 动词 | 输出 | 含义 |
|------|------|------|
| `%d` | `42` | 十进制 |
| `%b` | `101010` | 二进制 |
| `%o` | `52` | 八进制 |
| `%x` | `2a` | 十六进制（小写） |

**宽度与精度：**

```go
fmt.Printf("%5d\n", 7)      // "    7"  宽度 5，右对齐
fmt.Printf("%-5d|\n", 7)    // "7    |" 左对齐
fmt.Printf("%05d\n", 7)     // "00007"  前导零
fmt.Printf("%.2f\n", 3.14159) // "3.14"  保留 2 位小数
fmt.Printf("%8.2f\n", 3.14)   // 总宽 8，小数 2 位
```

### 2.5 默认格式与 Stringer

对任意值，`%v` 的默认行为大致为：

- 基本类型：人类可读形式
- 指针：若未实现 `Stringer`，打印地址；若指向的值实现了 `Stringer`，可能打印其 `String()`
- 结构体：`{字段值 ...}`，无字段名（要用 `%+v`）

实现 `fmt.Stringer` 可自定义打印：

```go
type user struct{ Name string }

func (u user) String() string {
    return "user:" + u.Name
}

fmt.Println(user{Name: "Ada"}) // user:Ada
```

`Println` / `Printf` 的 `%v` 在打印该类型时会调用 `String()`。**注意：** 不要在 `String()` 里再 `fmt.Print` 同一类型，否则会无限递归。

### 2.6 扫描输入：Scan / Scanln / Scanf

从标准输入（或 `Sscan` 从字符串）读入变量地址：

```go
var name string
var age int
fmt.Print("name age: ")
fmt.Scanln(&name, &age)
```

| 函数 | 行为 |
|------|------|
| `Scan` | 按空白分隔读入多个参数；换行也算空白 |
| `Scanln` | 读到换行结束；一行内多个值用空白分开 |
| `Scanf` | 按格式串解析，如 `"%s %d"` |

返回值 `(n int, err error)`：`n` 是成功赋值的参数个数。生产环境读复杂输入更常用 `bufio` + `strconv`；`Scan*` 适合学习与小工具。标准流与缓冲细节见 **2.9**。

### 2.7 Errorf 与 %w 错误包装

Go 1.13+ 推荐用 `fmt.Errorf` 构造错误，并用 `%w` 保留底层错误供 `errors.Is` / `errors.As` 使用：

```go
if err != nil {
    return fmt.Errorf("read config %s: %w", path, err)
}
```

| 写法 | 效果 |
|------|------|
| `fmt.Errorf("msg: %v", err)` | 仅字符串描述，**不能** `errors.Is` 匹配原错误 |
| `fmt.Errorf("msg: %w", err)` | 包装错误，可 unwrap / Is / As |

不要用 `errors.New(fmt.Sprintf(...))` 代替需要包装的场景。

### 2.8 fmt 与其他包的分工

| 需求 | 更合适的包 |
|------|------------|
| 程序内调试打印 | `fmt`（或后续学的 `log`） |
| 数字 ↔ 字符串转换 | `strconv`（`Atoi`、`Itoa`、`ParseFloat`） |
| 结构化日志 | `log/slog`、第三方 zap 等 |
| JSON / XML | `encoding/json` 等 |
| 高性能无反射拼接 | `strings.Builder`、少量场景 `strconv` |

`fmt` 适合**人类可读**输出；协议与持久化格式应用专用编码库。

### 2.9 标准输入输出（stdin / stdout / stderr）与缓冲区

`fmt` 的 `Print*` 默认写到 **stdout**，`Scan*` 默认从 **stdin** 读。理解这三个标准流和缓冲机制，能解释「管道里为什么看不到日志」「`Scan` 为什么吞掉下一行」等问题。

#### 2.9.1 三个标准流是什么

Go 通过 `os` 包暴露三个预打开的 `*os.File`，对应 Unix 文件描述符 0、1、2：

| 流 | Go 变量 | fd | 典型用途 |
|----|---------|-----|----------|
| 标准输入 | `os.Stdin` | 0 | 键盘、管道上游、重定向的文件 |
| 标准输出 | `os.Stdout` | 1 | 程序正常结果；`fmt.Print*` 默认写到这里 |
| 标准错误 | `os.Stderr` | 2 | 诊断、警告、进度；不应与「业务结果」混在一起 |

```go
fmt.Println("result")                          // → stdout
fmt.Fprintf(os.Stderr, "warn: %v\n", err)      // → stderr
fmt.Fscan(os.Stdin, &name)                     // ← stdin（与 Scan 等价，显式指定源）
```

**为什么要分 stdout 和 stderr？**

- **管道友好**：`myprog | grep ok` 只把 stdout 交给下游；stderr 仍显示在终端，调试信息不会污染管道数据
- **可分别重定向**：`myprog > out.txt 2> err.txt` 把正常输出和错误日志写到不同文件
- **约定**：stdout 放「可被下游程序消费的数据」；stderr 放「给人看的说明」

#### 2.9.2 Shell 重定向与管道（扩展）

在终端里，三个流可以被 shell 重定向（Go 程序无需改代码）：

```bash
go run . > result.txt          # 仅 stdout 进文件；stderr 仍在屏幕
go run . 2> debug.log          # 仅 stderr 进文件
go run . > out.txt 2>&1        # stdout 和 stderr 都进 out.txt
go run . 2>&1 | grep error     # 合并后再管道
echo "Alice 30" | go run . -mode=scan   # stdin 来自管道而非键盘
```

写 CLI 工具时：结构化结果（JSON、表格）走 **stdout**；日志、进度条、警告走 **stderr**，这样用户才能安全地 `result=$(mytool)` 或 `mytool | other`。

#### 2.9.3 缓冲区：为什么有时「迟迟不打印」

**缓冲区**是内存里暂存待写字节的一块区域，满或遇到换行/刷新时才真正写入内核或终端。

| 场景 | 常见缓冲策略 | 对 fmt 的影响 |
|------|--------------|---------------|
| 终端上的 stdout | 常为**行缓冲** | 带 `\n` 的 `Printf` 往往立刻可见 |
| 管道 / 重定向到文件 | 常为**全缓冲** | 无 `\n` 或缓冲未满时，输出可能延迟 |
| stderr | 通常**无缓冲或行缓冲** | 错误信息一般更快出现在屏幕上 |

`fmt` 包内部对 `os.Stdout` / `os.Stderr` 有包装，但本质仍受底层 `*os.File` 与是否 TTY 影响。若需要「立刻写出」（例如长时间运算中打进度），可以：

```go
fmt.Fprintf(os.Stderr, "step 1 done\n") // 换行有助于行缓冲刷新
// 或显式刷新（较少在 fmt 层做，更常见是对 *bufio.Writer）
os.Stderr.Sync()
```

生产环境更推荐 `log`（默认写 stderr）或 `log/slog`，它们会处理换行与一致性；本节重点是理解 **stdout ≠ stderr** 与 **缓冲导致的时间差**。

#### 2.9.4 fmt 与内存缓冲区：bytes.Buffer、strings.Builder

`Fprint*` 的第一个参数是 `io.Writer`，不限于文件——**内存里的缓冲区**同样可写：

```go
var buf bytes.Buffer
fmt.Fprintf(&buf, "id=%d name=%s", 1, "Bob")
s := buf.String() // "id=1 name=Bob"

// 仅拼 string、不需要 Write 接口时，strings.Builder 更轻
var b strings.Builder
fmt.Fprintf(&b, "score=%.1f", 98.5)
msg := b.String()
```

| 类型 | 实现 | 典型场景 |
|------|------|----------|
| `bytes.Buffer` | `io.Writer` + `io.Reader` | 测试里捕获输出、先格式化再 `Write` 到网络 |
| `strings.Builder` | 仅高效拼 `string` | 大量 `Sprintf` 式拼接，减少分配 |

测试示例：把 `fmt.Fprintf` 写到 `bytes.Buffer`，断言字符串内容，而不真的打印到终端。

#### 2.9.5 bufio：带缓冲的读写（与 Scan* 的分工）

`fmt.Scan*` 直接读 `os.Stdin`，对**交互式一行一行读**并不友好（见 4.6）。标准库 **`bufio`** 在 `io.Reader` / `io.Writer` 外再包一层缓冲，并提供了按行、按词扫描：

```go
reader := bufio.NewReader(os.Stdin)
line, _ := reader.ReadString('\n')   // 读一整行（含换行符）

scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    fmt.Println(scanner.Text())      // 按行处理
}

// 写文件时：先 bufio.Writer，减少 syscall
w := bufio.NewWriter(file)
fmt.Fprintf(w, "line %d\n", i)
w.Flush() // 别忘了 Flush，否则可能丢在缓冲区里
```

| 方式 | 适用 |
|------|------|
| `fmt.Scan` / `Scanln` / `Scanf` | 简单、格式固定的几个值；学习/demo |
| `bufio.Reader.ReadString` / `ReadBytes` | 读一整行再自己 `strings.Fields` / `strconv` 解析 |
| `bufio.Scanner` | 按行或按自定义分隔符扫描；大文件、日志处理 |
| `encoding/json` 等 | 结构化输入，而非空格分隔 |

**数据流关系（简图）：**

```
键盘/管道 → os.Stdin → bufio.Reader → 解析逻辑
                ↑
           fmt.Scan*（简单场景）

你的数据 → fmt.Fprintf → bytes.Buffer / bufio.Writer → 文件或 os.Stdout
```

#### 2.9.6 io.Writer / io.Reader：fmt 能对接的不仅是文件

`Fprint*` / `Fscanf` 依赖接口，因此同一套格式化 API 可写到多种目标：

| 目标 | 类型 | 示例 |
|------|------|------|
| 标准流 | `*os.File` | `os.Stdout`、`os.Stderr` |
| 文件 | `*os.File` | `f, _ := os.Create("out.txt"); fmt.Fprintln(f, "hi")` |
| 内存 | `*bytes.Buffer` | 测试、内存拼串 |
| HTTP 响应 | `http.ResponseWriter` | `fmt.Fprintf(w, "hello")`（也实现了 `Writer`） |
| 网络连接 | `net.Conn` | 部分场景直接写连接 |

**`Sscan*` 从字符串读**则是把 `string` 当作输入源，适合解析配置片段、测试数据，而不碰 stdin：

```go
var x, y int
fmt.Sscanf("3 14", "%d %d", &x, &y) // 从 string 解析，与 os.Stdin 无关
```

#### 2.9.7 与 log 包的分工（扩展）

| 包 | 默认输出 | 特点 |
|----|----------|------|
| `fmt` | stdout（Print*） | 无时间戳、无级别；适合程序「结果」或临时调试 |
| `log` | stderr | 自动时间戳、每条一行；适合简单运行日志 |
| `log/slog` | 可配置 | 结构化字段、级别；适合稍正式的服务 |

调试时用 `fmt.Println` 没问题；长期运行的服务里，诊断信息应走 **stderr** 或 `log`，避免和 stdout 上的 JSON/协议数据混在一起。

---

## 3. 动手实践

示例代码在 [`example/fmt/`](../example/fmt/)。

### 3.1 运行示例

```bash
cd example/fmt
go run .                      # 全部演示
go run . -mode=print          # Print / Println / Printf / Sprint / Fprintf
go run . -mode=format           # 格式动词与 Stringer
go run . -mode=scan             # Sscan 从字符串解析
go run . -mode=error            # Errorf 与 %w
```

预期（节选）：

```
--- Print / Println / Printf ---
Hello Go
line 1 true
name=Alice age=30
Sprintf: score=98.5
stderr: optional log
--- Format verbs ---
default %v: 42 3.14 hi
...
Stringer: Carol(28)
```

### 3.2 跟着改：Printf 对齐

在 `demoFormat` 中增加：

```go
fmt.Printf("|%10s|%10s|\n", "id", "name")
fmt.Printf("|%10d|%10s|\n", 1, "Alice")
```

预期：列宽 10，数字右对齐，便于扫一眼对齐的表格。

### 3.3 跟着改：从字符串扫描

```go
var a, b int
fmt.Sscanf("10 20 30", "%d %d", &a, &b)
fmt.Println(a, b) // 10 20（第三个数被忽略）
```

### 3.4 跟着改：错误包装

```go
base := fmt.Errorf("connection refused")
wrapped := fmt.Errorf("dial api: %w", base)
fmt.Println(errors.Is(wrapped, base)) // true（需 import "errors"）
```

### 3.5 跟着改：重定向与 stderr

```bash
go run . -mode=print > /tmp/out.txt 2> /tmp/err.txt
cat /tmp/out.txt    # 正常 Print 输出
cat /tmp/err.txt    # Fprintf(os.Stderr, ...) 的内容
```

体会：stdout 与 stderr 可以分别重定向，互不干扰。

### 3.6 跟着改：用 bytes.Buffer 捕获格式化结果

```go
var buf bytes.Buffer
fmt.Fprintf(&buf, "sum=%d", 1+2)
fmt.Println(buf.String()) // sum=3
```

适合单测里断言输出，而不污染终端。

### 3.7 自检清单

- [ ] 能说出 `Print`、`Println`、`Printf` 的区别
- [ ] 能解释 `%v`、`%+v`、`%#v`、`%q` 的差异
- [ ] 会用 `Sprintf` 生成字符串、`Fprintf` 写到 `os.Stderr`
- [ ] 能区分 `os.Stdin`、`os.Stdout`、`os.Stderr` 及各自典型用途
- [ ] 知道管道只传递 stdout，诊断信息应写 stderr
- [ ] 知道 `bytes.Buffer` / `bufio` 与 `fmt.Fprint*` 的关系
- [ ] 会用 `fmt.Errorf("...: %w", err)` 包装错误
- [ ] `go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 Printf 少参数或多参数

```go
fmt.Printf("%s %d\n", "only") // %!d(MISSING) 或运行时格式错误
```

**修复：** 格式串占位符数量与后面参数一致；复杂时先 `Sprintf` 赋给变量再打印，便于单测。

### 4.2 用 Println 以为会自定义格式

```go
fmt.Println(pi) // 默认精度，不是两位小数
```

**修复：** 需要小数位用 `Printf("%.2f", pi)`。

### 4.3 `%s` 打印 `[]byte` 与 `string` 的混淆

`%s` 对 `[]byte` 当文本；二进制数据应 `%x` 或 `%q`，否则可能打出乱码。

### 4.4 在 String() 里格式化自身

```go
func (u user) String() string {
    return fmt.Sprintf("%v", u) // 可能再次调用 String() → 栈溢出
}
```

**修复：** 只拼字段，如 `return u.Name + "(" + strconv.Itoa(u.Age) + ")"`。

### 4.5 用 %v 包装错误导致无法 errors.Is

```go
return fmt.Errorf("failed: %v", err) // 底层 err 被当成字符串
```

**修复：** 需要链式判断时用 `%w`：`fmt.Errorf("failed: %w", err)`。

### 4.6 Scan 留下换行符或读不全

交互程序里 `Scan` 与 `Scanln` 混用可能导致下一行读入空串。

**修复：** 同一流程统一用一种；读一整行用 `bufio.NewReader(os.Stdin).ReadString('\n')` 再解析。

### 4.7 把日志打到 stdout，破坏管道

```go
fmt.Println("debug: loading...") // 进了 stdout
// 用户执行：mytool | jq .  → jq 解析失败，因为 stdout 混入了非 JSON 文本
```

**修复：** 诊断、进度、警告用 `fmt.Fprintf(os.Stderr, ...)` 或 `log.Println`。

### 4.8 重定向到文件后「看不到输出」

程序 `fmt.Print("no newline")` 后立刻崩溃或长时间阻塞，重定向到文件时缓冲未刷新，文件为空或内容滞后。

**修复：** 行尾加 `\n`；或对 `bufio.Writer` 调用 `Flush()`；关键路径用 `log`（通常带换行）。

### 4.9 误把 stderr 当「只能写错误」

stderr 语义是 **diagnostic output（诊断输出）**，不仅是 `error` 类型——进度、警告、调试 trace 都应放 stderr，把 stdout 留给「程序交付物」。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. `Print` / `Println` / `Printf` 输出到 stdout；`Sprint*` 得字符串；`Fprint*` 写任意 `io.Writer`
2. `Scan*` 从 stdin 读；`Sscan*` 从 string 读——与标准流解耦，便于测试
3. **stdout** 放程序结果（可被管道消费）；**stderr** 放诊断/日志；**stdin** 是默认输入源
4. 终端、管道、文件上的**缓冲策略不同**，可能出现输出延迟；`bufio.Writer` 写完记得 `Flush`
5. `bytes.Buffer`、`strings.Builder` 可作 `Fprint*` 的内存目标；交互读行优先 `bufio`
6. 日常调试 `Println` 最快；要布局、精度用 `Printf` / `Sprintf`
7. `%v` 通用；结构体调试常用 `%+v`；`%#v` 接近 Go 源码风格；`%q` 看不可见字符
8. 实现 `String() string` 可定制类型的打印表现
9. 构造可 unwrap 的错误用 `fmt.Errorf("...: %w", err)`，配合 `errors.Is` / `errors.As`

**官方文档：**

- [pkg.go.dev/fmt](https://pkg.go.dev/fmt)
- [pkg.go.dev/os#Stdin](https://pkg.go.dev/os#Stdin) — 标准流
- [pkg.go.dev/bufio](https://pkg.go.dev/bufio) — 带缓冲 I/O
- [pkg.go.dev/bytes#Buffer](https://pkg.go.dev/bytes#Buffer) — 内存缓冲区
- [A Tour of Go：Formatting strings](https://go.dev/tour/methods/17)
- [Go 1.13：错误包装](https://go.dev/blog/go1.13-errors)

**与本仓库的关系：**

- 上一篇：[Go 变量、类型与常量](go-variables-types-and-constants.md)
- 示例：[`example/fmt/`](../example/fmt/)
- 相关：错误处理、日志（待写）
