# Go 变量、类型与常量：从声明到 iota

> 知识点总结：掌握 Go 中变量的声明与赋值方式、基本类型与零值、常量及 iota 枚举写法，并能在小示例中正确选用 `var`、`:=` 与 `const`。

---

## 1. 为什么需要了解这个

变量与类型是几乎所有 Go 程序的起点。新人常在这些地方卡住：

- 分不清 `var`、`=`、`:=` 各自用在什么场景
- 以为 Go 有「隐式类型转换」，写出 `int` 与 `float64` 直接运算的代码
- 把 `iota` 当成「自增变量」，在 `const` 块外使用或误解其重置规则
- 忽略**零值（zero value）**，以为未赋值的变量是 `nil` 或随机数

本篇是「语法基础」的第一步，建立在 [安装与入门](go-install-and-getting-started.md) 之后。后续控制流、函数、结构体都会反复用到这里的概念。

---

## 2. 核心概念

### 2.1 变量声明（Variable Declaration）

Go 是**静态类型**语言：每个变量在编译期就有确定类型，不能随意改变。

| 写法 | 语法 | 典型场景 |
|------|------|----------|
| **完整声明** | `var name type = value` | 包级变量、需要显式类型、先声明后赋值 |
| **类型推断** | `var name = value` | 类型由右侧表达式推断 |
| **仅声明** | `var name type` | 先占位，稍后赋值；会得到该类型的零值 |
| **短变量声明** | `name := value` | **函数内**最常用；声明并初始化 |
| **多变量** | `var a, b int = 1, 2` 或 `x, y := 1, 2` | 一行声明多个变量 |

**重要规则：**

- `:=` **只能在函数内**使用，不能用于包级（文件顶层）
- `:=` 左侧至少有一个**新变量名**，否则应改用 `=`
- 已声明的变量必须被使用，否则编译报错：`declared and not used`

```go
var count int           // 零值 0
var msg = "hello"       // 推断为 string
name := "Go"            // 短声明，仅在函数内

var a, b = 1, "two"     // a 为 int，b 为 string
x, y := 10, 20
x, err := doSomething() // 常见：返回值 + error
```

### 2.2 赋值（Assignment）

`:=` 是「声明 + 赋值」；对已存在变量只做赋值时用 `=`：

```go
count = 42
x, y = y, x   // 交换，无需临时变量
```

**多重赋值**可同时更新多个变量，函数也常返回多值：

```go
min, max := 1, 100
min, max = max, min
```

### 2.3 零值（Zero Value）

只声明、不赋初值时，Go 会给**类型对应的零值**，不是「未定义」：

| 类型 | 零值 |
|------|------|
| 数值（int、float 等） | `0` |
| `bool` | `false` |
| `string` | `""`（空字符串） |
| 指针、slice、map、channel、function、interface | `nil` |

```go
var n int
var ok bool
var s string
// n==0, ok==false, s==""
```

### 2.4 基本类型（Basic Types）

Go 内置类型分几大类（无需 import）：

| 类别 | 类型 | 说明 |
|------|------|------|
| 布尔 | `bool` | `true` / `false` |
| 字符串 | `string` | UTF-8 字节序列，不可变 |
| 有符号整数 | `int`, `int8`, `int16`, `int32`, `int64` | `int` 长度随平台（32 或 64 位） |
| 无符号整数 | `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` | `uint8` 即 **byte** 的别名 |
| 浮点 | `float32`, `float64` | 默认浮点字面量为 `float64` |
| 复数 | `complex64`, `complex128` | 较少用，知道即可 |
| 字符 | `rune`（`int32` 别名） | 表示 Unicode 码点；**byte** 是 `uint8` 别名 |

**习惯：**

- 整数默认用 `int`，除非有明确大小或二进制协议需求
- 浮点默认用 `float64`
- 文本处理用 `string`；单个 Unicode 字符用 `rune`；原始字节用 `byte`

### 2.5 类型转换（Type Conversion）

Go **没有** C/Java 那种隐式数值转换，必须**显式转换**：

```go
var i int = 42
var f float64 = float64(i)
var u uint = uint(f)   // 可能丢失精度，需自己负责
```

字符串与数字互转需用 `strconv` 等标准库，不能 `string(123)` 当数字用（那是 Unicode 码点）。

### 2.6 常量（Constant）

常量在**编译期**确定，运行时不能修改：

```go
const Pi = 3.14159
const Greeting = "hello"

const (
    MaxRetries = 3
    Timeout    = 30
)
```

| 概念 | 说明 |
|------|------|
| **typed constant** | `const n int = 10`，类型固定 |
| **untyped constant** | `const n = 10`，像「高精度数」，参与运算时再落到具体类型 |
| **可使用的类型** | 布尔、字符串、数值；**不能**是 slice、map、struct 等复合类型 |

无类型常量更灵活，例如：

```go
const x = 1
var i int = x
var f float64 = x   // 同一个 x 可赋给不同数值类型
```

### 2.7 iota：常量组内的行号计数器

`iota` 只能出现在 **`const` 常量组**里，表示当前行在组内的索引，**从 0 开始**，每行自动 +1：

```go
const (
    Sunday = iota // 0
    Monday        // 1
    Tuesday       // 2
)
```

**每个新的 `const (` 块都会让 iota 重新从 0 计数。**

常见用法：

```go
// 1. 枚举状态码
const (
    StatusOK = iota
    StatusError
    StatusPending
)

// 2. 跳过某些值：用 _ 占位
const (
    _  = iota             // 丢弃 0
    KB = 1 << (10 * iota) // 1<<10, 1<<20, ...
    MB
    GB
)

// 3. 同一行多个常量共享 iota
const (
    Read, Write, Execute = iota, iota, iota // 都是 0？不对——
)
// 实际上每行 iota 只递增一次，同一行多个 iota 取值相同：
const (
    A, B = iota, iota // A=0, B=0
    C, D              // C=1, D=1
)
```

**iota 不是变量**，不能 `iota++`，也不能在 `const` 外使用。

---

## 3. 动手实践

示例代码在 [`example/vars/`](../example/vars/)。

### 3.1 运行示例

```bash
cd example/vars
go run .                    # 运行全部演示
go run . -mode=zero         # 只看零值
go run . -mode=iota         # 只看 iota 枚举
```

### 3.2 跟着改：声明与赋值

打开 `main.go`，在 `main` 中尝试：

```go
var total int
total = 100
label := "score"
// label = 200        // 取消注释：类型不匹配，编译失败
fmt.Println(total, label)
```

预期：`100 score`；若把 `label = 200` 取消注释，应看到 `cannot use 200 (untyped int constant) as string`。

### 3.3 跟着改：类型转换

```go
a, b := 3, 4
// fmt.Println(a / b)           // 整数除法 → 0
fmt.Println(float64(a) / float64(b)) // → 0.75
```

### 3.4 跟着改：iota

查看 `constants.go` 中的 `Weekday` 与 `FilePerm`，在 `main` 里打印：

```go
fmt.Println(Monday, Tuesday, ReadPerm)
```

预期：`1 2 4`（具体值取决于常量定义，以代码为准）。

### 3.5 自检清单

- [ ] 能解释 `var`、`=`、`:=` 的区别
- [ ] 能说出 `int`、`string`、`bool` 的零值
- [ ] 示例 `go run .` 无报错
- [ ] 能独立写出一个带 `iota` 的 3 项枚举常量组

---

## 4. 常见坑与排查

### 4.1 在包级使用 `:=`

```go
// 文件顶层
count := 0   // 编译错误：syntax error: non-declaration statement outside function body
```

**修复：** 包级改用 `var count = 0` 或 `var count int`。

### 4.2 `:=` 左侧没有新变量

```go
x := 1
x := 2   // 编译错误：no new variables on left side of :=
```

**修复：** 第二次写 `x = 2`。若多返回值中只有一个新名字，如 `file, err := os.Open(...)` 之后 `file, err := ...` 也报错，应写 `file, err = ...`。

### 4.3 隐式类型转换不存在

```go
var i int = 10
var f float64 = i   // 编译错误
```

**修复：** `var f float64 = float64(i)`。

### 4.4 误以为 iota 会「跨 const 块连续递增」

```go
const A = iota // 0
const B = iota // 又是 0，不是 1
```

**修复：** 需要连续枚举时，放在**同一个** `const (` 块内。

### 4.5 用 `string(n)` 把数字转成字符串

```go
s := string(65) // 得到 "A"（Unicode 65），不是 "65"
```

**修复：** 数字转字符串用 `strconv.Itoa(65)` 或 `fmt.Sprintf("%d", 65)`。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. 函数内优先 `:=`；包级或需零值占位用 `var`；仅赋值用 `=`
2. 未初始化变量有**零值**，不是随机数
3. Go 必须**显式类型转换**，无隐式数值提升
4. `const` 在编译期固定；`iota` 只在 `const` 组内按行递增，新块重置为 0
5. 日常整数用 `int`，浮点用 `float64`，字符用 `rune`，字节用 `byte`

**官方文档：**

- [A Tour of Go：Basics](https://go.dev/tour/basics/1)
- [Effective Go：Declarations](https://go.dev/doc/effective_go#declarations)
- [Go 语言规范：常量](https://go.dev/ref/spec#Constants)
- [Go 语言规范：iota](https://go.dev/ref/spec#Iota)

**与本仓库的关系：**

- 上一篇：[Go 安装与入门](go-install-and-getting-started.md)
- 示例：[`example/vars/`](../example/vars/)
- 下一篇（计划）：控制流（if / for / switch）
