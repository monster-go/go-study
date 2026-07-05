# Go 语句控制：if、for、switch 与 range

> 知识点总结：掌握 Go 唯一的循环关键字 `for` 的多种形式、`if` 短声明、`switch` 的表达式与无标签分支，以及 `range` 遍历 slice/string/map；理解 **break 只跳出最内层** 等默认行为。

---

## 1. 为什么需要了解这个

控制流决定程序「走哪条路、循环多少次」。Go 的设计与 C/Java 有差异，新人常困惑：

- 找 `while` 关键字——Go **没有** `while`，用 `for` 代替
- `if` 里写 `if x := f(); x > 0` 的作用域规则不清楚
- `switch` 默认**不会贯穿（fallthrough）**，与 C 不同
- `for range` 里修改元素，搞不清何时改的是副本、何时改原数据
- `break` 在嵌套循环里只跳出一层

本篇建立在 [运算符](go-operators.md) 之后，是写任何「有分支、有循环」逻辑的前提。后续数组切片、map、并发都会反复用到 `for` 与 `range`。

---

## 2. 核心概念

### 2.1 if / else

```go
if score >= 60 {
    fmt.Println("及格")
} else if score >= 90 {
    fmt.Println("优秀")
} else {
    fmt.Println("不及格")
}
```

**带初始化的 if（if statement with init）：**

```go
if err := do(); err != nil {
    return err
}
// err 仅在此 if-else 块内可见
```

条件必须是 `bool` 表达式，不能写 `if n`（Go 没有 C 式「非零即真」）。

### 2.2 for：Go 唯一的循环关键字

Go 没有 `while`、`do-while`，三种常见 `for` 写法：

| 形式 | 语法 | 类似 |
|------|------|------|
| 三段式 | `for init; cond; post { }` | C 的 for |
| 仅条件 | `for cond { }` | while |
| 无限循环 | `for { }` | while(true) |

```go
// 三段式
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// 仅条件
n := 0
for n < 5 {
    n++
}

// 无限循环（用 break 退出）
for {
    if done {
        break
    }
}
```

### 2.3 break、continue 与标签（label）

| 语句 | 作用 |
|------|------|
| `break` | 跳出**最内层** `for`/`switch`/`select` |
| `continue` | 进入下一轮循环 |
| `goto` | 跳转到标签；少用，了解即可 |

跳出**外层**循环需标签：

```go
Outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i*j > 2 {
            break Outer
        }
    }
}
```

### 2.4 switch

**按值匹配：**

```go
switch day {
case 1:
    fmt.Println("周一")
case 2, 3:
    fmt.Println("周二或周三")
default:
    fmt.Println("其他")
}
```

**无表达式 switch（类似 if-else 链）：**

```go
switch {
case score >= 90:
    grade = "A"
case score >= 60:
    grade = "B"
default:
    grade = "C"
}
```

| 与 C 的差异 | Go 行为 |
|-------------|---------|
| fallthrough | **默认不贯穿**；需显式 `fallthrough` 才进入下一 case |
| case 类型 | 各 case 须与 switch 表达式**类型一致** |
| 多值 | `case 1, 2, 3:` 合法 |

**带初始化的 switch：**

```go
switch err := f(); err {
case nil:
    fmt.Println("ok")
default:
    fmt.Println(err)
}
```

类型 switch（`switch x := v.(type)`）在接口篇展开，此处知道存在即可。

### 2.5 range 遍历

`range` 可用于 slice、array、string、map、channel：

| 遍历对象 | 第一个值 | 第二个值 |
|----------|----------|----------|
| slice / array | 索引 `i` | 元素副本 `v` |
| string | 字节偏移 `i` | rune `r` |
| map | key | value（顺序随机） |
| channel | 接收到的值 | |

```go
nums := []int{10, 20, 30}
for i, v := range nums {
    fmt.Println(i, v)
}

// 只要索引
for i := range nums { ... }

// 只要值（用 _ 忽略索引）
for _, v := range nums { ... }
```

**修改 slice 元素：** `range` 的 `v` 是**副本**；要改原 slice 用索引：

```go
for i := range nums {
    nums[i] *= 2  // OK
}
for _, v := range nums {
    v *= 2        // 无效，只改了副本
}
```

**string 的 range：** 第二个值是 rune，不是 byte。

---

## 3. 动手实践

示例代码在 [`example/controlflow/`](../example/controlflow/)。

### 3.1 运行示例

```bash
cd example/controlflow
go run .                  # 全部演示
go run . -mode=if         # if / else 与短声明
go run . -mode=for        # 三种 for
go run . -mode=switch     # 值 switch 与无表达式 switch
go run . -mode=range      # slice 与 string
```

预期（节选）：

```
--- if / else ---
7 是奇数
v = 14 > 10
--- for 循环 ---
1..5 求和: 15
--- switch ---
周二或周三
工作日
```

### 3.2 跟着改：打印 99 乘法表（单循环版）

```go
for i := 1; i <= 9; i++ {
    for j := 1; j <= i; j++ {
        fmt.Printf("%d×%d=%d\t", j, i, i*j)
    }
    fmt.Println()
}
```

完整带边框版本见 [`example/day01/`](../example/day01/)。

### 3.3 跟着改：用 switch 替代 if-else 链

```go
switch month {
case 12, 1, 2:
    season = "冬"
case 3, 4, 5:
    season = "春"
default:
    season = "其他"
}
```

### 3.4 自检清单

- [ ] 能写出 `for` 的三种形式
- [ ] 知道 Go 没有 `while`，用 `for cond` 代替
- [ ] 知道 `switch` 默认不 fallthrough
- [ ] 知道 `range` 修改 slice 要用索引
- [ ] `go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 if 条件非 bool

```go
if n { } // 编译错误：n 是 int，不是 bool
```

**修复：** `if n != 0 { }` 或 `if n > 0 { }`。

### 4.2 range 修改 slice 无效

```go
for _, v := range nums {
    v = 0 // 只改副本
}
```

**修复：** `nums[i] = 0` 配合 `for i := range nums`。

### 4.3 switch 忘记 break 导致「贯穿」——在 Go 里通常不是问题

C 程序员习惯性在 case 末尾写 `break`；Go **默认已 break**，除非写 `fallthrough`。

### 4.4 闭包捕获循环变量（Go 1.22+ 前常见）

在 `for` 里启动 goroutine 若捕获循环变量 `i`，可能全部看到最终值。Go 1.22+ 每轮迭代有独立变量；旧版本需在循环内 `i := i` 复制。

**验证：** 写小测试打印 goroutine 中的 `i`，对照你的 Go 版本文档。

### 4.5 map 的 range 顺序随机

```go
for k, v := range m {
    // 每次运行顺序可能不同
}
```

**修复：** 需要有序输出时，先收集 key 再 `sort.Strings(keys)` 后遍历。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. Go 只有 `for`，无 `while`；`for cond { }` 即 while
2. `if` 可带短声明，变量作用域限于该 if-else 块
3. `switch` 默认不贯穿；无表达式形式可替代 if-else 链
4. `break`/`continue` 默认只影响最内层循环
5. `range` 遍历 slice 时，改元素要用索引；`v` 是副本
6. `range` 遍历 string 得到的是 rune

**官方文档：**

- [A Tour of Go：Flow control](https://go.dev/tour/flowcontrol/1)
- [Go 语言规范：For statements](https://go.dev/ref/spec#For_statements)

**与本仓库的关系：**

- 上一篇：[Go 字符串操作](go-string-operations.md)
- 示例：[`example/controlflow/`](../example/controlflow/)、[`example/day01/`](../example/day01/)
- 下一篇：[Go 数组与切片](go-arrays-and-slices.md)
