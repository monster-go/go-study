package main

import (
	"flag"
	"fmt"
)

func main() {
	// var total int
	// total = 100
	// label := "score"
	// // label = 200
	// fmt.Println(total, label)

	a, b := 3, 4
	fmt.Println(a / b)
	fmt.Println(float64(a) / float64(b))

	mode := flag.String("mode", "all", "demo mode: all, zero, iota")
	flag.Parse()

	switch *mode {
	case "zero":
		demoZeroValues()
	case "iota":
		demoIota()
	default:
		demoZeroValues()
		demoDeclarations()
		demoConstants()
		demoIota()
		demoConversion()
	}

	fmt.Println(Monday, Tuesday, ReadPerm)
}

func demoZeroValues() {
	var n int
	var ok bool
	var s string
	fmt.Println("zero values:", n, ok, fmt.Sprintf("%q", s))
}

func demoDeclarations() {
	var name string = "Go"
	age := 10
	var x, y int = 1, 2
	x, y = y, x

	fmt.Println("declarations:", name, age, x, y)
}

func demoConstants() {
	const Pi = 3.14159
	const (
		MaxRetries = 3
		AppName    = "vars-demo"
	)
	fmt.Println("constants:", Pi, MaxRetries, AppName)
}

func demoIota() {
	fmt.Println("weekday:", Monday, Tuesday, Saturday)
	fmt.Println("file perm bits:", ReadPerm, WritePerm, ExecPerm)
}

func demoConversion() {
	a, b := 3, 4
	fmt.Println("int div:", a/b, "float div:", float64(a)/float64(b))
}
