package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, func, struct, scope, defer, closure, builtin")
	flag.Parse()

	switch *mode {
	case "func":
		demoFunc()
	case "struct":
		demoStruct()
	case "scope":
		demoScope()
	case "defer":
		demoDefer()
	case "defer2":
		demoDefer2()
	case "closure":
		demoClosure()
	case "builtin":
		demoBuiltin()
	case "file":
		readFile()
	default:
		demoFunc()
		demoStruct()
		demoScope()
		demoDefer()
		demoClosure()
		demoBuiltin()
	}
}

// --- 函数 ---

func Add(a, b int) int {
	return a + b
}

func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为 0")
	}
	return a / b, nil
}

func demoFunc() {
	fmt.Println("--- 函数 ---")
	fmt.Printf("Add(2,3)=%d Sum(1,2,3)=%d\n", Add(2, 3), Sum(1, 2, 3))
	v, err := Div(10, 2)
	fmt.Printf("Div(10,2)=%.2f err=%v\n", v, err)
}

// --- 结构体 ---

type Person struct {
	Name string
	Age  int
}

func (p Person) Greet() string {
	return "Hi, " + p.Name
}

func (p *Person) Birthday() {
	p.Age++
}

func demoStruct() {
	fmt.Println("--- 结构体 ---")
	p := Person{Name: "Alice", Age: 30}
	fmt.Println(p.Greet(), "age=", p.Age)
	p.Birthday()
	fmt.Println("生日后 age=", p.Age)
}

// --- 作用域 ---

func demoScope() {
	fmt.Println("--- 作用域 ---")
	count := 10
	if true {
		count := 20
		fmt.Println("块内 count=", count)
	}
	fmt.Println("块外 count=", count)
}

// --- defer ---

func demoDefer() {
	fmt.Println("--- defer ---")
	fmt.Print("defer 注册顺序 vs 执行顺序: ")
	func() {
		defer fmt.Print(1, " ")
		defer fmt.Print(2, " ")
		defer fmt.Print(3, " ")
	}()
	fmt.Println()

	fmt.Print("defer 参数求值: 注册时 i=")
	func() {
		i := 0
		defer fmt.Println(i)
		i++
	}()
}

func calc(index string, a, b int) int {
	ret := a + b
	fmt.Println(index, a, b, ret)
	return ret
}

func demoDefer2() {
	x := 1
	y := 2
	defer calc("AA", x, calc("A", x, y))
	x = 10
	defer calc("BB", x, calc("B", x, y))
	y = 20
}

// --- 闭包 ---

func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func demoClosure() {
	fmt.Println("--- 闭包 ---")
	next := counter()
	fmt.Println("counter:", next(), next(), next())
}

// --- 内置函数 ---

func demoBuiltin() {
	fmt.Println("--- 内置函数 ---")
	s := make([]int, 0, 5)
	s = append(s, 1, 2, 3)
	fmt.Printf("len=%d cap=%d append后 len=%d\n", len(s), cap(s), len(append(s, 4)))
}

// 读文件
func readFile() {
	file, err:= os.Open("example.txt")
	if(err != nil) {
		fmt.Println("打开文件失败", err)
		return
	}

	defer file.Close()

	// 读取文件第一行 并打印
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if(err != nil) {
		fmt.Println("读取文件失败", err)
		return
	}
	fmt.Println("读取文件成功", string(buf[:n]))
}

func ParsePositive(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %v", s, err)
	}
	if i <= 0 {
		return 0, fmt.Errorf("value is not positive: %d", i)
	}
	return i, nil
}
