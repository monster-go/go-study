# Go 接口与错误处理：从 error 惯用法到隐式实现

> 知识点总结：掌握 Go 的 `(T, error)` 错误处理惯用法、`fmt.Errorf` 与 `%w` 包装、`errors.Is` / `errors.As` 判断与解包；理解接口（Interface）的隐式实现、空接口 `any`、类型断言与 type switch，以及「typed nil 导致 `i == nil` 为 false」的常见陷阱。

---

## 1. 为什么需要了解这个

函数与结构体解决「怎么组织代码与数据」，接口与错误处理解决「怎么抽象行为」与「怎么表达失败」。新人常在这些地方卡住：

- 看到满屏 `if err != nil` 觉得啰嗦，想用 try/catch 或忽略错误
- 分不清 `errors.New`、`fmt.Errorf`、`%v` 与 `%w` 该用哪个
- 包装多层错误后，`==` 比较 sentinel 失败，不知道用 `errors.Is`
- 以为 Java 式 `implements Shape`——Go **隐式满足**接口，方法集对不上才编译报错
- 值接收者与指针接收者导致「明明实现了方法却赋不进接口」
- `var i any = (*T)(nil)` 时 `i == nil` 为 **false**，排查 nil 判断 bug 很久

本篇建立在 [函数、结构体与作用域](go-functions-structs-and-scope.md) 之后。错误处理贯穿所有 I/O 与业务逻辑；接口是 Go 多态与标准库设计的核心（如 `io.Reader`、`fmt.Stringer`）。后续并发、HTTP、测试都会反复用到。

---

## 2. 核心概念

### 2.1 error 是什么

`error` 是内置的**接口类型**：

```go
type error interface {
    Error() string
}
```

任何实现了 `Error() string` 方法的类型都是 `error`。常见创建方式：

| 方式 | 示例 | 适用 |
|------|------|------|
| `errors.New` | `errors.New("not found")` | 固定文案、sentinel 错误 |
| `fmt.Errorf` | `fmt.Errorf("id=%d invalid", id)` | 带格式化上下文 |
| 自定义类型 | `type MyErr struct { ... }` + `Error()` | 需携带字段、区分错误种类 |

**惯用法：** 可恢复的业务失败用 `return ..., err`；不可恢复的编程错误才考虑 `panic`（见上一篇 defer/recover）。

```go
func Div(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为 0")
    }
    return a / b, nil
}

v, err := Div(10, 0)
if err != nil {
    // 处理错误：打日志、返回上层、降级等
    return err
}
```

**原理：** Go 没有异常控制流；错误是**普通返回值**，调用方必须显式处理。编译器不强制检查 `err`（与 Rust `Result` 不同），这是语言设计取舍，团队靠 code review 与 lint 约束。

### 2.2 Sentinel 错误与 errors.Is

**Sentinel error（哨兵错误）** 是包级变量，用 `==` 或 `errors.Is` 识别特定失败：

```go
var ErrNotFound = errors.New("not found")

func Find(id int) (string, error) {
    if id == 404 {
        return "", ErrNotFound
    }
    return "ok", nil
}
```

调用方：

```go
_, err := Find(404)
if errors.Is(err, ErrNotFound) {
    // 资源不存在，可返回 404
}
```

**为什么需要 `errors.Is`（Go 1.13+）：** 中间层用 `%w` 包装后，`err == ErrNotFound` 往往为 **false**，因为外层是新的 error 值。`errors.Is` 会沿包装链向下比较。

```
Find 返回 ErrNotFound
    ↓ 某层 fmt.Errorf("load user: %w", err)
包装后的 err == ErrNotFound  → false
errors.Is(err, ErrNotFound)  → true（递归 Unwrap）
```

### 2.3 错误包装：%w 与 errors.Unwrap

```go
if err != nil {
    return fmt.Errorf("open config: %w", err)
}
```

| 动词 | 效果 |
|------|------|
| `%w` | 包装底层 error，支持 `Unwrap` / `Is` / `As` |
| `%v` | 仅把错误**文字**拼进字符串，**丢失**包装链 |

`errors.Unwrap(err)` 返回被 `%w` 包的那一层；`Is` / `As` 内部会反复 Unwrap。

**示例推导：**

```go
base := strconv.ErrSyntax
wrapped := fmt.Errorf("parse: %w", base)

errors.Is(wrapped, strconv.ErrSyntax)  // true
wrapped == strconv.ErrSyntax           // false
```

### 2.4 errors.As：按类型提取

当错误是**自定义结构体**且需读字段时，用 `errors.As`：

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return e.Field + ": " + e.Msg
}

var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println("字段:", ve.Field)
}
```

**原理：** `As` 沿包装链查找，若某层 error 的**动态类型**可赋给 `&ve` 指向的类型，则赋值并返回 `true`。与 `Is`（比「值是否相等」）互补。

### 2.5 接口（Interface）与隐式实现

接口定义**方法集（method set）**；类型只要实现了这些方法，就**自动**满足接口，无需声明 `implements`：

```go
type Shape interface {
    Area() float64
}

type Circle struct{ R float64 }

func (c Circle) Area() float64 {
    return 3.14159 * c.R * c.R
}

var s Shape = Circle{R: 2} // OK
```

| 概念 | 说明 |
|------|------|
| 接口值 | 动态类型（具体类型）+ 动态值（该类型的数据） |
| 隐式实现 | 编译期检查方法集是否完整 |
| 小接口 | Go 倾向 `Read(p []byte) (n int, err error)` 这类窄接口，组合使用 |

#### 值接收者 vs 指针接收者与接口

若接口方法集含 `Area()`，而实现是 `func (c *Circle) Area()`，则：

- `*Circle` 满足 `Shape`
- `Circle`（值）**不一定**满足——值类型方法集不含指针接收者方法

```go
var s Shape
s = Circle{R: 1}   // 若 Area 在 *Circle 上 → 编译错误
s = &Circle{R: 1}  // OK
```

**原理：** 方法集规则——类型 `T` 的方法集含值接收者方法；`*T` 的方法集含 `T` 与 `*T` 的全部方法。接口赋值时编译器检查**实际类型**的方法集。

### 2.6 空接口 any 与多态参数

`any` 是 `interface{}` 的别名，表示「任意类型」：

```go
func Print(v any) {
    fmt.Println(v)
}
```

标准库中 `fmt.Println`、`json.Marshal` 等大量使用 `any`。代价是失去静态类型检查，需 type assertion 或 type switch 恢复具体类型。

### 2.7 类型断言（Type Assertion）

从接口值取出具体类型：

```go
var s Shape = Circle{R: 2}

c, ok := s.(Circle)   // 安全断言：失败时 ok=false，不 panic
if ok {
    fmt.Println(c.R)
}

// s.(Rect)            // 若类型不对 → panic
```

**原理：** 断言检查接口内的**动态类型**是否与目标类型一致（或可赋值）。

### 2.8 type switch

```go
func Describe(v any) string {
    switch x := v.(type) {
    case int:
        return fmt.Sprintf("int=%d", x)
    case string:
        return fmt.Sprintf("string=%q", x)
    case Shape:
        return fmt.Sprintf("area=%.2f", x.Area())
    default:
        return fmt.Sprintf("%T", x)
    }
}
```

`x` 在对应 `case` 块内已是具体类型，可直调方法。

### 2.9 接口组合

接口可嵌入其他接口，形成更大方法集：

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

type ReadCloser interface {
    Reader
    Closer
}
```

与 struct 匿名字段嵌入类似，是标准库 `io` 包的组织方式。

### 2.10 nil 接口陷阱（typed nil）

```go
var p *Person = nil
var i any = p

fmt.Println(p == nil) // true
fmt.Println(i == nil) // false !
```

**原理：** 接口值由 `(type, value)` 二元组构成。`i = p` 时 type 是 `*Person`，value 是 nil 指针——**接口「有类型」**，故 `i == nil` 为 false。只有 type 与 value 均为「无」时，接口才等于 nil。

排查：用反射或 `i == nil` 前先判断 `p == nil`；函数返回 `( *T, error)` 时避免返回 `(nil, nil)` 与 `(typed nil, err)` 混用导致调用方逻辑混乱。

---

## 3. 动手实践

### 3.1 运行仓库示例

```bash
cd example/errors-interfaces
go run .                      # 运行全部 demo
go run . -mode error          # 基础 error 与 errors.Is
go run . -mode wrap           # %w 包装与 errors.As
go run . -mode interface      # Shape 接口与多态
go run . -mode assert         # 类型断言与 type switch
go run . -mode niliface       # typed nil 与 i==nil
```

**预期（`-mode all` 片段）：**

```
--- 错误处理 ---
FindUser(1): name="user-1" err=<nil>
FindUser(404): err=not found IsNotFound=true
--- 错误包装 ---
ParseAge("abc"): parse age "abc": strconv.ParseInt: parsing "abc": invalid syntax
  errors.Is(ErrSyntax): true
ParseAge("200"): validation: field=age msg=out of range
  errors.As ValidationError: true field=age
--- 接口 ---
TotalArea: 24.57
--- 类型断言 ---
assert Circle: R=1.0
int=42
string="go"
Shape area=12.57
--- nil 接口 ---
p==nil: true  i==nil: false  i type=*main.Person
```

### 3.2 跟着写：分层包装错误

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read %s: %w", path, err)
    }
    cfg, err := parseConfig(data)
    if err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return cfg, nil
}
```

最外层日志可打印完整链；内层用 `errors.Is(err, os.ErrNotExist)` 判断文件缺失。

### 3.3 跟着写：为小类型定义 Stringer

```go
type Status int

const (
    Pending Status = iota
    Done
)

func (s Status) String() string {
    switch s {
    case Pending:
        return "pending"
    case Done:
        return "done"
    default:
        return fmt.Sprintf("Status(%d)", s)
    }
}
```

实现 `fmt.Stringer`（`String() string`）后，`fmt.Println(status)` 会输出可读字符串——标准库通过**接口**扩展行为，无需改 `fmt` 源码。

### 3.4 自检清单

- [ ] 能解释 `error` 接口与 `(T, error)` 惯用法
- [ ] 能区分 `%w` 与 `%v`，并用 `errors.Is` / `errors.As` 处理包装链
- [ ] 能定义小接口并用 struct 隐式实现
- [ ] 能解释值/指针接收者对接口赋值的影响
- [ ] 能写出 type switch，并解释 typed nil 为何 `i == nil` 为 false
- [ ] `go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 忽略 error 或只打日志不 return

```go
data, _ := os.ReadFile(path) // 危险：后续可能用 nil data
```

**修复：** 至少 `if err != nil { return err }`；若确实可忽略，注释说明原因。

### 4.2 用 == 比较包装后的 sentinel

```go
if err == ErrNotFound { ... } // 中间层 %w 包装后常为 false
```

**修复：** `errors.Is(err, ErrNotFound)`。

### 4.3 fmt.Errorf 用 %v 却期望 Is/As 生效

```go
return fmt.Errorf("wrap: %v", err) // 链断掉
```

**修复：** 需要保留链时用 `%w`。

### 4.4 值类型赋给要指针接收者的接口

```go
type Counter interface { Inc() }
type C struct{ n int }
func (c *C) Inc() { c.n++ }

var x Counter = C{} // 编译错误
```

**修复：** `var x Counter = &C{}`，或把方法改到值接收者（若语义允许）。

### 4.5 返回 typed nil 的 error

```go
func Get() error {
    var p *MyErr // p == nil
    return p     // 动态类型是 *MyErr，动态值是 nil → err != nil ！
}
```

调用方 `if err != nil` 为 **true**，尽管 `p` 本身是 nil 指针——与 2.10 节同一原理。**修复：** 切勿 return typed nil；成功路径 `return nil`，失败路径 `return fmt.Errorf(...)` 或 `return &MyErr{...}`。

### 4.6 在 type switch 里对 nil 接口 panic

对 `any` 做断言前，若可能为「无类型的 nil 接口」，直接 `v.(Concrete)` 会 panic。用 `v, ok := i.(T)` 或 `switch` 的 default 分支兜底。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. `error` 是带 `Error() string` 的接口；业务失败用返回值，非常少用 panic
2. Sentinel 用 `errors.Is`；自定义错误类型用 `errors.As` 取字段
3. `%w` 保留包装链，`%v` 只拼字符串
4. 接口隐式实现；注意值/指针接收者与方法集
5. `any` + type assertion / type switch 做运行时类型分支
6. typed nil：`var i any = (*T)(nil)` 时 `i != nil`

**官方文档：**

- [A Tour of Go：Errors](https://go.dev/tour/methods/19)
- [A Tour of Go：Interfaces](https://go.dev/tour/methods/9)
- [Go 1.13：Working with Errors](https://go.dev/blog/go1.13-errors)
- [Effective Go：Errors](https://go.dev/doc/effective_go#errors)
- [Effective Go：Interfaces](https://go.dev/doc/effective_go#interfaces)

**与本仓库的关系：**

- 上一篇：[Go 函数、结构体与作用域](go-functions-structs-and-scope.md)
- 示例：[`example/errors-interfaces/`](../example/errors-interfaces/)
- 下一篇：并发基础（goroutine 与 channel，待写）
