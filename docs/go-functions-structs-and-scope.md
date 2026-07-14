# Go 函数、结构体与作用域：从声明到 defer、闭包与内置函数

> 知识点总结：掌握函数声明与多返回值、结构体定义与初始化、块级作用域与变量遮蔽；理解 `defer` 的执行时机与 LIFO 顺序、闭包如何捕获外部变量、匿名函数的写法，以及 `len`/`make`/`append` 等内置函数的语义与适用场景。

---

## 1. 为什么需要了解这个

函数与结构体是 Go 组织逻辑与数据的两大支柱。新人常在这些地方卡住：

- 分不清「值接收者」与「指针接收者」，改 struct 字段不生效
- 以为 `defer` 在函数**返回后**才执行参数表达式——实际上**注册 defer 时**参数就已求值
- 闭包「记住」外部变量，循环里启动 goroutine 时全部看到同一个 `i`（Go 1.22 前常见）
- 把 `new(T)` 当成 C 的 `malloc`，或把 `make` 用于 struct
- 作用域搞不清：`:=` 在 inner 块里「遮蔽」外层同名变量，改的是副本

本篇建立在 [语句控制](go-control-flow.md) 与 [变量、类型与常量](go-variables-types-and-constants.md) 之后，是写可复用逻辑、定义自定义类型的基础。后续接口、错误处理、并发都会反复用到函数、闭包与 struct。

---

## 2. 核心概念

### 2.1 函数（Function）

#### 基本语法

```go
func Add(a, b int) int {
    return a + b
}
```

| 要素 | 说明 |
|------|------|
| 参数 | 相邻同类型可合并：`a, b int` |
| 返回值 | 可多个；可命名，命名返回值在函数入口处相当于已 `var` 声明 |
| 无返回值 | 可省略返回类型，或写 `()` |
| 可见性 | **首字母大写**导出（包外可见），小写仅包内 |

#### 多返回值

Go 原生支持多返回值，错误处理惯用法：

```go
func Div(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为 0")
    }
    return a / b, nil
}
```

#### 命名返回值（Named Result Parameters）

```go
func RectArea(w, h float64) (area float64) {
    area = w * h
    return // 裸 return，等价于 return area
}
```

**原理：** 命名返回值在函数体开头按零值初始化，作用域是整个函数；裸 `return` 返回当前命名变量的值。适合短函数，复杂逻辑更推荐显式 `return v, err` 以免可读性下降。

#### 变参函数（Variadic）

```go
func Sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

Sum(1, 2, 3)
Sum([]int{1, 2, 3}...) // 切片展开
```

`nums` 在函数内类型为 `[]int`；`...` 必须是最后一个参数。

#### 函数是一等公民

函数类型可赋值、传参、作为返回值：

```go
type Op func(int, int) int

var calc Op = Add
```

**比较规则：** 函数值**只能与 `nil` 比较**，不能用 `==` 比较两个函数是否「逻辑相同」（除非两者都是 `nil`）。

---

### 2.2 结构体（Struct）

结构体把多个字段组合成一种类型，是 Go 实现「对象风格」数据封装的基础（无 class 关键字）。

#### 定义与字面量

```go
type Person struct {
    Name string
    Age  int
}

p1 := Person{Name: "Alice", Age: 30} // 字段名可省略，按顺序填
p2 := Person{"Bob", 25}
var p3 Person // 零值：Name="", Age=0
```

| 初始化方式 | 说明 |
|------------|------|
| 字段名初始化 | `Person{Name: "x", Age: 1}`，推荐，顺序无关 |
| 按序初始化 | `Person{"x", 1}`，字段顺序必须一致 |
| 零值 | `var p Person`，各字段为类型零值 |
| `new` | `p := new(Person)` 返回 `*Person`，字段为零值 |
| 取址字面量 | `p := &Person{Name: "x"}`，得 `*Person` |

#### 字段访问与指针

```go
p := Person{Name: "Carol"}
p.Age = 28

pp := &p
pp.Age = 29 // 等价于 (*pp).Age，Go 自动解引用
```

**原理：** 对指针 `pp` 用 `.` 访问字段时，编译器自动写 `(*pp).Field`，无需 C 式 `->`。

#### 嵌套与匿名字段（Embedding）

```go
type Address struct {
    City string
}

type Employee struct {
    Person        // 匿名字段：嵌入
    Address
    Dept string
}

e := Employee{
    Person:  Person{Name: "Dave"},
    Address: Address{City: "Beijing"},
}
fmt.Println(e.Name, e.City) //  promoted field，可直接访问
```

嵌入的类型字段会**提升（promote）**到外层，可直接 `e.Name`；若外层有同名字段则外层优先。

##### 多个嵌入体有同名字段怎么办？

当**两个及以上**匿名字段都含有同名 promoted 字段时，编译器无法判断你要访问哪一个，直接写 `e.Name` 会报错：

```text
ambiguous selector e.Name
```

下面用 `Worker` 与 `Contact` 都带 `Name` 字段的场景说明常见处理方式。

**1. 用嵌入类型名作前缀（最常用）**

不冲突的字段仍可 promoted 直接访问；冲突字段必须带上嵌入类型名：

```go
type Worker struct {
    Name string
    Role string
}

type Contact struct {
    Name  string
    Email string
}

type Employee struct {
    Worker
    Contact
    Dept string
}

e := Employee{
    Worker:  Worker{Name: "Dave", Role: "Engineer"},
    Contact: Contact{Name: "Dave Zhang", Email: "d@co.com"},
    Dept:    "Platform",
}

// fmt.Println(e.Name)       // 编译错误：ambiguous selector e.Name
fmt.Println(e.Worker.Name)   // Dave — 工作档案里的名字
fmt.Println(e.Contact.Name)  // Dave Zhang — 通讯录里的名字
fmt.Println(e.Role)          // Engineer — 仅 Worker 有，可 promoted
fmt.Println(e.Email)         // d@co.com — 仅 Contact 有，可 promoted
fmt.Println(e.Dept)          // Platform — 外层自有字段
```

**原理：** 嵌入后外层同时拥有 `Worker` 与 `Contact` 两个**命名字段**（类型名即字段名）。`e.Worker.Name` 是明确路径；只有**唯一**能被 promote 的字段才允许省略前缀。

**2. 外层自己声明同名字段（外层优先）**

若业务上需要一个「默认展示名」，可在外层再声明 `Name`，它会遮蔽所有嵌入体的 `Name`：

```go
type EmployeeWithDisplayName struct {
    Name string // 外层字段，优先级最高
    Worker
    Contact
}

e := EmployeeWithDisplayName{
    Name:    "Dave (Display)",
    Worker:  Worker{Name: "Dave"},
    Contact: Contact{Name: "Dave Zhang"},
}

fmt.Println(e.Name)          // Dave (Display) — 外层
fmt.Println(e.Worker.Name)   // Dave — 仍须通过前缀访问嵌入体
fmt.Println(e.Contact.Name)  // Dave Zhang
```

此时 `e.Name` 合法，因为外层已提供唯一答案；嵌入体里的 `Name` 仍须用 `e.Worker.Name` / `e.Contact.Name` 区分。

**3. 改为具名字段，放弃嵌入（彻底避免 promote 冲突）**

若两个子 struct 字段高度重叠、很少需要 promoted 访问，可不用匿名字段，改用普通命名字段：

```go
type EmployeeExplicit struct {
    W Worker
    C Contact
}

e := EmployeeExplicit{
    W: Worker{Name: "Dave", Role: "Engineer"},
    C: Contact{Name: "Dave Zhang", Email: "d@co.com"},
}

fmt.Println(e.W.Name, e.C.Name) // 始终带前缀，语义最清晰
// fmt.Println(e.Name)          // 编译错误：EmployeeExplicit 没有 Name
```

**4. 在定义阶段消除重名（治本，改源类型）**

若 `Worker.Name` 与 `Contact.Name` 语义不同，可在源 struct 里用不同字段名，嵌入后自然不再冲突：

```go
type Worker struct {
    StaffName string // 原名 Name → StaffName
    Role      string
}

type Contact struct {
    LegalName string // 原名 Name → LegalName
    Email     string
}

type EmployeeRenamed struct {
    Worker
    Contact
}

e := EmployeeRenamed{
    Worker:  Worker{StaffName: "Dave", Role: "Engineer"},
    Contact: Contact{LegalName: "Dave Zhang", Email: "d@co.com"},
}

fmt.Println(e.StaffName, e.LegalName) // 均可 promoted，互不冲突
```

| 方式 | 适用场景 |
|------|----------|
| `e.Worker.Name` 前缀 | 保留嵌入、偶尔区分来源，**首选** |
| 外层声明同名字段 | 需要统一的「默认字段」，其余仍用前缀 |
| 具名字段 `W Worker` | 不想 promote，调用处始终显式 |
| 源 struct 改字段名 | 语义本就不同，从设计上避免重名 |

**小结：** promote 是语法糖，不是魔法——**唯一**时省略前缀，**歧义**时必须写清路径；方法提升（同名 method）规则与字段相同。

#### 结构体扩展（模拟继承）

Go 没有 `class` 与 `extends`，但可以通过**嵌入 struct** 复用基类的字段与方法，在外层**重写**同名方法，形成类似「基类 → 子类」的扩展写法。

需要区分两点：

- **像继承的部分：** 字段与方法 promoted、外层可定义同名方法覆盖默认行为
- **不像继承的部分：** `Dog` 嵌入 `Animal` 后**并不是** `Animal` 的子类型，不能向上转型；运行时多态须用**接口**

完整对照、多层扩展、值/指针嵌入与常见误区见专题：[Go 结构体扩展与模拟继承](go-struct-extension-and-inheritance.md)。

#### 方法（Method）简述

方法是带接收者（receiver）的函数：

```go
func (p Person) Greet() string {
    return "Hi, " + p.Name
}

func (p *Person) Birthday() {
    p.Age++
}
```

| 接收者 | 调用时 | 能否改原 struct |
|--------|--------|-----------------|
| 值 `T` | 复制一份 | 不能改调用者的字段（改的是副本） |
| 指针 `*T` | 可传值或指针 | 能改原 struct |

**选用原则：** 需要修改接收者、或 struct 较大避免复制、或与 `interface` 实现一致时用 `*T`；小且不可变的数据可用值接收者。

---

### 2.3 作用域（Scope）

作用域决定标识符在何处可见。Go 主要有：

| 作用域 | 范围 | 示例 |
|--------|------|------|
| **包（package）** | 整个 `.go` 文件所属包 | 包级 `var`、`const`、`func`、类型 |
| **文件** | 无单独文件作用域；同一包多文件共享包作用域 | — |
| **函数** | 函数体 | 参数、局部变量 |
| **块（block）** | `{}` 内 | `if`、`for`、`switch` 的子块 |

#### 块级短声明与 if 初始化

```go
if x := compute(); x > 0 {
    fmt.Println(x)
}
// x 在此不可见
```

与 [语句控制](go-control-flow.md) 中 `if` 短声明一致：**变量生命周期限于该块**。

#### 变量遮蔽（Shadowing）

内层块用 `:=` 声明与外层**同名**变量时，内层**遮蔽**外层：

```go
count := 10
if true {
    count := 20 // 新变量，外层 count 仍为 10
    fmt.Println(count)
}
fmt.Println(count) // 10
```

**原理：** `:=` 左侧至少有一个新名字时合法；若 `count` 在外层已存在，内层 `count := 20` 会声明**新的** `count`，外层不变。排查 bug 时可用 `go vet` 的 shadow 检查（需安装 shadow 工具）或仔细区分 `=` 与 `:=`。

#### 导出与可见性

- 包级标识符：**大写开头** → 导出（其他包可引用）
- **小写** → 仅本包可见

无 `public`/`private` 关键字，靠命名约定实现封装。

---

### 2.4 defer

`defer` 将函数调用**推迟**到当前函数返回前执行，常用于释放资源、解锁、关闭文件。

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // 注册：返回前关闭
    // ... 使用 f
    return nil
}
```

#### 执行顺序：LIFO（后进先出）

```go
defer fmt.Println(1)
defer fmt.Println(2)
defer fmt.Println(3)
// 输出：3, 2, 1
```

#### 参数求值时机：注册时，而非执行时

```go
i := 0
defer fmt.Println(i) // 打印 0：注册时 i 的值已拷贝
i++
return // 先打印 0，再返回
```

**原理：** `defer` 语句执行时，**参数表达式立即求值**，结果保存在 defer 记录里；真正调用发生在 `return` 路径上。若要在返回时看**最新**变量，传闭包：

```go
i := 0
defer func() { fmt.Println(i) }() // 返回前执行，打印 1
i++
```

#### defer 与返回值

命名返回值 + defer 可修改**有名**返回变量（理解即可，复杂逻辑慎用）：

```go
func f() (n int) {
    defer func() { n++ }()
    return 1 // 最终返回 2：return 先设 n=1，再跑 defer
}
```

---

### 2.5 匿名函数与闭包（Closure）

#### 匿名函数（Function Literal）

没有名字、可直接调用的函数值：

```go
square := func(x int) int { return x * x }
fmt.Println(square(5))

func() {
    fmt.Println("立即执行")
}()
```

#### 闭包

匿名函数**引用外部变量**时形成闭包；外部变量被「捕获」，在函数返回后仍有效：

```go
func counter() func() int {
    n := 0
    return func() int {
        n++
        return n
    }
}

next := counter()
fmt.Println(next(), next(), next()) // 1 2 3
```

**原理：** 返回的函数闭包持有 `n` 的**引用**（逃逸到堆上），每次调用共享同一 `n`。

#### 循环变量陷阱（Go 1.22 前）

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i) // Go 1.21 及以前：可能三次都打印 3
    }()
}
```

**原因（Go 1.21 及以前）：** 循环变量 `i` 在每次迭代中**复用同一地址**，goroutine 异步读到的是循环结束后的值。

**修复（旧版本）：** 迭代内复制：`i := i` 或 `go func(i int) { ... }(i)`。

**Go 1.22+：** 每轮 `for` 迭代有**独立的**循环变量，上述问题已修复。写库时若需兼容旧版，仍建议显式传参。

---

### 2.6 内置函数（Built-in Functions）

内置函数由编译器识别，**无需 import**。完整列表见 [Go 规范：Built-in functions](https://go.dev/ref/spec#Built-in_functions)。

#### 长度与容量

| 函数 | 作用 | 典型类型 |
|------|------|----------|
| `len(v)` | 元素个数 / 字节数 | string、array、slice、map、channel |
| `cap(v)` | 容量 | slice、array、channel |

```go
s := []int{1, 2, 3}
len(s) // 3
cap(s) // 底层数组容量，≥ len
```

#### 构造与内存

| 函数 | 作用 | 注意 |
|------|------|------|
| `make(T, args...)` | 分配并**初始化** slice、map、channel | 返回**已就绪**的引用类型，非指针 |
| `new(T)` | 分配零值内存 | 返回 `*T`，常用于需要指针的场合 |

```go
m := make(map[string]int)
sl := make([]int, 0, 10)
ch := make(chan int, 1)

p := new(int) // *int，值为 0
```

**原理对比：**

- `make` 只用于 **slice、map、channel**，创建的是「可用」的引用类型头（如 slice 的 ptr/len/cap）。
- `new(T)` 适用于**任意类型**，分配零值并返回指针；对 struct 更常写 `&T{}` 而非 `new(T)`，语义相同。

#### slice 操作

| 函数 | 作用 |
|------|------|
| `append(s, elems...)` | 追加元素，可能触发扩容并换底层数组 |
| `copy(dst, src)` | 复制元素，返回复制个数 `min(len(dst), len(src))` |

#### map 与 channel

| 函数 | 作用 |
|------|------|
| `delete(m, key)` | 删除 map 键；键不存在也安全 |
| `close(ch)` | 关闭 channel，发送方习惯在 defer 中 close |

#### panic / recover

| 函数 | 作用 |
|------|------|
| `panic(v)` | 触发 panic，开始栈展开 |
| `recover()` | 在 **defer 函数内**调用，捕获 panic，返回 panic 值 |

惯用法：只在少数场景（如 HTTP 中间件、测试）用 `recover`；业务错误优先 `return error`。

#### 其他

| 函数 | 说明 |
|------|------|
| `print` / `println` | 调试用，输出到 stderr，格式未保证；生产用 `fmt` |
| `complex` / `real` / `imag` | 复数 |
| `min` / `max` | Go 1.21+，有序类型比较 |
| `clear` | Go 1.21+，清空 map 或 slice 元素置零 |

查看文档：

```bash
go doc -u builtin
```

`-u` 显示未导出符号，才能看到内置的小写函数名。

---

## 3. 应用场景

学完语法后，更关键的是知道**什么时候该用**。下面按本篇主题列出常见工程场景（解决什么问题 → 怎么用）。

### 3.1 闭包：带状态、可配置的「函数工厂」

闭包解决的问题是：**既要记住一些状态或配置，又不想为此单独定义一个 struct + 方法**；调用方只拿一个 `func(...)` 就能用。

| 场景 | 解决什么问题 | 典型写法 |
|------|--------------|----------|
| **带状态的函数工厂** | 每次创建独立计数器 / 限流器，互不干扰 | 上文 `counter()`：`a := counter(); b := counter()` |
| **HTTP 中间件 / 装饰器** | 外层接收配置（超时、logger），内层统一包一层请求处理 | `func WithTimeout(d time.Duration) func(http.Handler) http.Handler` |
| **延迟绑定配置** | 创建时捕获依赖，多次调用复用同一配置 | `func NewGreeter(prefix string) func(string) string { return func(name string) string { return prefix + name } }` |

最短骨架（中间件风格）：

```go
func WithPrefix(prefix string) func(string) string {
    return func(msg string) string {
        return prefix + msg // 闭包捕获 prefix
    }
}

log := WithPrefix("[app] ")
fmt.Println(log("started")) // [app] started
```

**何时不用闭包：** 状态字段多、需要多方法协作时，用 struct + 方法更清晰；闭包适合「一个函数 + 少量捕获」的场景。

### 3.2 命名返回值：`defer` 改写结果

命名返回值真正有优势的场景是：**在 `defer` 里统一收尾并修改返回值**（例如包装 `error`、补默认值）。匿名返回值做不到「defer 里改结果槽」。

```go
func ReadConfig(path string) (cfg Config, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("read config %s: %w", path, err)
        }
    }()
    // ... 多处可能 return ..., err
    return cfg, err
}
```

短函数、逻辑简单时可用裸 `return`；复杂分支更推荐显式 `return v, err`。

### 3.3 defer：资源释放与收尾

`defer` 解决的问题是：**无论函数从哪条路径返回，都能保证清理逻辑执行**（关文件、解锁、还原状态）。

| 场景 | 解决什么问题 |
|------|--------------|
| `defer f.Close()` / `defer resp.Body.Close()` | 避免漏关、提前 return 导致泄漏 |
| `defer mu.Unlock()` | 加锁后任一路径都能解锁 |
| `defer` + `recover` | 在包边界吞掉 panic，转成错误日志 |

### 3.4 结构体与方法：数据 + 行为打包

| 场景 | 解决什么问题 |
|------|--------------|
| 领域模型（`User`、`Order`） | 把相关字段绑在一起，避免一长串散落参数 |
| 指针接收者改字段 | 方法需要改接收者状态时（如 `SetAge`） |
| 嵌入（Embedding） | 复用已有类型的字段/方法，少写转发样板 |
| 模拟继承 / 扩展 | 嵌入 + 方法重写；多态用接口，见 [结构体扩展与模拟继承](go-struct-extension-and-inheritance.md) |

### 3.5 多返回值与 `(T, error)`

Go 用多返回值表达「结果 + 是否失败」，而不是异常控制流。几乎所有可能失败的 I/O、解析、RPC 调用都会返回 `error`，调用方显式检查。

---

## 4. 动手实践

### 4.1 运行仓库示例

```bash
cd example/functions
go run .                    # 运行全部 demo
go run . -mode func           # 函数与多返回值
go run . -mode struct         # 结构体与方法
go run . -mode scope          # 作用域与遮蔽
go run . -mode defer          # defer 顺序与求值
go run . -mode closure        # 闭包与匿名函数
go run . -mode builtin        # 内置函数
```

**预期（`-mode all` 片段）：**

```
--- 函数 ---
Add(2,3)=5 Sum(1,2,3)=6
Div(10,2)=5.00 err=<nil>
--- 结构体 ---
Hi, Alice age=30
生日后 age=31
--- 作用域 ---
块内 count= 20
块外 count= 10
--- defer ---
defer 注册顺序 vs 执行顺序: 3 2 1
defer 参数求值: 注册时 i=0
--- 闭包 ---
counter: 1 2 3
--- 内置函数 ---
len=3 cap=5 append后 len=4
```

### 4.2 跟着写：带 error 的解析函数

```go
func ParsePositive(s string) (int, error) {
    n, err := strconv.Atoi(s)
    if err != nil {
        return 0, fmt.Errorf("parse %q: %w", s, err)
    }
    if n <= 0 {
        return 0, fmt.Errorf("want positive, got %d", n)
    }
    return n, nil
}
```

调用方始终检查 `err`，这是 Go 的常规风格。

### 4.3 跟着写：defer 关闭资源

```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

即使中间有 `return`，`Close` 也会在返回前执行。

### 4.4 自检清单

- [ ] 能写出多返回值函数并正确处理 `error`
- [ ] 能解释值接收者与指针接收者的区别
- [ ] 能说明 `defer` 的执行顺序与参数何时求值
- [ ] 能写一个返回闭包的工厂函数并解释变量如何被捕获
- [ ] 能说出闭包至少 2 个应用场景（如函数工厂、中间件/装饰器）
- [ ] 能区分 `make` 与 `new` 的适用类型
- [ ] `go run .` 各 `-mode` 无报错

---

## 5. 常见坑与排查

### 5.1 值接收者修改 struct 无效

```go
func (p Person) SetAge(a int) { p.Age = a } // 改副本
p.SetAge(99) // p.Age 不变
```

**修复：** 改为指针接收者 `func (p *Person) SetAge(a int)`，调用时 Go 自动取址。

### 5.2 defer 在循环里注册导致资源晚释放

```go
for _, f := range files {
    open(f)
    defer close() // 所有 close 堆到函数结束才执行
}
```

**修复：** 用匿名函数包一层，使 defer 在每次迭代结束执行：

```go
for _, f := range files {
    func() {
        handle := open(f)
        defer handle.Close()
        // ...
    }()
}
```

或提取为独立函数 `func process(f string) { ... defer ... }`。

### 5.3 闭包捕获循环变量（Go 1.21 及以前）

见 2.5 节。验证：打印 goroutine 中的下标，对照你的 Go 版本。

### 5.4 对 struct 使用 make

```go
p := make(Person) // 编译错误：make 不能用于 struct
```

**修复：** `p := Person{}` 或 `p := new(Person)` 或 `p := &Person{}`。

### 5.5 shadow 导致外层变量未改

```go
err := do()
if err != nil {
    err := fmt.Errorf("wrap: %w", err) // 新 err，外层 err 未更新
    return err
}
```

若目的是更新外层 `err`，用 `=` 而非 `:=`：`err = fmt.Errorf(...)`。

### 5.6 recover 写在 defer 外无效

```go
recover() // 无效
```

**修复：** 必须在 defer 的函数体内：

```go
defer func() {
    if r := recover(); r != nil {
        log.Println("recovered:", r)
    }
}()
```

---

## 6. 小结与延伸阅读

**要点回顾：**

1. 函数可多返回值；错误惯用法是 `(T, error)`；首字母控制包外可见性
2. struct 组合数据；匿名字段嵌入可 promoted；改字段常用指针接收者；扩展与「模拟继承」见 [专题文档](go-struct-extension-and-inheritance.md)
3. 作用域由块决定；`:=` 在内层可遮蔽外层同名变量
4. `defer` 函数返回前 LIFO 执行；**参数在 defer 注册时求值**
5. 匿名函数引用外部变量形成闭包；Go 1.22+ 循环变量每轮独立
6. `make` 用于 slice/map/channel；`new` 返回指向零值的 `*T`
7. 内置函数用 `go doc -u builtin` 查阅

**官方文档：**

- [A Tour of Go：Functions](https://go.dev/tour/basics/4)
- [A Tour of Go：More Types: Structs](https://go.dev/tour/moretypes/2)
- [Go 语言规范：Function declarations](https://go.dev/ref/spec#Function_declarations)
- [Go 语言规范：Built-in functions](https://go.dev/ref/spec#Built-in_functions)
- [Effective Go：Defer, Panic, and Recover](https://go.dev/doc/effective_go#defer)

**与本仓库的关系：**

- 上一篇：[Go 数组、切片与 map](go-arrays-and-slices.md)
- 示例：[`example/functions/`](../example/functions/)
- 专题：[Go 结构体扩展与模拟继承](go-struct-extension-and-inheritance.md)（嵌入、方法重写与接口多态）
- 下一篇：[Go 接口与错误处理](go-interfaces-and-error-handling.md)
