# Go 结构体扩展与模拟继承：组合、嵌入与方法提升

> 知识点总结：理解 Go **没有 class 式继承**，通过 struct **嵌入（embedding）** 复用字段与方法；掌握 promoted 访问、外层**重写**嵌入方法、与 OOP 继承的差异；知道何时用嵌入、何时用接口实现多态。

---

## 1. 为什么需要了解这个

从 Java、C++ 转来的开发者常问：「Go 怎么写继承？」答案是：**Go 刻意不提供继承**，用**组合（composition）** 代替 **is-a 继承**。

若按 OOP 习惯硬写，容易踩坑：

- 以为 `Dog` 嵌入 `Animal` 后，`Dog` 就是 `Animal` 的子类型——**不是**，不能赋给 `Animal` 变量
- 嵌入基类方法后，在外层「重写」了同名方法，却期望通过基类接口多态调用——**做不到**，须用 [接口](go-interfaces-and-error-handling.md)
- 多层嵌入 + 同名方法，不清楚调用的是哪一层
- 把嵌入当成「自动继承所有行为」，忽略值/指针接收者对方法集的影响

本篇建立在 [函数、结构体与作用域](go-functions-structs-and-scope.md) 的嵌入小节之后，专门讲「用 struct 模拟继承」的写法、边界与替代方案。

---

## 2. 核心概念

### 2.1 Go 的选择：组合优于继承

| OOP（Java/C++） | Go |
|-----------------|-----|
| `class Dog extends Animal` | `type Dog struct { Animal }`（嵌入） |
| 子类 is-a 父类，可向上转型 | **无子类型关系**，`Dog` 不能当 `Animal` 用 |
| 虚函数 / 多态 | **接口**隐式实现，见 [接口与错误处理](go-interfaces-and-error-handling.md) |
| 重写 override + 多态 dispatch | 外层同名方法**遮蔽** promoted 方法，**无**动态绑定 |

Go 的设计哲学：**用 struct 组合数据，用 interface 抽象行为**。嵌入是语法糖，减少转发样板，不是类型系统里的继承。

### 2.2 最小示例：嵌入 = 字段 + 方法提升

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal // 匿名字段：嵌入
    Breed  string
}

d := Dog{
    Animal: Animal{Name: "Buddy"},
    Breed:  "Labrador",
}

fmt.Println(d.Name)   // Buddy — 字段 promoted
fmt.Println(d.Speak()) // Buddy makes a sound — 方法 promoted
fmt.Println(d.Breed)   // Labrador — 外层自有字段
```

**原理：** 嵌入 `Animal` 后，外层 `Dog` 拥有名为 `Animal` 的字段；`Name` 与 `Speak` 被**提升**到 `Dog`，可写 `d.Name`、`d.Speak()`，等价于 `d.Animal.Name`、`d.Animal.Speak()`。

### 2.3 「重写」方法：外层遮蔽，非多态

外层定义与嵌入类型**同名同签名**的方法时，直接调用走外层；嵌入体上的方法仍可通过前缀访问：

```go
func (d Dog) Speak() string {
    return d.Name + " barks"
}

d.Speak()         // Buddy barks — 外层方法
d.Animal.Speak()  // Buddy makes a sound — 嵌入体原方法
```

这与 Java 的 `@Override` **看起来像**，但有关键区别：

```go
var a Animal = d // 编译错误：Dog 不能作为 Animal 使用
```

`Dog` 不是 `Animal` 的子类型。外层 `Speak` 只是**静态绑定**到 `Dog`，不存在「基类引用调用子类实现」。

### 2.4 用接口实现真正的多态

若需要「多种动物，统一 Speak()，运行时选实现」，应定义接口 + 各类型实现：

```go
type Speaker interface {
    Speak() string
}

func Greet(s Speaker) {
    fmt.Println(s.Speak())
}

Greet(Dog{Animal: Animal{Name: "Buddy"}})     // Buddy barks
Greet(Animal{Name: "Kitty"})                  // Kitty makes a sound
```

**选用原则：**

| 需求 | 做法 |
|------|------|
| 复用字段、少写 `d.Animal.Xxx` | 嵌入 struct |
| 复用方法且允许外层改行为 | 嵌入 + 外层同名方法 |
| 多种类型同一 API、可替换 | **接口** |
| 严格 is-a、单继承树 | Go 不支持，用组合 + 接口 |

### 2.5 多层扩展

嵌入可以链式叠加，每一层只 promote **一层**直接嵌入的类型：

```go
type Mammal struct {
    Animal
    WarmBlooded bool
}

type Dog struct {
    Mammal
    Breed string
}

d := Dog{
    Mammal: Mammal{
        Animal:      Animal{Name: "Buddy"},
        WarmBlooded: true,
    },
    Breed: "Labrador",
}

fmt.Println(d.Name)        // Buddy — 来自 Animal，经 Mammal promote
fmt.Println(d.WarmBlooded) // true — 来自 Mammal
fmt.Println(d.Breed)       // Labrador
```

多层时若各层有同名方法，规则与 [同名字段](go-functions-structs-and-scope.md) 相同：**唯一**则 promote；**歧义**则必须写前缀（如 `d.Animal.Speak()` vs `d.Mammal.Speak()`，若两层都未重写且仅 Animal 有 `Speak`，则 `d.Speak()` 仍唯一）。

### 2.6 值嵌入 vs 指针嵌入

嵌入 `Animal` 与嵌入 `*Animal` 对方法集有影响（与 [接口实现](go-interfaces-and-error-handling.md) 一致）：

```go
type DogByValue struct {
    Animal
}

type DogByPtr struct {
    *Animal
}

var _ Speaker = DogByValue{Animal: Animal{Name: "A"}} // OK，值接收者方法在值上

// 若 Animal.Speak 只有指针接收者 func (a *Animal) Speak()：
// DogByValue 无法通过 *Animal 的指针接收者方法满足 Speaker（除非 Animal 也有值接收者 Speak）
```

**实践建议：** 嵌入的类型若方法多为 `*T` 接收者，考虑嵌入 `*Animal` 并在构造时保证非 nil。

### 2.7 与「继承」对照表

| 能力 | OOP 继承 | Go 嵌入 |
|------|----------|---------|
| 复用字段 | ✓ | ✓（promote） |
| 复用方法 | ✓ | ✓（promote） |
| 子类型 / 向上转型 | ✓ | ✗ |
| 重写 + 基类引用多态 | ✓ | ✗（须接口） |
| 多重继承 | 语言相关 | 可嵌入多个 struct（同名字段/方法受歧义规则约束） |
| 构造链 super() | ✓ | 手动初始化嵌入字段：`Dog{Animal: Animal{...}}` |

---

## 3. 应用场景

### 3.1 领域模型：通用基类 + 特化子类

```go
type Timestamped struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

type User struct {
    Timestamped
    ID   int64
    Name string
}

type Order struct {
    Timestamped
    ID     int64
    Amount float64
}
```

**解决什么问题：** `CreatedAt` / `UpdatedAt` 只写一次，多个业务 struct 共享；无需继承树，各 struct 独立。

### 3.2 扩展标准库或第三方类型（谨慎）

```go
type MyBuffer struct {
    bytes.Buffer
}

func (m *MyBuffer) WriteString(s string) (int, error) {
    return m.Write([]byte(s))
}
```

**注意：** 嵌入 `bytes.Buffer` 会 promote 其全部方法；外层新增方法不能与 promoted 方法冲突（签名相同则视为重复定义）。更常见的扩展方式是**包装（wrapper）** 具名字段 + 显式转发，而非嵌入。

### 3.3 HTTP Handler 组合

标准库常见模式：嵌入 `http.HandlerFunc` 或结构体组合 middleware，在保留内层行为的同时加外层逻辑——本质是**装饰**，不是继承。并发与 HTTP 专题可再展开。

---

## 4. 动手实践

### 4.1 跟着写：动物层级

```go
package main

import "fmt"

type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + ": ..."
}

type Dog struct {
    Animal
}

func (d Dog) Speak() string {
    return d.Name + ": woof"
}

type Cat struct {
    Animal
}

func (c Cat) Speak() string {
    return c.Name + ": meow"
}

type Speaker interface {
    Speak() string
}

func main() {
    animals := []Speaker{
        Dog{Animal: Animal{Name: "Buddy"}},
        Cat{Animal: Animal{Name: "Whiskers"}},
        Animal{Name: "Generic"},
    }
    for _, a := range animals {
        fmt.Println(a.Speak())
    }
}
```

**预期输出：**

```
Buddy: woof
Whiskers: meow
Generic: ...
```

### 4.2 自检清单

- [ ] 能解释 `d.Name` 与 `d.Animal.Name` 等价
- [ ] 能说明为何 `Animal d = Dog{}` 在 Go 中不成立
- [ ] 能写出外层 `Speak` 遮蔽嵌入 `Speak` 的示例
- [ ] 需要多态时，能改为 `[]Speaker` + 接口

---

## 5. 常见误区

### 5.1 把嵌入当成继承类型

```go
func Feed(a Animal) { /* ... */ }
Feed(Dog{...}) // 编译错误
```

**修复：** 参数改为接口，或改为 `Feed(d Dog)` 接受具体类型。

### 5.2 期望 promoted 方法「虚调用」

```go
type Base struct{}
func (Base) Foo() { fmt.Println("base") }

type Derived struct{ Base }
func (Derived) Foo() { fmt.Println("derived") }

var b Base = Derived{} // 非法

d := Derived{}
d.Foo()      // derived
d.Base.Foo() // base — 无多态，两条路径静态可见
```

### 5.3 嵌入 nil 指针

```go
type Dog struct {
    *Animal
}

d := Dog{} // d.Animal == nil
d.Speak()  // panic: nil pointer dereference
```

**修复：** 构造时赋值：`Dog{Animal: &Animal{Name: "x"}}`。

### 5.4 同名字段 / 方法歧义

多个嵌入体提供同名 promoted 成员时，须用类型前缀，详见 [函数、结构体与作用域 — 多个嵌入体有同名字段](go-functions-structs-and-scope.md)。

---

## 6. 小结与延伸阅读

**要点回顾：**

1. Go **无 class 继承**；用 struct **嵌入**复用字段与方法（promotion）
2. 外层同名方法**遮蔽**嵌入方法，**不是** OOP 意义上的 override 多态
3. 子类型与向上转型 **不存在**；多态靠 **interface**
4. 构造嵌入 struct 须显式初始化嵌套字段；指针嵌入注意 nil
5. 同名字段/方法冲突时用嵌入类型名作前缀，或从设计上改名

**官方文档：**

- [A Tour of Go：Methods and pointer indirection](https://go.dev/tour/methods/4)
- [Effective Go：Embedding](https://go.dev/doc/effective_go#embedding)
- [Go 语言规范：Struct types](https://go.dev/ref/spec#Struct_types)

**与本仓库的关系：**

- 前置：[Go 函数、结构体与作用域](go-functions-structs-and-scope.md)（嵌入、同名字段、方法接收者）
- 后续：[Go 接口与错误处理](go-interfaces-and-error-handling.md)（多态与隐式实现）
