package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// 本地 MySQL 配置——默认密码为空占位，部署时通过环境变量覆盖
const (
	defaultMySQLUser = "root"
	defaultMySQLPass = "" // 请通过 MYSQL_PASSWORD 环境变量设置
	defaultMySQLHost = "127.0.0.1:3306"
	defaultMySQLDB   = "sys_students"
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

	// 使用 mysql.Config 构建 DSN，避免密码含特殊字符时解析错误
	cfg := mysql.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  "tcp",
		Addr:                 host,
		DBName:               name,
		ParseTime:            true,
		AllowNativePasswords: true,
	}
	// 字符集通过 Params 设置
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}
	dsn := cfg.FormatDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开业务库失败: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)             // 最大打开连接数
	db.SetMaxIdleConns(5)              // 最大空闲连接数
	db.SetConnMaxLifetime(5 * time.Minute) // 连接最长存活时间

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接业务库失败: %w", err)
	}

	// 自动初始化表结构
	if err := initTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return &DBStuManager{db: db}, nil
}

// initTable 自动创建库和表（幂等——IF NOT EXISTS）
func initTable(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先建库
	if _, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS sys_students DEFAULT CHARACTER SET utf8mb4"); err != nil {
		return fmt.Errorf("建库失败: %w", err)
	}

	// 切换库——注意：db 的 DSN 已经指定了库名，这里只是确保
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS students (
		id   INT PRIMARY KEY,
		name VARCHAR(128) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return fmt.Errorf("建表失败: %w", err)
	}

	return nil
}

func (m *DBStuManager) Close() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}

// ============================================================
// 纯数据操作层（返回 error，不处理 I/O，可单元测试）
// ============================================================

func (m *DBStuManager) showAllStudentsData(ctx context.Context) ([]Student, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id, name FROM students ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.Id, &s.Name); err != nil {
			return nil, fmt.Errorf("读取行失败: %w", err)
		}
		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果失败: %w", err)
	}
	return students, nil
}

func (m *DBStuManager) addStudentData(ctx context.Context, id int, name string) error {
	_, err := m.db.ExecContext(ctx, "INSERT INTO students (id, name) VALUES (?, ?)", id, name)
	return err
}

// deleteStudentData 删除并返回删除前的姓名，方便调用方展示
func (m *DBStuManager) deleteStudentData(ctx context.Context, id int) (string, error) {
	// 在一个事务中：先查后删（原子操作，避免竞态）
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback() // 若 Commit 成功则 noop

	var name string
	err = tx.QueryRowContext(ctx, "SELECT name FROM students WHERE id = ? FOR UPDATE", id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("学号不存在")
	}
	if err != nil {
		return "", fmt.Errorf("查询失败: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM students WHERE id = ?", id); err != nil {
		return "", fmt.Errorf("删除失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("提交事务失败: %w", err)
	}
	return name, nil
}

func (m *DBStuManager) modifyStudentData(ctx context.Context, id int, name string) error {
	res, err := m.db.ExecContext(ctx, "UPDATE students SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return fmt.Errorf("修改失败: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("学号不存在")
	}
	return nil
}

func (m *DBStuManager) queryStudentData(ctx context.Context, id int) (Student, error) {
	var s Student
	err := m.db.QueryRowContext(ctx, "SELECT id, name FROM students WHERE id = ?", id).Scan(&s.Id, &s.Name)
	if err == sql.ErrNoRows {
		return s, fmt.Errorf("学号不存在")
	}
	if err != nil {
		return s, fmt.Errorf("查询失败: %w", err)
	}
	return s, nil
}

// ============================================================
// 用户交互层（控制台 I/O，调用纯数据层）
// ============================================================

func (m *DBStuManager) ShowAllStudents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	students, err := m.showAllStudentsData(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(students) == 0 {
		fmt.Println("当前没有学生信息")
		return
	}
	for _, s := range students {
		fmt.Printf("学号：%d 姓名：%s\n", s.Id, s.Name)
	}
}

// readInt 从标准输入读取一个整数，带重试
func readInt(prompt string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < 3; i++ {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("读取输入失败: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Println("输入不能为空，请重新输入")
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("请输入有效数字")
			continue
		}
		if n <= 0 {
			fmt.Println("学号必须为正整数")
			continue
		}
		return n, nil
	}
	return 0, fmt.Errorf("输入错误次数过多")
}

// readString 从标准输入读取一行字符串，带重试
func readString(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < 3; i++ {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("读取输入失败: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Println("姓名不能为空，请重新输入")
			continue
		}
		return line, nil
	}
	return "", fmt.Errorf("输入错误次数过多")
}

func (m *DBStuManager) AddStudent() {
	id, err := readInt("请输入学号：")
	if err != nil {
		fmt.Println(err)
		return
	}
	name, err := readString("请输入姓名：")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.addStudentData(ctx, id, name); err != nil {
		fmt.Println("添加失败:", err)
		return
	}
	fmt.Println("添加学生信息成功")
}

func (m *DBStuManager) DeleteStudent() {
	id, err := readInt("请输入要删除的学生学号：")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name, err := m.deleteStudentData(ctx, id)
	if err != nil {
		fmt.Println("删除失败:", err)
		return
	}
	fmt.Printf("学号:%d 姓名:%s 的学生已删除成功！\n", id, name)
}

func (m *DBStuManager) ModifyStudent() {
	id, err := readInt("请输入要修改的学生学号：")
	if err != nil {
		fmt.Println(err)
		return
	}
	name, err := readString("请输入新的姓名：")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.modifyStudentData(ctx, id, name); err != nil {
		fmt.Println("修改失败:", err)
		return
	}
	fmt.Printf("学号:%d 的学生已修改成功！\n", id)
}

func (m *DBStuManager) QueryStudent() {
	id, err := readInt("请输入要查询的学生学号：")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := m.queryStudentData(ctx, id)
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
