package main

import (
	"encoding/json"
	"fmt"
)

// =============================================
// 类型断言（Type Assertion）示例
// =============================================

func demoTypeAssertion() {
	fmt.Println("--- 类型断言 ---")

	var v any = "hello"

	// 安全形式
	s, ok := v.(string)
	fmt.Printf("safe: %q ok=%t\n", s, ok)

	// 不安全形式（故意用 defer recover 兜底）
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("unsafe(nil check): ok=false\n")
			}
		}()
		_ = v.(int) // panic: interface conversion
	}()

	// JSON 场景：map[string]any 取值
	data := map[string]any{
		"name":    "Go",
		"version": 1.21,
	}
	name, _ := data["name"].(string)
	ver, _ := data["version"].(float64)
	fmt.Printf("name=%s, version=%.2f\n", name, ver)
}

// =============================================
// type switch 示例
// =============================================

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int=%d", x)
	case string:
		return fmt.Sprintf("string=%q", x)
	case bool:
		return fmt.Sprintf("bool(%t)", x)
	default:
		return fmt.Sprintf("unknown(%T)", x)
	}
}

func demoTypeSwitch() {
	fmt.Println("\n--- type switch ---")

	items := []any{42, "go", true, nil, 3.14}
	for _, item := range items {
		fmt.Println(describe(item))
	}
}

// =============================================
// JSON 反序列化 + 类型断言
// =============================================

func demoJSONUnmarshal() {
	fmt.Println("\n--- JSON 解包 ---")

	jsonStr := `{"name":"go","version":1.21}`
	var raw any
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		fmt.Println("json error:", err)
		return
	}

	m, ok := raw.(map[string]any)
	if !ok {
		fmt.Println("not a map")
		return
	}

	name, _ := m["name"].(string)
	ver, _ := m["version"].(float64) // JSON 数字默认是 float64
	fmt.Printf("name=%s, version=%.2f\n", name, ver)
}

func main() {
	demoTypeAssertion()
	demoTypeSwitch()
	demoJSONUnmarshal()
}
