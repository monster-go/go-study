package main

import (
	"flag"
	"fmt"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, arithmetic, compare, logic, bitwise, assign")
	flag.Parse()

	switch *mode {
	case "arithmetic":
		demoArithmetic()
	case "compare":
		demoCompare()
	case "logic":
		demoLogic()
	case "bitwise":
		demoBitwise()
	case "assign":
		demoAssign()
	default:
		demoArithmetic()
		demoCompare()
		demoLogic()
		demoBitwise()
		demoAssign()
	}
}

func demoArithmetic() {
	fmt.Println("--- 算术运算符 ---")
	a, b := 7, 3
	fmt.Printf("%d + %d = %d\n", a, b, a+b)
	fmt.Printf("%d - %d = %d\n", a, b, a-b)
	fmt.Printf("%d * %d = %d\n", a, b, a*b)
	fmt.Printf("%d / %d = %d (整数除法)\n", a, b, a/b)
	fmt.Printf("%d %% %d = %d\n", a, b, a%b)
	fmt.Printf("float64(%d)/float64(%d) = %.3f\n", a, b, float64(a)/float64(b))
}

func demoCompare() {
	fmt.Println("--- 关系运算符 ---")
	x, y := 10, 20
	fmt.Println("x == y:", x == y)
	fmt.Println("x != y:", x != y)
	fmt.Println("x < y:", x < y)
	fmt.Println("x >= y:", x >= y)
}

func demoLogic() {
	fmt.Println("--- 逻辑运算符 ---")
	ok := true
	fail := false
	fmt.Println("ok && fail:", ok && fail)
	fmt.Println("ok || fail:", ok || fail)
	fmt.Println("!ok:", !ok)
}

func demoBitwise() {
	fmt.Println("--- 位运算符 ---")
	m, n := 12, 10 // 1100, 1010
	fmt.Printf("%d & %d = %d\n", m, n, m&n)
	fmt.Printf("%d | %d = %d\n", m, n, m|n)
	fmt.Printf("%d ^ %d = %d\n", m, n, m^n)
	fmt.Printf("%d << 1 = %d\n", m, m<<1)
	fmt.Printf("%d >> 1 = %d\n", m, m>>1)
}

func demoAssign() {
	fmt.Println("--- 赋值与自增 ---")
	n := 1
	n += 2
	fmt.Println("n += 2 ->", n)
	n++
	fmt.Println("n++ ->", n)
	p := &n
	fmt.Println("*p =", *p, "地址 &n =", &n)
}
