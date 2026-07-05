package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, concat, strings, strconv, rune")
	flag.Parse()

	switch *mode {
	case "concat":
		demoConcat()
	case "strings":
		demoStringsPkg()
	case "strconv":
		demoStrconv()
	case "rune":
		demoRune()
	default:
		demoConcat()
		demoStringsPkg()
		demoStrconv()
		demoRune()
	}
}

func demoConcat() {
	fmt.Println("--- 拼接 ---")
	a, b := "hello", "go"
	fmt.Println(a + " " + b)
	var builder strings.Builder
	builder.WriteString(a)
	builder.WriteByte(' ')
	builder.WriteString(b)
	fmt.Println("Builder:", builder.String())
}

func demoStringsPkg() {
	fmt.Println("--- strings 包 ---")
	s := "  hello,go,world  "
	fmt.Println("Trim:", strings.TrimSpace(s))
	fmt.Println("Split:", strings.Split(strings.TrimSpace(s), ","))
	fmt.Println("Contains:", strings.Contains("golang", "go"))
	fmt.Println("Join:", strings.Join([]string{"a", "b", "c"}, "-"))
	fmt.Println("Replace:", strings.ReplaceAll("foo-bar-foo", "foo", "baz"))
}

func demoStrconv() {
	fmt.Println("--- strconv 转换 ---")
	n, err := strconv.Atoi("42")
	fmt.Println("Atoi:", n, err)
	fmt.Println("Itoa:", strconv.Itoa(100))
	f, _ := strconv.ParseFloat("3.14", 64)
	fmt.Println("ParseFloat:", f)
	fmt.Println("FormatFloat:", strconv.FormatFloat(f, 'f', 2, 64))
}

func demoRune() {
	fmt.Println("--- 遍历 rune ---")
	s := "Go语言"
	fmt.Println("len(s) 字节数:", len(s))
	fmt.Println("rune 数:", len([]rune(s)))
	for i, r := range s {
		fmt.Printf("  [%d] %q Han=%v\n", i, r, unicode.Is(unicode.Han, r))
	}
}
