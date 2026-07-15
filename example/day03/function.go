package main

import "fmt"

// 函数版：用包级 map + 普通函数实现业务
var funcStudents = make(map[int]Student)

func showAllStudents() {
	if len(funcStudents) == 0 {
		fmt.Println("当前没有学生信息")
		return
	}
	for _, s := range funcStudents {
		fmt.Printf("学号：%d 姓名：%s\n", s.Id, s.Name)
	}
}

func addStudent() {
	var id int
	var name string
	fmt.Println("请输入学号：")
	fmt.Scanln(&id)
	fmt.Println("请输入姓名：")
	fmt.Scanln(&name)
	funcStudents[id] = Student{Id: id, Name: name}
	fmt.Println("添加学生信息成功")
}

func deleteStudent() {
	var id int
	fmt.Println("请输入要删除的学生学号：")
	fmt.Scanln(&id)
	s, ok := funcStudents[id]
	if !ok {
		fmt.Println("删除失败，学号不存在")
		return
	}
	delete(funcStudents, id)
	fmt.Printf("学号:%d 姓名:%s 的学生已删除成功！\n", s.Id, s.Name)
}

func modifyStudent() {
	var id int
	var name string
	fmt.Println("请输入要修改的学生学号：")
	fmt.Scanln(&id)

	s, ok := funcStudents[id]
	if !ok {
		fmt.Println("修改失败，学号不存在")
		return
	}

	fmt.Println("请输入新的姓名：")
	fmt.Scanln(&name)
	s.Name = name
	funcStudents[id] = s
	fmt.Printf("学号:%d 的学生已修改成功！\n", s.Id)
}

func queryStudent() {
	var id int
	fmt.Println("请输入要查询的学生学号：")
	fmt.Scanln(&id)

	s, ok := funcStudents[id]
	if !ok {
		fmt.Println("查询失败，学号不存在")
		return
	}
	fmt.Printf("查询成功！学号:%d 姓名:%s\n", s.Id, s.Name)
}

// FuncStuManager 把函数版包装成接口实现，供多态切换
type FuncStuManager struct{}

func NewFuncStuManager() *FuncStuManager {
	return &FuncStuManager{}
}

func (f *FuncStuManager) ShowAllStudents() { showAllStudents() }
func (f *FuncStuManager) AddStudent()      { addStudent() }
func (f *FuncStuManager) DeleteStudent()   { deleteStudent() }
func (f *FuncStuManager) ModifyStudent()   { modifyStudent() }
func (f *FuncStuManager) QueryStudent()    { queryStudent() }
