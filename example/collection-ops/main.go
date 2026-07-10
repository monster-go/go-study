package main

import (
	"cmp"
	"flag"
	"fmt"
	"maps"
	"slices"
	"sort"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, sort, delete, search, equal, clone, clear, cmp")
	flag.Parse()

	switch *mode {
	case "sort":
		demoSort()
	case "delete":
		demoDelete()
	case "search":
		demoSearch()
	case "equal":
		demoEqual()
	case "clone":
		demoClone()
	case "clear":
		demoClear()
	case "cmp":
		demoCmp()
	default:
		demoSort()
		demoDelete()
		demoSearch()
		demoEqual()
		demoClone()
		demoClear()
		demoCmp()
	}
}

func demoSort() {
	fmt.Println("--- 排序：sort.Strings vs slices.Sort ---")
	words := []string{"go", "rust", "c", "java"}
	legacy := slices.Clone(words)
	slices.Sort(legacy)
	fmt.Println("slices.Sort:", legacy)

	words2 := []string{"go", "rust", "c", "java"}
	sort.Strings(words2)
	fmt.Println("sort.Strings:", words2)

	type item struct {
		name  string
		score int
	}
	items := []item{{"bob", 80}, {"alice", 95}, {"carol", 80}}
	slices.SortFunc(items, func(a, b item) int {
		if c := cmp.Compare(b.score, a.score); c != 0 {
			return c
		}
		return cmp.Compare(a.name, b.name)
	})
	fmt.Println("SortFunc(分数降序, 姓名升序):", items)
}

func demoDelete() {
	fmt.Println("--- 删除 slice 元素 ---")
	s := []int{10, 20, 30, 40, 50}

	byAppend := slices.Clone(s)
	byAppend = append(byAppend[:2], byAppend[3:]...)
	fmt.Println("append 拼接删下标 2:", byAppend)

	bySlices := slices.Delete(slices.Clone(s), 2, 3)
	fmt.Println("slices.Delete(2,3):", bySlices)

	byFunc := slices.DeleteFunc(slices.Clone(s), func(v int) bool { return v%20 == 0 })
	fmt.Println("DeleteFunc(删 20 的倍数):", byFunc)

	fast := slices.Clone(s)
	i := 2
	fast[i] = fast[len(fast)-1]
	fast = fast[:len(fast)-1]
	fmt.Println("swap+截断(不保持顺序):", fast)
}

func demoSearch() {
	fmt.Println("--- 查找 ---")
	tags := []string{"go", "web", "api"}
	fmt.Println("Contains go:", slices.Contains(tags, "go"))
	fmt.Println("Index api:", slices.Index(tags, "api"))

	sorted := []int{1, 3, 5, 7, 9}
	idx, found := slices.BinarySearch(sorted, 5)
	fmt.Println("BinarySearch 5:", idx, "found:", found)
}

func demoEqual() {
	fmt.Println("--- 相等比较 ---")
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	fmt.Println("slices.Equal(a,b):", slices.Equal(a, b))
	fmt.Println("a==nil:", a == nil, "a 只能与 nil 用 ==")

	m1 := map[string]int{"x": 1}
	m2 := map[string]int{"x": 1}
	fmt.Println("maps.Equal(m1,m2):", maps.Equal(m1, m2))
}

func demoClone() {
	fmt.Println("--- 克隆 / 合并 / 插入 ---")
	src := []int{1, 2, 3}
	dup := slices.Clone(src)
	dup[0] = 99
	fmt.Println("原切片:", src, "克隆后改 dup[0]:", dup)

	all := slices.Concat([]int{1}, []int{2, 3}, []int{4})
	fmt.Println("Concat:", all)

	s := []int{1, 2, 3}
	s = slices.Insert(s, 1, 99)
	fmt.Println("Insert 下标 1:", s)
}

func demoClear() {
	fmt.Println("--- 清空 ---")
	m := map[string]int{"a": 1, "b": 2}
	clear(m)
	fmt.Println("clear(map) 后:", m, "len:", len(m))

	s := make([]int, 0, 8)
	s = append(s, 1, 2, 3)
	capBefore := cap(s)
	s = s[:0]
	fmt.Println("s[:0] 后 len:", len(s), "cap 保留:", cap(s), "== capBefore:", cap(s) == capBefore)
}

func demoCmp() {
	fmt.Println("--- cmp.Compare ---")
	fmt.Println("Compare(3,5):", cmp.Compare(3, 5))
	fmt.Println("Compare(5,5):", cmp.Compare(5, 5))
	fmt.Println("Compare(7,5):", cmp.Compare(7, 5))
}
