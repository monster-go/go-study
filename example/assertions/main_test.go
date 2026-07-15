package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================
// testify 测试断言示例
// =============================================

func Add(a, b int) int {
	return a + b
}

func parseAge(s string) (int, error) {
	var age int
	if err := json.Unmarshal([]byte(s), &age); err != nil {
		return 0, err
	}
	return age, nil
}

// 基本断言
func TestAdd(t *testing.T) {
	got := Add(1, 2)

	assert.Equal(t, 3, got)    // 期望值在前，实际值在后
	assert.NotNil(t, got)
	assert.True(t, got > 0)
}

// 表格驱动测试
func TestAddTable(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 1, 2, 3},
		{"negative", -1, -2, -3},
		{"zero", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

// require vs assert：require 失败后立即中止
func TestParseAge(t *testing.T) {
	// require：前置条件必须成功
	age, err := parseAge("25")
	require.NoError(t, err) // 如果出错，这里就 Fatal 停止了
	assert.Equal(t, 25, age)

	// assert：错误存在即可
	_, err = parseAge("abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// 故意失败——验证 assert 的失败报告格式
func TestAddFailures_DemoOnly(t *testing.T) {
	t.Log("演示：assert.Equal 的失败报告")

	t.Run("wrong value", func(t *testing.T) {
		got := Add(1, 2)
		// 断言失败不会阻塞后续，但会标记测试失败
		assert.Equal(t, 99, got, "期望值 99，实际 %d", got)
	})
}

// JSON 断言
func TestGetStringJSON(t *testing.T) {
	data := map[string]any{
		"name": "Go",
		"ver":  1.21,
	}

	name, ok := data["name"].(string)
	assert.True(t, ok)
	assert.Equal(t, "Go", name)

	ver, ok := data["ver"].(float64)
	assert.True(t, ok)
	assert.Equal(t, 1.21, ver)
}
