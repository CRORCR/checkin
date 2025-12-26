package checkin

import (
	"fmt"
	"strings"
)

// CheckinRecord 打卡记录（使用位运算）
// 使用int64存储最近64天的打卡记录，每一位表示一天
// 第0位（最低位）表示今天，第1位表示昨天，以此类推
type CheckinRecord int64

// NewCheckinRecord 创建新的打卡记录
func NewCheckinRecord() CheckinRecord {
	return 0
}

// FromInt64 从int64创建打卡记录
func FromInt64(value int64) CheckinRecord {
	return CheckinRecord(value)
}

// ToInt64 转换为int64（用于存储到数据库）
func (c CheckinRecord) ToInt64() int64 {
	return int64(c)
}

// Checkin 打卡（设置今天的位为1）
func (c CheckinRecord) Checkin() CheckinRecord {
	return c | 1
}

// CheckinDay 打卡指定天（0=今天，1=昨天，2=前天...）
func (c CheckinRecord) CheckinDay(day int) CheckinRecord {
	if day < 0 || day >= 64 {
		return c
	}
	return c | (1 << day)
}

// IsCheckedToday 今天是否打卡
func (c CheckinRecord) IsCheckedToday() bool {
	return (c & 1) == 1
}

// IsCheckedDay 指定天是否打卡（0=今天，1=昨天，2=前天...）
func (c CheckinRecord) IsCheckedDay(day int) bool {
	if day < 0 || day >= 64 {
		return false
	}
	return (c & (1 << day)) != 0
}

// ContinuousDays 获取从今天开始的连续打卡天数
func (c CheckinRecord) ContinuousDays() int {
	count := 0
	for i := 0; i < 64; i++ {
		if (c & (1 << i)) != 0 {
			count++
		} else {
			break // 遇到第一个0就停止
		}
	}
	return count
}

// TotalDaysInPeriod 获取最近N天的总打卡天数
// 例如：GetTotalDays(7) 获取最近7天打卡了几天
func (c CheckinRecord) TotalDaysInPeriod(days int) int {
	if days <= 0 || days > 64 {
		days = 64
	}

	count := 0
	for i := 0; i < days; i++ {
		if (c & (1 << i)) != 0 {
			count++
		}
	}
	return count
}

// ShiftDay 时间推移（将记录整体左移一位，为新的一天准备）
// 调用时机：每天零点时调用，表示进入新的一天
func (c CheckinRecord) ShiftDay() CheckinRecord {
	return c << 1
}

// Clear 清空所有打卡记录
func (c CheckinRecord) Clear() CheckinRecord {
	return 0
}

// String 打卡记录的字符串表示（用于调试）
// 例如："✓✓✗✓✓✓✓" 表示最近7天的打卡情况
func (c CheckinRecord) String() string {
	return c.StringWithDays(7)
}

// StringWithDays 显示最近N天的打卡情况
func (c CheckinRecord) StringWithDays(days int) string {
	if days <= 0 || days > 64 {
		days = 64
	}

	var result strings.Builder
	for i := 0; i < days; i++ {
		if (c & (1 << i)) != 0 {
			result.WriteString("✓")
		} else {
			result.WriteString("✗")
		}
	}

	// 反转字符串（因为我们是从今天往前数的）
	str := result.String()
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Binary 返回二进制表示（用于调试）
func (c CheckinRecord) Binary() string {
	return fmt.Sprintf("0b%b", c)
}

// GetDaysBitmap 获取最近N天的位图数组（方便前端展示）
// 返回 []bool，索引0表示今天，索引1表示昨天
func (c CheckinRecord) GetDaysBitmap(days int) []bool {
	if days <= 0 || days > 64 {
		days = 64
	}

	result := make([]bool, days)
	for i := 0; i < days; i++ {
		result[i] = (c & (1 << i)) != 0
	}
	return result
}

// MaxContinuousDays 获取历史最大连续打卡天数
// 扫描所有位，找出最长的连续1
func (c CheckinRecord) MaxContinuousDays() int {
	if c == 0 {
		return 0
	}

	maxCount := 0
	currentCount := 0

	for i := 0; i < 64; i++ {
		if (c & (1 << i)) != 0 {
			currentCount++
			if currentCount > maxCount {
				maxCount = currentCount
			}
		} else {
			currentCount = 0
		}
	}

	return maxCount
}

// CheckinRate 获取最近N天的打卡率（0.0-1.0）
func (c CheckinRecord) CheckinRate(days int) float64 {
	if days <= 0 {
		return 0.0
	}
	total := c.TotalDaysInPeriod(days)
	return float64(total) / float64(days)
}
