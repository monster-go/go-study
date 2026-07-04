# Go fmt 包：格式化输入输出

> 知识点总结：掌握 `fmt` 的打印、格式化字符串、扫描输入与 `Errorf` 错误包装，能根据场景选用 `Print`/`Printf`/`Sprintf`/`Fprintf` 及常用动词（`%v`、`%d`、`%s`、`%w` 等）。

---

## 1. 为什么需要了解这个

几乎每个 Go 程序都会用到 `fmt`：调试输出、拼日志、读用户输入、构造错误信息。新人常见困惑包括：

- 分不清 `Print`、`Println`、`Printf` 何时用哪个
- 以为 `fmt.Println(a, b)` 和 `fmt.Printf("%v %v", a, b)` 完全一样（空格与换行规则不同）
- 格式化动词记不全，或 `%v` / `%+v` / `%#v` 混用
- 用 `+` 拼接错误字符串，而不是 `fmt.Errorf` 与 `%w` 包装

本篇建立在 [变量与常量](go-variables-types-and-constants.md) 之后——你已会基本类型与 `import`，现在把「把数据变成可读文本」这一层补上。后续错误处理、日志、HTTP 响应都会反复用到 `fmt`。

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
- **`Fprint*`**：第一个参数是 `io.Writer`——适合写文件、`bytes.Buffer`、HTTP `ResponseWriter`

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
| `%b` / `%o` / `%x` | 二/八/十六进制 | `101010`、`52`、`2a` |
| `%f` / `%e` / `%g` | 浮点 | 默认、科学计数、自动选短格式 |
| `%s` | 字符串或 `[]byte` | 原文 |
| `%q` | 带引号的字符串 | `"hello"`，不可见字符会转义 |
| `%p` | 指针地址 | `0xc0000140a0` |
| `%w` | **仅 Errorf**：包装底层 error | 见 2.7 |
| `%%` | 字面量 `%` | |

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

返回值 `(n int, err error)`：`n` 是成功赋值的参数个数。生产环境读复杂输入更常用 `bufio` + `strconv`；`Scan*` 适合学习与小工具。

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

### 3.5 自检清单

- [ ] 能说出 `Print`、`Println`、`Printf` 的区别
- [ ] 能解释 `%v`、`%+v`、`%#v`、`%q` 的差异
- [ ] 会用 `Sprintf` 生成字符串、`Fprintf` 写到 `os.Stderr`
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

---

## 5. 小结与延伸阅读

**要点回顾：**

1. `Print` / `Println` / `Printf` 输出到 stdout；`Sprint*` 得字符串；`Fprint*` 写 `io.Writer`
2. 日常调试 `Println` 最快；要布局、精度用 `Printf` / `Sprintf`
3. `%v` 通用；结构体调试常用 `%+v`；`%#v` 接近 Go 源码风格；`%q` 看不可见字符
4. 实现 `String() string` 可定制类型的打印表现
5. 构造可 unwrap 的错误用 `fmt.Errorf("...: %w", err)`，配合 `errors.Is` / `errors.As`

**官方文档：**

- [pkg.go.dev/fmt](https://pkg.go.dev/fmt)
- [A Tour of Go：Formatting strings](https://go.dev/tour/methods/17)
- [Go 1.13：错误包装](https://go.dev/blog/go1.13-errors)

**与本仓库的关系：**

- 上一篇：[Go 变量、类型与常量](go-variables-types-and-constants.md)
- 示例：[`example/fmt/`](../example/fmt/)
- 相关：错误处理、日志（待写）
