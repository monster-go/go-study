package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// 本地 MySQL 安装后按需修改以下配置；也可用环境变量覆盖（见 NewDBStuManager）
const (
	defaultMySQLUser = "root"
	defaultMySQLPass = "root"
	defaultMySQLHost = "127.0.0.1:3306"
	defaultMySQLDB   = "go_student"
)

// DBStuManager 数据库版：数据持久化到 MySQL
type DBStuManager struct {
	db *sql.DB
}

func NewDBStuManager() (*DBStuManager, error) {
	user := envOr("MYSQL_USER", defaultMySQLUser)
	pass := envOr("MYSQL_PASSWORD", defaultMySQLPass)
	host := envOr("MYSQL_HOST", defaultMySQLHost)
	name := envOr("MYSQL_DB", defaultMySQLDB)

	// 先连上实例（不指定库），自动建库
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=true", user, pass, host)
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return nil, fmt.Errorf("打开连接失败: %w", err)
	}
	defer rootDB.Close()

	if err := rootDB.Ping(); err != nil {
		return nil, fmt.Errorf("无法连接 MySQL（请确认已安装并启动，且账号密码正确）: %w", err)
	}

	createDB := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		name,
	)
	if _, err := rootDB.Exec(createDB); err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开业务库失败: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接业务库失败: %w", err)
	}

	const createTable = `
CREATE TABLE IF NOT EXISTS students (
	id INT NOT NULL PRIMARY KEY,
	name VARCHAR(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return &DBStuManager{db: db}, nil
}

func (m *DBStuManager) Close() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}

func (m *DBStuManager) ShowAllStudents() {
	rows, err := m.db.Query("SELECT id, name FROM students ORDER BY id")
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.Id, &s.Name); err != nil {
			fmt.Println("读取行失败:", err)
			return
		}
		fmt.Printf("学号：%d 姓名：%s\n", s.Id, s.Name)
		count++
	}
	if err := rows.Err(); err != nil {
		fmt.Println("遍历结果失败:", err)
		return
	}
	if count == 0 {
		fmt.Println("当前没有学生信息")
	}
}

func (m *DBStuManager) AddStudent() {
	var id int
	var name string
	fmt.Println("请输入学号：")
	fmt.Scanln(&id)
	fmt.Println("请输入姓名：")
	fmt.Scanln(&name)

	_, err := m.db.Exec("INSERT INTO students (id, name) VALUES (?, ?)", id, name)
	if err != nil {
		fmt.Println("添加失败:", err)
		return
	}
	fmt.Println("添加学生信息成功")
}

func (m *DBStuManager) DeleteStudent() {
	var id int
	fmt.Println("请输入要删除的学生学号：")
	fmt.Scanln(&id)

	var name string
	err := m.db.QueryRow("SELECT name FROM students WHERE id = ?", id).Scan(&name)
	if err == sql.ErrNoRows {
		fmt.Println("删除失败，学号不存在")
		return
	}
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}

	if _, err := m.db.Exec("DELETE FROM students WHERE id = ?", id); err != nil {
		fmt.Println("删除失败:", err)
		return
	}
	fmt.Printf("学号:%d 姓名:%s 的学生已删除成功！\n", id, name)
}

func (m *DBStuManager) ModifyStudent() {
	var id int
	var name string
	fmt.Println("请输入要修改的学生学号：")
	fmt.Scanln(&id)

	var exists int
	err := m.db.QueryRow("SELECT 1 FROM students WHERE id = ?", id).Scan(&exists)
	if err == sql.ErrNoRows {
		fmt.Println("修改失败，学号不存在")
		return
	}
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}

	fmt.Println("请输入新的姓名：")
	fmt.Scanln(&name)

	if _, err := m.db.Exec("UPDATE students SET name = ? WHERE id = ?", name, id); err != nil {
		fmt.Println("修改失败:", err)
		return
	}
	fmt.Printf("学号:%d 的学生已修改成功！\n", id)
}

func (m *DBStuManager) QueryStudent() {
	var id int
	fmt.Println("请输入要查询的学生学号：")
	fmt.Scanln(&id)

	var s Student
	err := m.db.QueryRow("SELECT id, name FROM students WHERE id = ?", id).Scan(&s.Id, &s.Name)
	if err == sql.ErrNoRows {
		fmt.Println("查询失败，学号不存在")
		return
	}
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Printf("查询成功！学号:%d 姓名:%s\n", s.Id, s.Name)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
