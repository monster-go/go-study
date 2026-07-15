package main

import "fmt"

// StuManager 结构体版：数据存放在结构体字段中
type StuManager struct {
	students map[int]Student
}

func NewStuManager() *StuManager {
	return &StuManager{
		students: make(map[int]Student),
	}
}

func (m *StuManager) ShowAllStudents() {
	if len(m.students) == 0 {
		fmt.Println("当前没有学生信息")
		return
	}
	for _, s := range m.students {
		fmt.Printf("学号：%d 姓名：%s\n", s.Id, s.Name)
	}
}

func (m *StuManager) AddStudent() {
	var id int
	var name string
	fmt.Println("请输入学号：")
	fmt.Scanln(&id)
	fmt.Println("请输入姓名：")
	fmt.Scanln(&name)
	m.students[id] = Student{Id: id, Name: name}
	fmt.Println("添加学生信息成功")
}

func (m *StuManager) DeleteStudent() {
	var id int
	fmt.Println("请输入要删除的学生学号：")
	fmt.Scanln(&id)
	s, ok := m.students[id]
	if !ok {
		fmt.Println("删除失败，学号不存在")
		return
	}
	delete(m.students, id)
	fmt.Printf("学号:%d 姓名:%s 的学生已删除成功！\n", s.Id, s.Name)
}

func (m *StuManager) ModifyStudent() {
	var id int
	var name string
	fmt.Println("请输入要修改的学生学号：")
	fmt.Scanln(&id)

	s, ok := m.students[id]
	if !ok {
		fmt.Println("修改失败，学号不存在")
		return
	}

	fmt.Println("请输入新的姓名：")
	fmt.Scanln(&name)
	s.Name = name
	m.students[id] = s
	fmt.Printf("学号:%d 的学生已修改成功！\n", s.Id)
}

func (m *StuManager) QueryStudent() {
	var id int
	fmt.Println("请输入要查询的学生学号：")
	fmt.Scanln(&id)

	s, ok := m.students[id]
	if !ok {
		fmt.Println("查询失败，学号不存在")
		return
	}
	fmt.Printf("查询成功！学号:%d 姓名:%s\n", s.Id, s.Name)
}
