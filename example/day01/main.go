package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	fmt.Println("=== 练习1：打印变量值和类型 ===")
	exercise1()

	fmt.Println("\n=== 练习2：统计汉字数量 ===")
	exercise2()

	fmt.Println("\n=== 练习3：99乘法表 ===")
	exercise3()
}

// 练习1：定义整型、浮点型、布尔型、字符串型变量，使用 %T 打印值和类型
func exercise1() {
	var i int = 42
	var f float64 = 3.14
	var b bool = true
	var s string = "Go"

	fmt.Printf("整型   值: %v, 类型: %T\n", i, i)
	fmt.Printf("浮点型 值: %v, 类型: %T\n", f, f)
	fmt.Printf("布尔型 值: %v, 类型: %T\n", b, b)
	fmt.Printf("字符串 值: %v, 类型: %T\n", s, s)
}

// 练习2：统计字符串 "hello沙河小王子" 中汉字数量
func exercise2() {
	s := "hello沙河小王子"
	count := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			count++
		}
	}
	fmt.Printf("字符串 %q 中汉字数量: %d\n", s, count)
}

// 练习3：打印 99 乘法表（下三角，带边框）
func exercise3() {
	const cellWidth = 7

	for row := 1; row <= 9; row++ {
		for col := 1; col <= row; col++ {
			fmt.Print("+", strings.Repeat("-", cellWidth))
		}
		fmt.Println("+")

		for col := 1; col <= row; col++ {
			text := fmt.Sprintf("%d×%d=%d", col, row, col*row)
			pad := cellWidth - len([]rune(text))
			if pad < 0 {
				pad = 0
			}
			fmt.Print("|", text, strings.Repeat(" ", pad))
		}
		fmt.Println("|")
	}

	for col := 1; col <= 9; col++ {
		fmt.Print("+", strings.Repeat("-", cellWidth))
	}
	fmt.Println("+")
}
