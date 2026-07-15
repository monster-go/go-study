# Go 断言：从接口类型断言到 testify 测试断言

> 知识点总结：掌握 Go 中「断言」的两种形态——运行时从接口取回具体类型的**类型断言**，与单元测试中校验预期的**测试断言**；理解类型断言的两种写法（安全与非安全）、type switch 用法，以及 testify/assert 与 testify/require 的取舍。

---

## 1. 为什么需要了解这个

断言（Assertion）在 Go 里体现为**两个完全不同的概念**，新人容易混：

- **类型断言**：接口变量 `any` 存了具体值，想取回来用——`v, ok := x.(T)`
- **测试断言**：写单元测试时验证结果是否等于预期——`assert.Equal(t, want, got)`

新手最常见的问题：

- 从 `interface{}` 取具体值时，「带 ok」的形式和「不带 ok」的有什么区别
- 以为 `switch v := x.(type)` 是某种语法糖，不知道它能省掉大量 `if/else`
- 写单元测试时只会手写 `if got != want { t.Errorf(...) }`，不知道 testify 一行解决
- 分不清 testify 的 `assert` 和 `require`——前者失败后继续跑，后者立即中止

类型断言在 [接口与错误处理](go-interfaces-and-error-handling.md) 中已有介绍，本篇会系统整理并补充测试断言这一独立知识块。

---

## 2. 核心概念

### 2.1 什么是断言（Assertion）

断言的本质是**在某个时间点做一项检查，若检查失败则发出信号**。Go 中两种断言的差异：

| 维度 | 类型断言 | 测试断言 |
|------|---------|---------|
| **时机** | 运行期，从接口取具体值时 | 测试期间，验证计算结果时 |
| **失败信号** | panic（不安全形式）/ `ok=false`（安全形式） | 测试失败报告（`t.Error` / `t.Fatal`） |
| **核心语法** | `x.(T)` / `switch x.(type)` | `assert.Equal(t, ...)` / `require.Equal(t, ...)` |
| **常用场景** | JSON 反序列化、反射、泛型前多态 | 单元测试、集成测试 |
| **内置于标准库** | ✅ 是 | ❌ 需第三方包 `testify` |

### 2.2 类型断言（Type Assertion）

从接口值中提取具体类型值：

```go
var v any = "hello"

// 不安全形式：失败时 panic
s := v.(string)           // ok: "hello"
n := v.(int)              // panic: interface conversion

// 安全形式：失败时 ok=false，不 panic
s, ok := v.(string)       // ok=true, s="hello"
n, ok := v.(int)          // ok=false, n=0
```

**原理：** 接口值底层是 `(type, value)` 二元组。类型断言检查**运行时类型**是否与目标匹配，不比较内容。

#### type switch — 多分支类型匹配

```go
func typeOf(v any) string {
    switch x := v.(type) {
    case nil:
        return "nil"
    case int, int8, int16, int32, int64:
        return "integer"
    case string:
        return fmt.Sprintf("string(%q)", x) // x 是 string
    case bool:
        return fmt.Sprintf("bool(%t)", x)   // x 是 bool
    default:
        return fmt.Sprintf("unknown(%T)", x)
    }
}
```

> `x` 在每个 `case` 内已被转型为对应类型，可直接调用该类型的方法。

### 2.3 测试断言（Test Assertion）

Go 标准库 `testing` 不内置断言函数，需使用测试框架。最流行的是 [testify](https://github.com/stretchr/testify)。

```go
import "github.com/stretchr/testify/assert"

func TestAdd(t *testing.T) {
    result := Add(1, 2)
    assert.Equal(t, 3, result)
    assert.NotNil(t, result)
    assert.True(t, result > 0)
}
```

#### assert vs require

| 包 | 失败后行为 | 底层调用 |
|------|-----------|---------|
| `assert` | 继续执行后续断言 | `t.Error` |
| `require` | 立即终止当前测试 | `t.Fatal` |

**选型原则：** 前置条件（如数据初始化成功）用 `require`；结果验证用 `assert`，以便一次运行看到尽可能多的失败信息。

---

## 3. 应用场景

### 场景 1：JSON 反序列化后的类型恢复

`json.Unmarshal` 返回 `any`（通常是 `map[string]any`），取值时必须断言：

```go
var data any
json.Unmarshal([]byte(`{"name":"go","ver":1.21}`), &data)

m := data.(map[string]any)
name := m["name"].(string)   // "go"
ver  := m["ver"].(float64)   // JSON 数字默认是 float64
```

### 场景 2：自定义错误类型的解包

`errors.As` 内部就是做了类型断言 + 沿包装链递归：

```go
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println("字段:", ve.Field) // 取自定义错误字段
}
```

可参考 [接口与错误处理](go-interfaces-and-error-handling.md) 第 2.4 节。

### 场景 3：单元测试中校验函数行为

```go
func TestParseAge(t *testing.T) {
    age, err := parseAge("25")
    require.NoError(t, err)     // 前置：不能有错
    assert.Equal(t, 25, age)   // 验证结果

    _, err = parseAge("abc")
    assert.Error(t, err)        // 验证错误存在
    assert.Contains(t, err.Error(), "invalid")
}
```

### 场景 4：通用处理器的参数分发

```go
func handleEvent(v any) {
    switch e := v.(type) {
    case *ClickEvent:
        log.Printf("click at (%d,%d)", e.X, e.Y)
    case *ScrollEvent:
        log.Printf("scroll delta=%d", e.Delta)
    default:
        log.Printf("unhandled: %T", e)
    }
}
```

典型于事件驱动、消息队列 handler、AST 节点处理等场景。

---

## 4. 动手实践

### 4.1 运行仓库示例

```bash
cd example/assertions
go run .                   # 类型断言 + type switch
go test -v                 # 测试断言（testify）
```

**预期输出片段：**

```
--- 类型断言 ---
safe: "hello" ok=true
unsafe(nil check): ok=false
--- type switch ---
int=42
string="go"
bool(true)
nil
--- JSON 解包 ---
name=go, version=1.21
```

**测试输出片段：**

```
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestParseAge
--- PASS: TestParseAge (0.00s)
=== RUN   TestAddFailures
    main_test.go:xx: Expected value 99, but got 3
--- PASS: TestAddFailures (0.00s)
PASS
```

### 4.2 跟着写：安全断言处理 JSON

```go
func getString(m map[string]any, key string) string {
    v, ok := m[key]
    if !ok {
        return ""
    }
    s, ok := v.(string)
    if !ok {
        return ""
    }
    return s
}

func main() {
    data := map[string]any{
        "name": "Go",
        "ver":  1.21,
    }
    fmt.Println(getString(data, "name")) // "Go"
    fmt.Println(getString(data, "ver"))  // "" (因为 ver 是 float64)
}
```

### 4.3 跟着写：用 testify 写单元测试

```go
// main.go
func Add(a, b int) int { return a + b }

// main_test.go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 4.4 自检清单

- [ ] 能说出「安全断言」和「不安全断言」的区别
- [ ] 能写出 `switch x := v.(type)` 并说出用途
- [ ] 能安装 testify 并写出测试文件
- [ ] 能区分 `assert.Equal` 和 `require.Equal` 的使用时机
- [ ] `cd example/assertions && go run . && go test -v` 无报错

---

## 5. 常见坑与排查

### 5.1 对非接口类型做类型断言

```go
var s string = "hello"
s.(string)          // 编译错误：s 不是接口类型
```

**原因：** 类型断言只对**接口类型**有效（`any`、`error`、自定义接口等）。非接口类型直接使用即可，无需断言。

### 5.2 JSON 数字默认是 float64

```go
data := map[string]any{"count": 3}
c := data["count"].(int) // panic: float64 != int
```

**修复：** JSON 规范没有整型——`json.Unmarshal` 默认用 `float64`。解决方法：

```go
c := int(data["count"].(float64))  // 先断言 float64 再转型
// 或用 json.NewDecoder + UseNumber()
```

### 5.3 未导入 testify 直接跑测试

```bash
go test ./...
# 报错：undefined: assert
```

**修复：** 确保 `go.mod` 中有 `require github.com/stretchr/testify`，且文件中 `import` 了。

### 5.4 测试断言误传参数顺序

```go
assert.Equal(t, got, want) // 顺序错了！
```

`Equal(t, expected, actual)` 才是约定顺序。用反后，出错时的报告会误导：

```
Expected: <实际值>  ← 误导
Actual:   <期望值>  ← 误导
```

**修复：** 始终 `assert.Equal(t, want, got)`——**期望值在前，实际值在后**。

### 5.5 在错误的 goroutine 中调用 t.Fatal

```go
go func() {
    result := doSomething()
    require.Equal(t, 42, result) // 崩溃或死锁
}()
```

**原因：** `t.Fatal` / `require` 必须在**测试 goroutine** 中调用。**修复：** 使用 `t.Run` 或 channels 把结果带回主 goroutine。

---

## 6. 小结与延伸阅读

**要点回顾：**

1. 类型断言从接口取具体类型：安全形式 `v, ok := x.(T)`，永远不要用不安全形式
2. `switch x := v.(type)` 在多个类型分支时比多个 `if` 断言更清晰
3. 类型断言只对**接口类型**有效；普通类型直接用，无需断言
4. testify 是 Go 最流行的测试断言库，`assert` 失败后继续，`require` 立即终止
5. 测试断言始终 `assert.Equal(t, expected, actual)`——期望在前，实际在后

**官方文档：**

- [A Tour of Go：Type assertions](https://go.dev/tour/methods/15)
- [A Tour of Go：Type switches](https://go.dev/tour/methods/16)
- [Effective Go：Interface conversions and type assertions](https://go.dev/doc/effective_go#interface_conversions)
- [testify GitHub](https://github.com/stretchr/testify)
- [Go testing 官方文档](https://pkg.go.dev/testing)

**与本仓库的关系：**

- 关联：[Go 接口与错误处理](go-interfaces-and-error-handling.md)（第 2.7 节类型断言的基础内容）
- 示例：[`example/assertions/`](../example/assertions/)
- 下一篇：并发基础（goroutine 与 channel，待写）
