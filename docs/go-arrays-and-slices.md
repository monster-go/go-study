# Go 数组、切片与 map：定长数组、动态切片与键值映射

> 知识点总结：区分**数组（array）**与**切片（slice）**；掌握 `len`/`cap`、`make`/`append`、切片表达式与底层数组共享；理解 **map（映射）** 的创建、读写、删除与 `range` 遍历；能避免切片越界、共享底层意外，以及向 nil map 写入等常见错误。

---

## 1. 为什么需要了解这个

切片是 Go 最常用的集合类型之一，但语义比 Java `ArrayList` 或 Python `list` 更底层。map 则类似 Python `dict` 或 Java `HashMap`，但同样有一些 Go 特有的规则。新人常困惑：

- 以为 `[]int` 是数组——实际是**切片类型**
- `append` 后原切片「变了没」、何时会触发扩容
- `s[1:3]` 与 `len`、`cap` 的关系算不清
- 两个切片共享底层数组，改一处另一处也变
- `nil` 切片与空切片 `[]int{}` 能否 `append`、JSON 表现是否相同
- 向 **nil map** 写入直接 panic，与 nil 切片行为不同
- 读不存在的 map key 得到零值，误以为 key 一定存在

本篇建立在 [语句控制](go-control-flow.md) 之后（会 `for range` 遍历）。slice 与 map 是 Go 最常用的两种集合类型；后续函数参数、JSON 编码都会大量涉及它们。

---

## 2. 核心概念

### 2.1 数组 vs 切片

| | 数组（array） | 切片（slice） |
|---|---------------|---------------|
| 声明 | `[N]T` 或 `var a [3]int` | `[]T` 或 `var s []int` |
| 长度 | **编译期固定**，是类型的一部分 | **运行时可变**（通过 append） |
| 传参 | 按值复制整个数组 | 传递切片头（指针+len+cap），共享底层 |
| 常用度 | 较少，特定场景（缓冲、矩阵） | **日常首选** |

```go
var arr [3]int           // 数组，零值 [0 0 0]
arr2 := [3]int{1, 2, 3}

var sl []int             // nil 切片，len=0 cap=0
sl2 := []int{1, 2, 3}    // 切片字面量
```

`[3]int` 与 `[4]int` 是**不同类型**；`[]int` 长度不在类型里。

### 2.2 切片的内部结构（心智模型）

切片是对**底层数组**的引用，包含三个字段：

```
┌─────────────┐
│ 指针 → 底层数组 │
│ len  逻辑长度  │
│ cap  底层容量  │
└─────────────┘
```

- **len**：当前可见元素个数
- **cap**：从切片起始位置到底层数组末尾的容量（可 append 而不搬迁的上限）

```go
s := []int{1, 2, 3, 4, 5}
// 底层数组: index  0   1   2   3   4
//                 [1] [2] [3] [4] [5]
// s: ptr→index 0, len=5, cap=5

sub := s[1:4]  // 从 index 1 到 4（不含 4）→ 元素 [2 3 4]
```

**`s[1:4]` 的 len / cap 怎么算：**

切片表达式 `s[low:high]` 取半开区间 `[low, high)`：

```
low=1, high=4
len = high - low = 4 - 1 = 3        → 可见元素 [2, 3, 4]
cap = cap(s) - low = 5 - 1 = 4      → 从 sub 起点到数组末尾还能 append 4 个位置
```

```
底层:  [1] [2] [3] [4] [5]
              ↑ sub 起点 (index 1)
              └─ sub 可见 3 个 ─┘└ cap 延伸到末尾，共 4 格 ─┘
sub:       [2] [3] [4]   （len=3, cap=4）
```

### 2.3 切片表达式

对 `s[low:high]`：

| 形式 | 含义 | len | cap |
|------|------|-----|-----|
| `s[low:high]` | 下标 `[low, high)` | `high - low` | `cap(s) - low` |
| `s[:high]` | 从开头到 high | `high` | `cap(s)` |
| `s[low:]` | 从 low 到末尾 | `len(s) - low` | `cap(s) - low` |
| `s[:]` | 整段 | `len(s)` | `cap(s)` |
| `s[low:high:max]` | **三索引**，限制 cap | `high-low` | `max-low` |

三索引用于**阻止 append 覆盖原切片后续元素**：

```go
sub := s[1:3:3] // cap(sub) = max - low = 3 - 1 = 2
```

**三索引 `s[low:high:max]` 的 cap 推导：**

```
s = [1 2 3 4 5]   len=5 cap=5
sub = s[1:3:3]    low=1, high=3, max=3

len = high - low = 3 - 1 = 2     → 元素 [2, 3]
cap = max - low = 3 - 1 = 2        → append 最多再写 0 格（已满）
```

cap 被限制为 2 后，对 `sub` 做 `append` 会**分配新数组**，不会覆盖 `s` 里 index 3、4 的元素。

### 2.4 make 与字面量

```go
s1 := make([]int, 5)      // len=5 cap=5，元素为零值
s2 := make([]int, 0, 10)  // len=0 cap=10，预分配容量
s3 := []int{}             // 空切片，非 nil
var s4 []int              // nil 切片
```

| | nil 切片 | 空切片 `[]T{}` 或 `make([]T,0)` |
|---|----------|----------------------------------|
| `len` / `cap` | 0 / 0 | 0 / 0（make 可指定 cap） |
| `append` | 可用 | 可用 |
| `== nil` | true | false |

多数场景两者可互换；需要区分「未初始化」语义时用 nil。

### 2.5 append 与扩容

```go
s := []int{1, 2}
s = append(s, 3, 4)  // 必须接收返回值
```

| 要点 | 说明 |
|------|------|
| 返回值 | `append` 可能返回新底层数组，**必须** `s = append(s, x)` |
| 容量足够 | 在原数组后写入，len 增加，**同底层** |
| 容量不足 | 分配更大数组（通常约 2 倍扩容），**复制**元素，原切片不变 |

**扩容为何 cap 从 2 变成 4（可复算）：**

```go
s := make([]int, 0, 2)   // len=0, cap=2，底层数组预留 2 格
s = append(s, 1, 2)      // len=2, cap=2，刚好填满
s = append(s, 3)           // 需要第 3 格，cap 不够 → 分配新数组
```

Go 运行时在 cap 不足时通常按**约 2 倍**分配新底层数组（小切片常见 2→4→8…），把旧元素复制过去，再写入新元素：

```
append 前:  cap=2  [1][2]           ← 已满
append 后:  cap=4  [1][2][3][ ]     ← 新数组，多出的空位供后续 append
```

因此 `append` **必须**写 `s = append(s, x)`——扩容后指针可能指向新数组，返回值才是最新切片头。

**共享底层的示例：**

```go
a := []int{1, 2, 3}
b := append(a, 4)
a[0] = 99
// 若未扩容，b 与 a 共享底层，b[0] 也是 99
```

### 2.6 copy

```go
dst := make([]int, len(src))
n := copy(dst, src) // n 为实际复制个数 min(len(dst), len(src))
```

`copy` 不依赖 `append`，适合显式拷贝避免共享。

### 2.7 数组作为值类型

```go
func modify(arr [3]int) {
    arr[0] = 100 // 只改副本
}
```

大数组传参会复制；需要修改或避免复制时传**切片**或指针。

### 2.8 map 是什么

**map（映射）** 是键值对的无序集合，底层是哈希表（hash table）。按 key 查找平均 O(1)，适合「按标识查数据」：

| | 切片 `[]T` | map `map[K]V` |
|---|------------|---------------|
| 访问方式 | 整数下标 `s[i]` | 键 `m[key]` |
| 顺序 | 有顺序（下标 0, 1, 2…） | **遍历顺序随机** |
| 零值 | `nil` 切片可读 len、可 append | `nil` map **不能写入** |
| 比较 | 不能 `==` 比内容（仅 `== nil`） | 同上，见 [运算符文档](go-operators.md) |

```go
var m map[string]int        // nil map
m2 := map[string]int{}      // 空 map，已初始化，可写入
m3 := make(map[string]int)  // 空 map，常用写法
scores := map[string]int{"alice": 95, "bob": 88}
```

### 2.9 map 的创建与读写

**三种常见创建方式：**

```go
// 1. 字面量（已知初始键值）
ages := map[string]int{"alice": 30, "bob": 25}

// 2. make（键值后续动态添加）
counts := make(map[string]int)

// 3. var 声明 → nil map，只能读不能写
var cache map[string]string
```

**读写：**

```go
ages["alice"] = 31           // 写入或更新
v := ages["alice"]           // 读取；key 不存在时 v 为 value 类型的零值
v, ok := ages["carol"]       // comma-ok：ok=false 表示 key 不存在
delete(ages, "bob")          // 删除；key 不存在时不 panic
n := len(ages)               // 当前键值对个数
```

**comma-ok 为何要区分「不存在」与「值为零」：**

```go
m := map[string]int{"zero": 0}
v := m["zero"]      // v=0
v2 := m["missing"]  // v2 也是 0——无法区分「没有 key」还是「值就是 0」

v3, ok := m["zero"]    // v3=0, ok=true
v4, ok2 := m["missing"] // v4=0, ok2=false
```

### 2.10 nil map 与空 map

| | `var m map[K]V` | `map[K]V{}` 或 `make(map[K]V)` |
|---|-----------------|----------------------------------|
| 值 | `nil` | 非 nil，指向空哈希表 |
| `len(m)` | 0 | 0 |
| 读取 `m[k]` | 返回零值，不 panic | 返回零值，不 panic |
| 写入 `m[k]=v` | **panic** | 正常 |
| `m == nil` | true | false |

```go
var m map[string]int
// m["x"] = 1  // panic: assignment to entry in nil map

m = make(map[string]int) // 或 m = map[string]int{}
m["x"] = 1               // OK
```

**原理：** nil map 没有分配哈希表存储；写入必须先 `make` 或字面量初始化。读取不触发存储分配，所以安全。

### 2.11 map 的键类型与引用语义

**键必须是可比较（comparable）类型：** 布尔、数值、字符串、指针、数组（元素可比较时）、结构体（字段均可比较时）、interface（动态值可比较时）。

**不能作键：** slice、map、function——因为它们不能用 `==` 比较。

```go
// m := map[[]int]string{} // 编译错误：invalid map key type []int
m := map[[2]int]string{}   // 数组作键 OK（长度是类型一部分）
```

map 是**引用类型**：传参时不复制整张表，函数内修改会影响外部。与 slice 一样，**不能用 `==` 比较内容**（只能 `== nil`）；比较内容用 Go 1.21+ 的 `maps.Equal`（见 [运算符文档](go-operators.md)）。

**不能对 map 元素取地址：**

```go
m := map[string]int{"a": 1}
// p := &m["a"] // 编译错误：cannot take address of m["a"]
```

原因：map 扩容时会 rehash，元素地址可能变化。

### 2.12 遍历 map 与并发注意

```go
for k, v := range m {
    fmt.Println(k, v)
}

for k := range m { /* 只要 key */ }
for _, v := range m { /* 只要 value */ }
```

**遍历顺序随机**——不要依赖 `range map` 的输出顺序；需要有序输出时，先把 key 收集到 slice 再排序。

map **不是并发安全**的：多个 goroutine 同时读写同一 map 会 panic。并发场景用 `sync.Mutex` 保护，或 `sync.Map`（特定模式）。本篇先掌握单 goroutine 用法。

---

## 3. 动手实践

示例代码在 [`example/slices/`](../example/slices/)。

### 3.1 运行示例

```bash
cd example/slices
go run .                    # 全部演示
go run . -mode=array        # 定长数组
go run . -mode=slice        # len/cap 与切片表达式
go run . -mode=append       # append 与 copy
go run . -mode=share        # 底层数组共享
```

预期（节选）：

```
--- 切片 ---
s: [1 2 3 4 5] len: 5 cap: 5
s[1:4]: [2 3 4] len: 3 cap: 4
--- 底层数组共享 ---
改 alias 后 orig: [1 99 3 4]
```

### 3.2 跟着改：观察 append 扩容

```go
s := make([]int, 0, 2)
fmt.Println(cap(s))
s = append(s, 1, 2)
fmt.Println(cap(s)) // 2
s = append(s, 3)
fmt.Println(cap(s)) // 通常变为 4
```

### 3.3 跟着改：三索引避免覆盖

```go
data := []int{0, 1, 2, 3, 4}
head := data[:2:2] // cap=2
head = append(head, 99)
fmt.Println(data) // [0 1] 未被 99 覆盖（append 新数组）
```

### 3.4 自检清单（数组与切片）

- [ ] 能区分 `[3]int` 与 `[]int`
- [ ] 能解释 `s[1:4]` 的 len 和 cap
- [ ] 知道 `append` 必须接收返回值
- [ ] 知道子切片与原切片可能共享底层数组
- [ ] `go run .` 各 `-mode` 无报错

### 3.5 map 动手实践

示例代码在 [`example/maps/`](../example/maps/)。

#### 3.5.1 运行示例

```bash
cd example/maps
go run .                    # 全部演示
go run . -mode=basic        # 字面量与 make
go run . -mode=ok           # comma-ok 判断 key 是否存在
go run . -mode=nil          # nil map 与空 map
go run . -mode=range        # range 遍历
go run . -mode=delete       # delete 与 len
```

预期（节选）：

```
--- map 基本操作 ---
字面量: map[a:1 b:2]
make + 赋值: map[x:10]
--- comma-ok ---
存在 go: 1 true
不存在 java: 0 false
--- nil map ---
nil map len: 0 == nil: true
空 map {} len: 0 == nil: false
```

#### 3.5.2 跟着改：统计词频

```go
text := []string{"go", "map", "go", "slice"}
freq := make(map[string]int)
for _, word := range text {
    freq[word]++
}
fmt.Println(freq) // map[go:2 map:1 slice:1]
```

`freq[word]++` 等价于：若 key 不存在，读出零值 0，加 1 后写回——这正是 map 计数器的惯用法。

#### 3.5.3 跟着改：安全读取配置

```go
config := map[string]string{"host": "localhost"}
if port, ok := config["port"]; ok {
    fmt.Println("port:", port)
} else {
    fmt.Println("port 未配置，使用默认 8080")
}
```

#### 3.5.4 自检清单（map）

- [ ] 能写出三种 map 创建方式并说明 nil map 为何不能写入
- [ ] 能用 comma-ok 区分「key 不存在」与「值为零」
- [ ] 知道 `range map` 顺序不固定
- [ ] `cd example/maps && go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 切片越界

```go
s := []int{1, 2, 3}
s[3] = 4 // panic: index out of range
```

**修复：** 访问前检查 `i < len(s)`；扩展用 `append`。

### 4.2 append 未赋值

```go
append(s, 1) // 编译可能通过但结果丢弃；s 未变
```

**修复：** `s = append(s, 1)`。

### 4.3 共享底层导致「莫名被改」

```go
sub := orig[0:2]
sub[0] = 999 // orig[0] 也变 999
```

**修复：** 需要独立副本时 `copy` 或 `append([]int(nil), sub...)`。

### 4.4 用 len 当「还有空间 append」

`cap - len` 才是剩余容量；对子切片 append 可能覆盖**原切片**后面的元素（共享底层且 cap 延伸到后面时）。

**修复：** 三索引 `s[low:high:high]` 限制 cap，或 `append` 前先 `copy` 到新切片。

### 4.5 二维切片未正确初始化

```go
matrix := make([][]int, 3) // 3 个 nil 内层切片
matrix[0][0] = 1           // panic
```

**修复：** 循环 `matrix[i] = make([]int, cols)`。

### 4.6 向 nil map 写入

```go
var m map[string]int
m["k"] = 1 // panic: assignment to entry in nil map
```

**修复：** 先 `m = make(map[string]int)` 或 `m = map[string]int{}`。

### 4.7 把「读到零值」当成「key 存在」

```go
m := map[string]int{"count": 0}
if m["count"] != 0 { /* ... */ } // 永远不进分支，但 key 其实存在
```

**修复：** 用 `v, ok := m["count"]` 判断存在性；或单独约定 sentinel 值。

### 4.8 依赖 map 遍历顺序

```go
for k, v := range m {
    fmt.Println(k, v) // 两次运行顺序可能不同
}
```

**修复：** 需要稳定顺序时，收集 key 到 slice，`sort.Strings(keys)` 后再访问。

### 4.9 并发读写同一 map

多个 goroutine 同时写（或一读一写）会触发 `fatal error: concurrent map writes`。

**修复：** 加锁、用 channel 串行化，或评估 `sync.Map`；不要在未同步时共享 map。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. 数组 `[N]T` 长度固定；切片 `[]T` 是日常使用的动态视图
2. 切片 = 指针 + len + cap；对底层数组的引用
3. `s[low:high]` 半开区间；cap 通常延伸到原数组末尾
4. `append` 可能扩容并返回新切片，务必 `s = append(s, x)`
5. 子切片与原切片共享底层时，修改会相互影响
6. `copy` 用于显式复制；三索引用于限制 append 范围
7. map `map[K]V` 是键值哈希表；键须可比较，遍历顺序随机
8. nil map 可读不可写；空 map `{}` / `make` 后可正常增删
9. 用 comma-ok 判断 key 是否存在；比较 map 内容用 `maps.Equal`

**官方文档：**

- [A Tour of Go：Slices](https://go.dev/tour/moretypes/7)
- [A Tour of Go：Maps](https://go.dev/tour/moretypes/8)
- [Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [pkg.go.dev/builtin#append](https://pkg.go.dev/builtin#append)
- [pkg.go.dev/maps](https://pkg.go.dev/maps)（Go 1.21+，`maps.Equal` 等）

**与本仓库的关系：**

- 上一篇：[Go 语句控制](go-control-flow.md)
- 示例：[`example/slices/`](../example/slices/)、[`example/maps/`](../example/maps/)
- 相关：[运算符文档](go-operators.md)（slice/map 比较规则）、channel（待写）
