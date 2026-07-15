package main

import (
	"fmt"
	"os"
)

// StudentSystem 统一接口：函数版、结构体版都实现它，实现多态切换
type StudentSystem interface {
	ShowAllStudents()
	AddStudent()
	DeleteStudent()
	ModifyStudent()
	QueryStudent()
}
type Student struct {
	Id   int
	Name string
}

func ShowMenus() {
	m := `
------------ welcome to the system ----------------
----------------------------------------------------
		1. 展示所有学生信息
		2. 添加学生信息
		3. 删除学生信息
		4. 修改学生信息
		5. 查询学生信息
		6. 退出系统
----------------------------------------------------
		please select:
	`
	fmt.Println(m)
}

func chooseSystem() StudentSystem {
	fmt.Println("请选择实现方式：")
	fmt.Println("  1. 函数版")
	fmt.Println("  2. 结构体版")
	fmt.Print("请输入：")

	var mode int
	fmt.Scanln(&mode)

	switch mode {
	case 1:
		fmt.Println("已选择：函数版")
		return NewFuncStuManager()
	default:
		fmt.Println("已选择：结构体版")
		return NewStuManager()
	}
}

func run(sys StudentSystem) {
	var choice int
	for {
		ShowMenus()
		fmt.Scanln(&choice)
		switch choice {
		case 1:
			sys.ShowAllStudents()
		case 2:
			sys.AddStudent()
		case 3:
			sys.DeleteStudent()
		case 4:
			sys.ModifyStudent()
		case 5:
			sys.QueryStudent()
		case 6:
			fallthrough
		default:
			fmt.Println("正在退出...bye!")
			os.Exit(0)
		}
	}
}

func main() {
	sys := chooseSystem()
	run(sys)
}
