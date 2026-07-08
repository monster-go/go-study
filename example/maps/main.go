package main

import (
	"flag"
	"fmt"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, basic, ok, nil, range, delete")
	flag.Parse()

	switch *mode {
	case "basic":
		demoBasic()
	case "ok":
		demoOk()
	case "nil":
		demoNil()
	case "range":
		demoRange()
	case "delete":
		demoDelete()
	default:
		demoBasic()
		demoOk()
		demoNil()
		demoRange()
		demoDelete()
	}
}

func demoBasic() {
	fmt.Println("--- map 基本操作 ---")
	m := map[string]int{"a": 1, "b": 2}
	fmt.Println("字面量:", m)
	m2 := make(map[string]int)
	m2["x"] = 10
	fmt.Println("make + 赋值:", m2)
}

func demoOk() {
	fmt.Println("--- comma-ok ---")
	m := map[string]int{"go": 1}
	v, ok := m["go"]
	fmt.Println("存在 go:", v, ok)
	v2, ok2 := m["java"]
	fmt.Println("不存在 java:", v2, ok2)
}

func demoNil() {
	fmt.Println("--- nil map ---")
	var m map[string]int
	fmt.Println("nil map len:", len(m), "== nil:", m == nil)
	m2 := map[string]int{}
	fmt.Println("空 map {} len:", len(m2), "== nil:", m2 == nil)
	m3 := make(map[string]int)
	m3["k"] = 1
	fmt.Println("make 后可写:", m3)
}

func demoRange() {
	fmt.Println("--- range（顺序不固定）---")
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for k, v := range m {
		fmt.Printf("  %s -> %d\n", k, v)
	}
}

func demoDelete() {
	fmt.Println("--- delete ---")
	m := map[string]int{"a": 1, "b": 2}
	delete(m, "a")
	fmt.Println("delete a 后:", m, "len:", len(m))
	delete(m, "missing") // 删不存在的 key 不 panic
	fmt.Println("删不存在的 key 后:", m)
}
