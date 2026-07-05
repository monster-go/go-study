package main

import (
	"flag"
	"fmt"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, array, slice, append, share")
	flag.Parse()

	switch *mode {
	case "array":
		demoArray()
	case "slice":
		demoSlice()
	case "append":
		demoAppend()
	case "share":
		demoShare()
	default:
		demoArray()
		demoSlice()
		demoAppend()
		demoShare()
	}
}

func demoArray() {
	fmt.Println("--- 数组 ---")
	var a [3]int
	a[0] = 1
	fmt.Println("a:", a, "len:", len(a))
	b := [3]int{1, 2, 3}
	fmt.Println("b:", b)
}

func demoSlice() {
	fmt.Println("--- 切片 ---")
	s := []int{1, 2, 3, 4, 5}
	fmt.Println("s:", s, "len:", len(s), "cap:", cap(s))
	sub := s[1:4]
	fmt.Println("s[1:4]:", sub, "len:", len(sub), "cap:", cap(sub))
	empty := make([]int, 0, 5)
	fmt.Println("make([]int,0,5):", empty, "len:", len(empty), "cap:", cap(empty))
}

func demoAppend() {
	fmt.Println("--- append / copy ---")
	s := []int{1, 2}
	s = append(s, 3, 4)
	fmt.Println("append 后:", s, "cap:", cap(s))
	dst := make([]int, len(s))
	n := copy(dst, s)
	fmt.Println("copy ->", dst, "复制个数:", n)
}

func demoShare() {
	fmt.Println("--- 底层数组共享 ---")
	orig := []int{1, 2, 3, 4}
	alias := orig[1:3]
	alias[0] = 99
	fmt.Println("改 alias 后 orig:", orig)
}
