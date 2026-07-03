package main

// Weekday 用 iota 做连续枚举，新 const 块会从 0 重新计数。
const (
	Sunday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// 用 _ 跳过 0，并用位运算生成 2 的幂。
const (
	_        = iota
	ReadPerm = 1 << iota // 2
	WritePerm            // 4
	ExecPerm             // 8
)
