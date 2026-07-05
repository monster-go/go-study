# Go 字符串操作：拼接、标准库与 Unicode

> 知识点总结：掌握 Go 字符串的不可变语义、`+` 与 `strings.Builder` 拼接、`strings`/`strconv` 常用 API，以及按 **rune** 处理 Unicode 文本；能正确区分**字节长度**与**字符个数**。

---

## 1. 为什么需要了解这个

字符串几乎出现在每个 Go 程序里：配置、日志、HTTP、JSON。新人常见困惑：

- 用 `len(s)` 统计「字符数」，中文环境下结果偏大（`len` 是**字节数**）
- 大量 `+` 拼接导致性能差、内存分配多
- 用 `string(65)` 以为得到 `"65"`，实际得到 `"A"`（Unicode 码点）
- 不知道 `strings` 与 `strconv` 的分工，重复造轮子
- 修改字符串用 `s[0] = 'x'` 编译失败（字符串**不可变**）

本篇建立在 [变量与常量](go-variables-types-and-constants.md) 与 [运算符](go-operators.md) 之后。后续控制流、切片、JSON 编码都会处理字符串。

---

## 2. 核心概念

### 2.1 字符串是什么

Go 的 `string` 是**只读的字节序列**，默认按 **UTF-8** 编码存储 Unicode 文本：

| 特性 | 说明 |
|------|------|
| 不可变 | 不能修改某个字节；「修改」需生成新字符串 |
| `len(s)` | 返回**字节数**，不是 rune 个数 |
| 索引 `s[i]` | 得到 `byte`（`uint8`），不是完整 Unicode 字符 |
| 遍历 | `for i, r := range s` 按 **rune**（码点）遍历 |

```go
s := "Go语言"
len(s)           // 8（字节）
len([]rune(s))   // 4（字符：G,o,语,言）
```

### 2.2 拼接方式对比

| 方式 | 适用场景 | 注意 |
|------|----------|------|
| `+` | 少量、简单拼接 | 每次 `+` 可能分配新字符串 |
| `fmt.Sprintf` | 需要格式化 | 有反射开销，非热点可用 |
| `strings.Builder` | 循环中大量拼接 | 推荐；内部扩容，最终 `String()` |
| `strings.Join` | 已有 `[]string` | 指定分隔符一次 Join |
| `bytes.Buffer` | 需要同时 `Write` 字节 | 也可 `WriteString` |

```go
// 少量
msg := "hello" + " " + "world"

// 循环拼接推荐
var b strings.Builder
for _, w := range words {
    b.WriteString(w)
}
result := b.String()
```

### 2.3 strings 包常用函数

| 函数 | 作用 | 示例 |
|------|------|------|
| `Contains` / `HasPrefix` / `HasSuffix` | 子串判断 | `strings.Contains(s, "go")` |
| `Index` / `LastIndex` | 查找位置 | 找不到返回 `-1` |
| `Split` / `SplitN` | 分割 | `strings.Split("a,b", ",")` |
| `Join` | 连接 | `strings.Join([]string{"a","b"}, "-")` |
| `Trim` / `TrimSpace` / `TrimPrefix` | 去空白或前后缀 | `strings.TrimSpace(s)` |
| `Replace` / `ReplaceAll` | 替换 | `ReplaceAll(s, "old", "new")` |
| `ToLower` / `ToUpper` | 大小写 | 注意 Unicode  locale |
| `Compare` | 字典序比较 | 相等返回 `0` |

完整列表见 [pkg.go.dev/strings](https://pkg.go.dev/strings)。

### 2.4 strconv：字符串与数字互转

| 函数 | 方向 | 示例 |
|------|------|------|
| `Atoi` / `ParseInt` | string → int | `strconv.Atoi("42")` |
| `Itoa` / `FormatInt` | int → string | `strconv.Itoa(42)` |
| `ParseFloat` / `FormatFloat` | float ↔ string | 注意 `bitSize` 参数 |
| `FormatBool` / `ParseBool` | bool ↔ string | |
| `Quote` / `Unquote` | 带转义的字符串 | 调试输出 |

**务必检查 `error`：**

```go
n, err := strconv.Atoi(input)
if err != nil {
    return fmt.Errorf("invalid number: %w", err)
}
```

不要用 `string(123)` 把数字变字符串——那是把码点 `123` 转成字符 `{`。

### 2.5 rune、byte 与 range

| 类型 | 含义 |
|------|------|
| `byte` | `uint8` 别名，一个字节 |
| `rune` | `int32` 别名，一个 Unicode 码点 |

```go
for i, r := range "沙河" {
    fmt.Printf("%d: %c\n", i, r)
}
// i 是字节偏移，不是「第几个字符」；中文每字占 3 字节 UTF-8
```

判断汉字等脚本可用 `unicode` 包：

```go
unicode.Is(unicode.Han, r)
```

### 2.6 string 与 []byte 转换

```go
b := []byte("hello")   // 拷贝一份字节
s := string(b)         // 拷贝为 string
```

转换会**复制数据**（除非编译器优化某些场景）。修改 `b` 不影响 `s`。

需要原地处理二进制或 IO 时用 `[]byte`；业务文本用 `string`。

---

## 3. 动手实践

示例代码在 [`example/strings/`](../example/strings/)。

### 3.1 运行示例

```bash
cd example/strings
go run .                    # 全部演示
go run . -mode=concat       # + 与 Builder
go run . -mode=strings      # Split/Join/Trim 等
go run . -mode=strconv      # 数字转换
go run . -mode=rune         # len 与 range
```

预期（节选）：

```
--- strings 包 ---
Split: [hello go world]
--- strconv 转换 ---
Atoi: 42 <nil>
--- 遍历 rune ---
len(s) 字节数: 8
rune 数: 4
```

### 3.2 跟着改：统计汉字（复习 day01）

```go
s := "hello沙河小王子"
count := 0
for _, r := range s {
    if unicode.Is(unicode.Han, r) {
        count++
    }
}
fmt.Println(count) // 5
```

### 3.3 跟着改：Split 解析 CSV 行

```go
line := "name,age,city"
fields := strings.Split(line, ",")
fmt.Println(fields[0], fields[1])
```

### 3.4 自检清单

- [ ] 能解释 `len("中文")` 为何不是 2
- [ ] 知道循环拼接优先 `strings.Builder`
- [ ] 会用 `strconv.Atoi` 并处理 `error`
- [ ] 会用 `range` 遍历字符串中的 rune
- [ ] `go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 用 len 当字符数

```go
fmt.Println(len("你好")) // 6，不是 2
```

**修复：** `len([]rune(s))` 或 `utf8.RuneCountInString(s)`。

### 4.2 string(数字) 的误解

```go
fmt.Println(string(65)) // "A"，不是 "65"
```

**修复：** `strconv.Itoa(65)` 或 `fmt.Sprintf("%d", 65)`。

### 4.3 大量 + 拼接

```go
var s string
for _, v := range items {
    s += v // 每次可能分配新底层数组
}
```

**修复：** `strings.Builder` 或 `strings.Join`。

### 4.4 用字节索引截断 UTF-8

```go
s := "你好世界"
bad := s[:3] // 可能截断半个汉字，产生非法 UTF-8
```

**修复：** 按 rune 处理：`string([]rune(s)[:1])`，或 `range` 计数后再切。

### 4.5 忽略 strconv 的 error

```go
n, _ := strconv.Atoi(userInput) // 非法输入时 n=0，难以区分
```

**修复：** 始终检查 `err`，向用户返回明确错误信息。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. `string` 不可变；`len` 是字节数，中文须用 rune 统计字符
2. 少量拼接用 `+`；循环拼接用 `strings.Builder`；已有切片用 `Join`
3. 数字 ↔ 字符串用 `strconv`，不用 `string(int)`
4. `for _, r := range s` 按 Unicode 码点遍历
5. `strings` 负责文本处理，`strconv` 负责类型转换

**官方文档：**

- [pkg.go.dev/strings](https://pkg.go.dev/strings)
- [pkg.go.dev/strconv](https://pkg.go.dev/strconv)
- [pkg.go.dev/unicode](https://pkg.go.dev/unicode)
- [A Tour of Go：Strings](https://go.dev/tour/basics/13)

**与本仓库的关系：**

- 上一篇：[Go 运算符](go-operators.md)
- 示例：[`example/strings/`](../example/strings/)、[`example/day01/`](../example/day01/)（汉字统计练习）
- 下一篇：[Go 语句控制](go-control-flow.md)
