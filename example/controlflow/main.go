package main

import (
	"flag"
	"fmt"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, if, for, switch, range")
	flag.Parse()

	switch *mode {
	case "if":
		demoIf()
	case "for":
		demoFor()
	case "switch":
		demoSwitch()
	case "range":
		demoRange()
	default:
		demoIf()
		demoFor()
		demoSwitch()
		demoRange()
	}
}

func demoIf() {
	fmt.Println("--- if / else ---")
	n := 7
	if n%2 == 0 {
		fmt.Println(n, "是偶数")
	} else {
		fmt.Println(n, "是奇数")
	}
	if v := n * 2; v > 10 {
		fmt.Println("v =", v, "> 10")
	}
}

func demoFor() {
	fmt.Println("--- for 循环 ---")
	sum := 0
	for i := 1; i <= 5; i++ {
		sum += i
	}
	fmt.Println("1..5 求和:", sum)

	i := 0
	for i < 3 {
		fmt.Print(i, " ")
		i++
	}
	fmt.Println()
}

func demoSwitch() {
	fmt.Println("--- switch ---")
	day := 3
	switch day {
	case 1:
		fmt.Println("周一")
	case 2, 3:
		fmt.Println("周二或周三")
	default:
		fmt.Println("其他")
	}

	switch {
	case day < 0:
		fmt.Println("无效")
	case day <= 5:
		fmt.Println("工作日")
	default:
		fmt.Println("周末")
	}
}

func demoRange() {
	fmt.Println("--- range ---")
	nums := []int{10, 20, 30}
	for i, v := range nums {
		fmt.Printf("  nums[%d]=%d\n", i, v)
	}
	for i, ch := range "Hi" {
		fmt.Printf("  [%d] %q\n", i, ch)
	}
}
