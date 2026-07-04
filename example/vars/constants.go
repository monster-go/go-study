package main

// Weekday 用 iota 做连续枚举，新 const 块会从 0 重新计数。
// const (
// 	Sunday = iota
// 	Monday
// 	Tuesday
// 	Wednesday
// 	Thursday
// 	Friday
// 	Saturday
// )

const (
	Sunday = 0
	Monday = 1
	Tuesday = 2
	Wednesday = 3
	Thursday = 4
	Friday = 5
	Saturday = 6
)


// 用 _ 跳过 0，并用位运算生成 2 的幂。
const (
	_        = iota
	ReadPerm = 1 << iota // 2
	WritePerm            // 4
	ExecPerm             // 8
)
