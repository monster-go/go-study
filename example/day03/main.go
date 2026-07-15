package main

import (
	"fmt"
	"os"
)

// StudentSystem 统一接口：函数版、结构体版、数据库版都实现它，实现多态切换
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
	fmt.Println("  3. 数据库版 (MySQL)")
	fmt.Print("请输入：")

	var mode int
	fmt.Scanln(&mode)

	switch mode {
	case 1:
		fmt.Println("已选择：函数版")
		return NewFuncStuManager()
	case 3:
		fmt.Println("已选择：数据库版")
		m, err := NewDBStuManager()
		if err != nil {
			fmt.Println(err)
			fmt.Println("提示：安装并启动 MySQL 后，检查 db.go 中的账号密码；也可用环境变量 MYSQL_USER / MYSQL_PASSWORD / MYSQL_HOST / MYSQL_DB")
			os.Exit(1)
		}
		return m
	default:
		fmt.Println("已选择：结构体版")
		return NewStuManager()
	}
}

func run(sys StudentSystem) {
	if closer, ok := sys.(interface{ Close() error }); ok {
		defer closer.Close()
	}

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
			return
		}
	}
}

func main() {
	sys := chooseSystem()
	run(sys)
}
