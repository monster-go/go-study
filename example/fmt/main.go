package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "all", "demo mode: all, print, format, scan, error")
	flag.Parse()

	switch *mode {
	case "print":
		demoPrint()
	case "format":
		demoFormat()
	case "scan":
		demoScan()
	case "error":
		demoError()
	default:
		demoPrint()
		demoFormat()
		demoError()
	}
}

func demoPrint() {
	fmt.Println("--- Print / Println / Printf ---")
	fmt.Print("Hello", " ", "Go")
	fmt.Println()
	fmt.Println("line", 1, true)
	fmt.Printf("name=%s age=%d\n", "Alice", 30)

	s := fmt.Sprintf("score=%.1f", 98.5)
	fmt.Println("Sprintf:", s)

	fmt.Fprintf(os.Stderr, "stderr: %s\n", "optional log")
}

func demoFormat() {
	fmt.Println("--- Format verbs ---")
	fmt.Printf("default %%v: %v %v %v\n", 42, 3.14, "hi")
	n := 42
	fmt.Printf("types: %d %f %q %t %p\n", 42, 3.14, "hi", true, &n)
	fmt.Printf("struct %%+v: %+v\n", person{Name: "Bob", Age: 25})
	fmt.Printf("struct %%#v: %#v\n", person{Name: "Bob", Age: 25})
	fmt.Printf("width/pad: %5d | %05d | %8.2f\n", 7, 7, 3.14159)
	fmt.Println("Stringer:", namedUser{Name: "Carol", Age: 28})
}

func demoScan() {
	fmt.Println("--- Scan family (see doc for interactive usage) ---")
	var name string
	var age int
	n, err := fmt.Sscan("Diana 22", &name, &age)
	fmt.Printf("Sscan n=%d name=%q age=%d err=%v\n", n, name, age, err)
}

func demoError() {
	fmt.Println("--- Errorf ---")
	err := fmt.Errorf("load config: %w", fmt.Errorf("file not found"))
	fmt.Println(err)
	fmt.Println("Unwrap:", fmt.Errorf("wrap: %w", err).Error())
}

type person struct {
	Name string
	Age  int
}

type namedUser struct {
	Name string
	Age  int
}

func (u namedUser) String() string {
	return fmt.Sprintf("%s(%d)", u.Name, u.Age)
}
