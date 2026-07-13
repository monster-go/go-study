package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, error, wrap, interface, assert, niliface")
	flag.Parse()

	switch *mode {
	case "error":
		demoError()
	case "wrap":
		demoWrap()
	case "interface":
		demoInterface()
	case "assert":
		demoAssert()
	case "niliface":
		demoNilIface()
	default:
		demoError()
		demoWrap()
		demoInterface()
		demoAssert()
		demoNilIface()
	}
}

// --- 错误处理 ---

var ErrNotFound = errors.New("not found")

func FindUser(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("invalid id %d", id)
	}
	if id == 404 {
		return "", ErrNotFound
	}
	return fmt.Sprintf("user-%d", id), nil
}

func demoError() {
	fmt.Println("--- 错误处理 ---")
	name, err := FindUser(1)
	fmt.Printf("FindUser(1): name=%q err=%v\n", name, err)

	_, err = FindUser(404)
	fmt.Printf("FindUser(404): err=%v IsNotFound=%v\n", err, errors.Is(err, ErrNotFound))
}

// --- 错误包装 ---

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: field=%s msg=%s", e.Field, e.Msg)
}

func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse age %q: %w", s, err)
	}
	if n < 0 || n > 150 {
		return 0, &ValidationError{Field: "age", Msg: "out of range"}
	}
	return n, nil
}

func demoWrap() {
	fmt.Println("--- 错误包装 ---")
	_, err := ParseAge("abc")
	fmt.Printf("ParseAge(\"abc\"): %v\n", err)
	fmt.Printf("  errors.Is(ErrSyntax): %v\n", errors.Is(err, strconv.ErrSyntax))

	_, err = ParseAge("200")
	fmt.Printf("ParseAge(\"200\"): %v\n", err)
	var ve *ValidationError
	fmt.Printf("  errors.As ValidationError: %v field=%s\n", errors.As(err, &ve), ve.Field)
}

// --- 接口 ---

type Shape interface {
	Area() float64
}

type Circle struct {
	R float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.R * c.R
}

type Rect struct {
	W, H float64
}

func (r Rect) Area() float64 {
	return r.W * r.H
}

func TotalArea(shapes ...Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func demoInterface() {
	fmt.Println("--- 接口 ---")
	shapes := []Shape{Circle{R: 2}, Rect{W: 3, H: 4}}
	fmt.Printf("TotalArea: %.2f\n", TotalArea(shapes...))
}

// --- 类型断言与 type switch ---

func Describe(v any) string {
	switch x := v.(type) {
	case int:
		return fmt.Sprintf("int=%d", x)
	case string:
		return fmt.Sprintf("string=%q", x)
	case Shape:
		return fmt.Sprintf("Shape area=%.2f", x.Area())
	default:
		return fmt.Sprintf("%T", x)
	}
}

func demoAssert() {
	fmt.Println("--- 类型断言 ---")
	var s Shape = Circle{R: 1}
	if c, ok := s.(Circle); ok {
		fmt.Printf("assert Circle: R=%.1f\n", c.R)
	}
	fmt.Println(Describe(42))
	fmt.Println(Describe("go"))
	fmt.Println(Describe(Circle{R: 2}))
}

// --- nil 接口 ---

type Person struct {
	Name string
}

func demoNilIface() {
	fmt.Println("--- nil 接口 ---")
	var p *Person
	var i any = p
	fmt.Printf("p==nil: %v  i==nil: %v  i type=%T\n", p == nil, i == nil, i)
}
