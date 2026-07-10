# Go 集合与类型操作 API 演进：从 sort 到 slices / maps / cmp

> 知识点总结：理解 Go 标准库中「同一操作多种写法」背后的演进脉络；掌握 Go 1.21+ 的 `slices`、`maps`、`cmp` 与内置 `clear` 的用法；能在读老代码与写新代码之间做出合理选型。

---

## 1. 为什么需要了解这个

Go 不像某些语言那样频繁「删旧 API」，但会**持续追加**更现代的标准库。新人常困惑：

- 教程写 `sort.Strings`，搜到的答案写 `slices.Sort`，到底用哪个？
- 删 slice 元素有 `append` 拼接、`slices.Delete`、swap 截断——哪种才对？
- `reflect.DeepEqual` 和 `slices.Equal` 有什么区别？
- 字符串操作用 `strings` 还是 `bytes`？数字转字符串用 `strconv` 还是 `fmt`？

**本质原因：** Go 1.0 起就有 `sort`、`reflect` 等包；Go 1.18 引入**泛型**；Go 1.21 新增 **`slices` / `maps` / `cmp`** 三个泛型包，把 slice、map 上的常见操作统一成类型安全的 API。旧写法**仍然有效**，保证百万行存量代码可编译。

本篇建立在 [数组、切片与 map](go-arrays-and-slices.md) 之后。你已理解 slice/map 语义，这里聚焦「**怎么操作**」以及「**为什么有两套写法**」。

---

## 2. 核心概念

### 2.1 演进时间线（心智模型）

```
Go 1.0    内置 + sort / reflect / strings / strconv
   ↓
Go 1.8    sort.Slice / sort.SliceStable（闭包比较）
   ↓
Go 1.13   errors.Is / errors.As（错误链）
   ↓
Go 1.18   泛型（语言层）
   ↓
Go 1.20   math/rand 自动 seed；crypto/rand 等调整
   ↓
Go 1.21   slices / maps / cmp 包；内置 clear() 支持 slice 与 map
   ↓
Go 1.22   for range 循环变量语义修复；整数 range 等
```

**选型原则（2026 年新代码）：**

| 层级 | 何时用 |
|------|--------|
| **内置** `append`、`delete`、`copy`、`len`、`clear` | 语言级、最高频 |
| **slices / maps / cmp**（Go 1.21+） | slice/map 上的排序、查找、克隆、比较 |
| **sort / reflect** | 维护老代码；`sort.Interface` 遗留类型 |
| **手写循环** | 逻辑特殊、性能极致优化、或目标 Go 版本 < 1.21 |

---

### 2.2 排序（Sort）

| 写法 | 引入版本 | 适用 | 缺点 |
|------|----------|------|------|
| `sort.Strings(s)` / `sort.Ints(s)` | Go 1.0 | `[]string` / `[]int` 等少数类型 | 类型不通用 |
| `sort.Sort(sort.StringSlice(s))` | Go 1.0 | 实现 `sort.Interface` 的自定义类型 | 样板代码多 |
| `sort.Slice(s, less)` | Go 1.8 | 任意切片，闭包比较 | `less` 参数是 `any`，无类型安全 |
| `slices.Sort(s)` | Go 1.21 | 元素可 `<` 的 `[]E` | 需 Go 1.21+ |
| `slices.SortFunc(s, cmp)` | Go 1.21 | 自定义比较，返回 -1/0/1 | 需配合 `cmp` 包 |
| `slices.SortStable` / `SortStableFunc` | Go 1.21 | 稳定排序（相等元素保持原序） | 略慢于不稳定排序 |

```go
import (
    "cmp"
    "slices"
    "sort"
)

// ① 旧：类型专用
s := []string{"c", "a", "b"}
sort.Strings(s)

// ② 旧：闭包
type Person struct{ Name string; Age int }
people := []Person{{"bob", 30}, {"alice", 25}}
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})

// ③ 新：泛型 + cmp（推荐）
slices.Sort(s)
slices.SortFunc(people, func(a, b Person) int {
    return cmp.Compare(a.Age, b.Age)
})
```

**map 有序遍历：** map 本身无序；需稳定输出时，收集 key → 排序 → 再访问（见 [example/maps](../example/maps/) 的 `demoOrder`）。

---

### 2.3 删除（Delete）

#### map：内置 `delete`，写法唯一

```go
delete(m, "key") // key 不存在也不 panic
```

#### slice：多种写法，语义不同

| 写法 | 保持顺序 | 复杂度 | 说明 |
|------|----------|--------|------|
| `s = append(s[:i], s[i+1:]...)` | 是 | O(n) | 经典写法，Go 1.0 即有 |
| `s = slices.Delete(s, i, j)` | 是 | O(n) | Go 1.21+，语义清晰 |
| `s = slices.DeleteFunc(s, pred)` | 是 | O(n) | 按条件批量删除 |
| swap + `s = s[:len(s)-1]` | **否** | O(1) | 性能优先、顺序无所谓 |

```go
// 删下标 2（保持顺序）
s = slices.Delete(s, 2, 3)

// 删所有偶数
s = slices.DeleteFunc(s, func(v int) bool { return v%2 == 0 })

// O(1) 删下标 i（不保持顺序）
i := 2
s[i] = s[len(s)-1]
s = s[:len(s)-1]
```

#### map 批量删除

```go
// 旧：range + delete
for k, v := range m {
    if v < 0 {
        delete(m, k)
    }
}

// 新：Go 1.21+
maps.DeleteFunc(m, func(k string, v int) bool { return v < 0 })
```

#### string：不可变，「删除」= 构造新串

字符串不能原地删字符：

```go
s := "hello"
s = s[:2] + s[3:]           // "helo"
s = strings.Replace(s, "l", "", 1) // 删第一个 "l"
```

---

### 2.4 查找与包含（Search / Contains）

| 操作 | 旧写法 | 新写法（Go 1.21+） |
|------|--------|-------------------|
| slice 是否含元素 | 手写 `for` 循环 | `slices.Contains(s, x)` |
| slice 找下标 | 手写循环 | `slices.Index` / `IndexFunc` |
| slice 二分查找（已排序） | `sort.Search` + 断言 | `slices.BinarySearch` / `BinarySearchFunc` |
| map 是否含 key | `_, ok := m[k]` | `maps.Contains(m, k)`（语义相同，可选） |

```go
// slice
if slices.Contains(tags, "go") { /* ... */ }
i := slices.Index(nums, 42) // 找不到返回 -1

// 已排序 slice 上二分
i, found := slices.BinarySearch(sorted, 42)

// map：comma-ok 仍是 idiomatic 写法
if v, ok := m["key"]; ok { /* ... */ }
// 或
if maps.Contains(m, "key") { /* ... */ }
```

**strings / bytes 对称 API：** 处理 `[]byte` 用 `bytes.Contains`，处理 `string` 用 `strings.Contains`——函数签名几乎相同，因为早期 Go 尚未泛型，只能分类型维护。

---

### 2.5 比较相等（Equal / Compare）

| 写法 | 适用 | 注意 |
|------|------|------|
| `==` | 可比较的基本类型、指针、数组（元素可比较） | slice/map **不能** `==` 比内容 |
| `slices.Equal(a, b)` | 两个 `[]E`（元素可 `==`） | Go 1.21+ |
| `maps.Equal(a, b)` | 两个 `map[K]V` | Go 1.21+ |
| `reflect.DeepEqual(a, b)` | 任意类型、嵌套结构 | 慢、无编译期类型检查 |
| `cmp.Compare(a, b)` | 有序类型的三值比较（-1/0/1） | Go 1.21+，配合 `SortFunc` |
| `strings.Compare` / `bytes.Compare` | 字典序 | 返回 int，不是 bool |

```go
// 错误：编译不过
// s1 == s2  // slice 只能与 nil 比较

// 推荐
slices.Equal(s1, s2)
maps.Equal(m1, m2)

// 嵌套复杂结构、interface{} 字段时仍可能用 DeepEqual
reflect.DeepEqual(a, b)
```

---

### 2.6 复制、合并、插入（Clone / Copy / Concat / Insert）

| 操作 | 旧写法 | 新写法 |
|------|--------|--------|
| 克隆 slice | `append([]T(nil), s...)` | `slices.Clone(s)` |
| 克隆 map | 手写 `for k,v := range m { m2[k]=v }` | `maps.Clone(m)` |
| 复制到已有 buffer | `copy(dst, src)`（内置） | 仍用 `copy` |
| 合并多个 slice | 多次 `append` | `slices.Concat(a, b, c)` |
| 插入元素 | `append(s[:i], append([]T{x}, s[i:]...)...)` | `slices.Insert(s, i, x)` |
| 替换一段 | 手动 `append` 拼接 | `slices.Replace(s, i, j, x...)` |

```go
dup := slices.Clone(original)
m2 := maps.Clone(m1)
all := slices.Concat(part1, part2, part3)
s = slices.Insert(s, 2, 99)
s = slices.Replace(s, 1, 3, 7, 8) // 用 7,8 替换 [1:3)
```

---

### 2.7 压缩、反转、清空（Compact / Reverse / Clear）

| 操作 | 说明 | API |
|------|------|-----|
| 去掉连续重复 | 类似 `uniq` | `slices.Compact` / `CompactFunc` |
| 反转 | 原地反转 | `slices.Reverse` |
| 清空 slice | len 置 0，**保留 cap** | `clear(s)` 或 `s = s[:0]` |
| 清空 map | 删除所有键值对 | `clear(m)`（Go 1.21+）或 `for k := range m { delete(m, k) }` |

```go
s := []int{1, 1, 2, 2, 3}
s = slices.Compact(s) // []int{1, 2, 3}，len 变短

clear(m)  // 清空 map，比逐个 delete 更高效
clear(s)  // 所有元素置为类型零值，len 不变
s = s[:0] // 常用：逻辑清空但保留已分配 cap
```

---

### 2.8 最值与排序辅助（Min / Max / Sorted）

| 操作 | 旧写法 | 新写法 |
|------|--------|--------|
| slice 最小/最大 | 手写循环 | `slices.Min` / `slices.Max`（空 slice panic） |
| 两值比较 | `if a < b` | `cmp.Compare(a, b)` |
| 判断是否已排序 | 手写循环 | `slices.IsSorted` / `IsSortedFunc` |

```go
max := slices.Max(scores) // 空 slice 会 panic，调用前检查 len
cmp.Compare(3, 5)         // -1
```

---

### 2.9 字符串与字节：对称包

Go 将 `string` 与 `[]byte` 的操作分成两个包，API 高度对称：

| 需求 | string | []byte |
|------|--------|--------|
| 包含子串 | `strings.Contains` | `bytes.Contains` |
| 分割 | `strings.Split` | `bytes.Split` |
| 拼接 | `strings.Join` / `Builder` | `bytes.Join` / `Buffer` |
| 修剪 | `strings.TrimSpace` | `bytes.TrimSpace` |

**拼接演进：**

| 方式 | 场景 |
|------|------|
| `+` | 少量字面量拼接 |
| `strings.Builder` | 循环中大量拼接（推荐） |
| `fmt.Sprintf` | 需要格式化，非热点路径 |
| `strings.Join` | 已有 `[]string` |

详见 [字符串操作文档](go-string-operations.md)。

---

### 2.10 数字与字符串互转：strconv vs fmt

| 需求 | 推荐 | 避免 |
|------|------|------|
| int → string | `strconv.Itoa(n)` | `string(n)`（那是 Unicode 码点） |
| 指定进制 | `strconv.FormatInt(n, 16)` | |
| 带格式占位 | `fmt.Sprintf("%05d", n)` | 热点路径大量 Sprintf |
| string → int | `strconv.Atoi` + 检查 `err` | |

---

### 2.11 错误处理：errors.Is / errors.As

Go 1.13 前常用类型断言判断错误类型；现在推荐包装错误 + 标准库探测：

```go
// 旧
if err, ok := err.(*os.PathError); ok { /* ... */ }

// 新
if errors.Is(err, os.ErrNotExist) { /* ... */ }
var pe *os.PathError
if errors.As(err, &pe) { /* ... */ }
```

`fmt.Errorf("...: %w", err)` 包装后，`errors.Is` / `As` 仍能穿透错误链。

---

### 2.12 随机数：math/rand 演进

| 写法 | 说明 |
|------|------|
| `rand.Intn(n)`（Go 1.22+） | 全局源已自动 seed，文档仍建议并发场景用独立 `Rand` |
| `r := rand.New(rand.NewSource(seed))` | 可复现、可并发（每 goroutine 一个实例） |
| `crypto/rand` | 密码学安全随机，API 不同 |

---

### 2.13 上下文与 HTTP：带 Context 的新 API

许多标准库函数在 Go 1.7+ 增加了 `WithContext` 变体，旧函数仍保留：

```go
// 旧
req, _ := http.NewRequest("GET", url, nil)

// 新（推荐）
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

同类模式：`context.WithTimeout`、`context.WithCancel` 取代手动 deadline 管理。

---

### 2.14 并发 map：sync.Map vs map+锁

| 方式 | 适用 |
|------|------|
| `map` + `sync.Mutex` / `RWMutex` | 通用，逻辑清晰 |
| `sync.Map` | 读多写少、key 集合相对稳定 |
| 普通 map 无锁并发写 | **禁止**，runtime 直接 fatal |

这与 API「演进」不同，而是**不同并发模型**的选型，但新人常和 `maps.Clone` 等混淆。

---

### 2.15 总览对照表

| 操作 | 内置 | 旧标准库 | Go 1.21+ 新包 |
|------|------|----------|---------------|
| 排序 | — | `sort.Strings` / `sort.Slice` | `slices.Sort` / `SortFunc` |
| 删 map key | `delete` | — | — |
| 删 slice 元素 | — | `append` 拼接 | `slices.Delete` / `DeleteFunc` |
| 查找 | — | 手写循环 / `sort.Search` | `slices.Contains` / `BinarySearch` |
| 相等 | `==`（有限类型） | `reflect.DeepEqual` | `slices.Equal` / `maps.Equal` |
| 克隆 | `copy` | `append([]T(nil), s...)` | `slices.Clone` / `maps.Clone` |
| 合并 | `append` | — | `slices.Concat` |
| 清空 | `clear` | `for+delete` | `clear` |
| 比较两值 | — | `if a < b` | `cmp.Compare` |

---

## 3. 动手实践

示例目录：[`example/collection-ops/`](../example/collection-ops/)

### 3.1 运行全部演示

```bash
cd example/collection-ops
go run .
```

### 3.2 分模式运行

```bash
go run . -mode sort      # sort vs slices.Sort
go run . -mode delete    # append 删 vs slices.Delete
go run . -mode search    # Contains / Index / BinarySearch
go run . -mode equal     # == 限制 vs slices.Equal / maps.Equal
go run . -mode clone     # Clone / Concat / Insert
go run . -mode clear     # clear(slice) vs clear(map)
go run . -mode cmp       # cmp.Compare + SortFunc
```

### 3.3 预期输出要点

- **sort**：同一组数据，`sort.Strings` 与 `slices.Sort` 结果一致
- **delete**：`slices.Delete` 与 `append` 拼接删元素结果一致
- **search**：`slices.BinarySearch` 在已排序 slice 上返回正确下标
- **equal**：`slices.Equal` 比较内容；`==` 对 slice 只能判 nil
- **clear**：`clear(m)` 后 `len(m)==0`；`s = s[:0]` 保留 cap

### 3.4 自检清单

- [ ] 能说出 `sort.Strings` 与 `slices.Sort` 各自适用的 Go 版本
- [ ] 能写出三种 slice 删除方式并说明是否保持顺序
- [ ] 知道 slice/map 为何不能 `==` 比内容，以及用什么替代
- [ ] 能区分 `clear(s)` 与 `s = s[:0]` 对 cap 的影响
- [ ] `cd example/collection-ops && go run .` 各 `-mode` 无报错

---

## 4. 常见坑与排查

### 4.1 混用新旧 API 导致 go.mod 版本不够

`slices`、`maps`、`cmp` 需要 **Go 1.21+**。若 `go.mod` 写 `go 1.20`，编译报 `package slices is not in GOROOT`。

**修复：** 将 `go.mod` 中 `go` 版本改为 `1.21` 或更高，并安装对应工具链。

### 4.2 `slices.Max` / `slices.Min` 对空 slice panic

```go
var s []int
_ = slices.Max(s) // panic: slices.Max: empty list
```

**修复：** 调用前 `if len(s) > 0`；或手写循环返回 `(value, ok)`。

### 4.3 `slices.Delete` 下标越界

```go
s = slices.Delete(s, 5, 6) // 若 len(s) < 6 会 panic
```

**修复：** 与手动 `append` 删一样，先保证 `0 <= i <= j <= len(s)`。

### 4.4 误以为 `slices.Clone` 深拷贝元素

`Clone` 只复制**切片头 + 底层数组的一份**；若元素本身是指针或 slice，内外层仍共享。

**修复：** 需要深拷贝时，对元素逐层 `Clone` 或自定义拷贝逻辑。

### 4.5 `reflect.DeepEqual` 与 `slices.Equal` 行为差异

`DeepEqual` 对未导出字段、循环引用、`NaN` 等有特殊规则；`slices.Equal` 仅元素 `==`。

**修复：** 简单同类型 slice 用 `slices.Equal`；复杂嵌套用 `DeepEqual` 并理解其语义。

### 4.6 `clear(s)` 与 `s = s[:0]` 混淆

- `clear(s)`：len **不变**，元素置零值
- `s = s[:0]`：len 变 0，cap **保留**（后续 append 可能复用内存）

**修复：** 逻辑清空待复用 buffer 用 `s = s[:0]`；需要零值占位用 `clear(s)`。

### 4.7 在 string 上尝试「原地删除」

```go
// s[0] = 'x'  // 编译错误： string 不可变
```

**修复：** 生成新字符串，见 2.3 节。

---

## 5. 小结与延伸阅读

**要点回顾：**

1. Go 的「多种写法」主要来自**向后兼容 + 泛型现代化**，不是设计混乱
2. Go 1.21+ 新代码：slice/map 操作优先查 **`slices`、`maps`、`cmp`** 与内置 **`clear`**
3. **排序**：`slices.Sort` / `SortFunc` 替代多数 `sort.Strings` / `sort.Slice` 场景
4. **删除**：map 用 `delete`；slice 用 `slices.Delete`；string 只能构造新串
5. **相等**：内容比较用 `slices.Equal` / `maps.Equal`；复杂结构才用 `reflect.DeepEqual`
6. **字符串/字节、strconv/fmt、errors.Is、WithContext** 等也遵循「旧 API 保留、新 API 更安全」的模式

**官方文档：**

- [Go 1.21 Release Notes（slices, maps, cmp）](https://go.dev/doc/go1.21#slices)
- [pkg.go.dev/slices](https://pkg.go.dev/slices)
- [pkg.go.dev/maps](https://pkg.go.dev/maps)
- [pkg.go.dev/cmp](https://pkg.go.dev/cmp)
- [pkg.go.dev/sort](https://pkg.go.dev/sort)

**与本仓库的关系：**

- 前置：[Go 数组、切片与 map](go-arrays-and-slices.md)
- 相关：[Go 字符串操作](go-string-operations.md)、[Go 运算符](go-operators.md)（`==` 限制）
- 示例：[`example/collection-ops/`](../example/collection-ops/)、[`example/maps/`](../example/maps/)
