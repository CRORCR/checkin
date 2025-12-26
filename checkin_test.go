package checkin

import (
	"testing"
)

func TestCheckinRecord_Basic(t *testing.T) {
	// 创建新记录
	record := NewCheckinRecord()
	// 初始应该是0
	if record.ToInt64() != 0 {
		t.Errorf("新记录应该为0，实际为 %d", record.ToInt64())
	}

	// 今天打卡
	record = record.Checkin()
	if !record.IsCheckedToday() {
		t.Error("今天应该已打卡")
	}

	// 连续打卡天数应该是1
	if record.ContinuousDays() != 1 {
		t.Errorf("连续打卡天数应该是1，实际为 %d", record.ContinuousDays())
	}
}

func TestCheckinRecord_ContinuousDays(t *testing.T) {
	tests := []struct {
		name     string
		record   CheckinRecord
		expected int
	}{
		{
			name:     "无打卡",
			record:   0,
			expected: 0,
		},
		{
			name:     "连续1天",
			record:   0b1,
			expected: 1,
		},
		{
			name:     "连续3天",
			record:   0b111,
			expected: 3,
		},
		{
			name:     "连续7天",
			record:   0b1111111,
			expected: 7,
		},
		{
			name:     "中断：今天和昨天打卡，前天没打",
			record:   0b1011,
			expected: 2,
		},
		{
			name:     "中断：只有今天打卡",
			record:   0b1001,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.ContinuousDays()
			if got != tt.expected {
				t.Errorf("ContinuousDays() = %d, want %d (binary: %b)", got, tt.expected, tt.record)
			}
		})
	}
}

func TestCheckinRecord_TotalDaysInPeriod(t *testing.T) {
	// 0b1010101 = 最近7天打了4天卡（间隔打卡）
	record := CheckinRecord(0b1010101)

	total := record.TotalDaysInPeriod(7)
	if total != 4 {
		t.Errorf("最近7天应该打了4天卡，实际为 %d", total)
	}

	// 只统计最近3天
	total3 := record.TotalDaysInPeriod(3)
	if total3 != 2 {
		t.Errorf("最近3天应该打了2天卡，实际为 %d", total3)
	}
}

func TestCheckinRecord_ShiftDay(t *testing.T) {
	// 模拟：今天打卡 0b1
	record := CheckinRecord(0b1)

	// 时间推移到明天（今天的记录变成昨天的）
	record = record.ShiftDay()

	// 现在的值应该是 0b10
	if record != 0b10 {
		t.Errorf("ShiftDay后应该是 0b10，实际为 %b", record)
	}

	// 昨天应该有打卡记录
	if !record.IsCheckedDay(1) {
		t.Error("昨天应该有打卡记录")
	}

	// 今天没打卡
	if record.IsCheckedToday() {
		t.Error("今天还没打卡")
	}
}

func TestCheckinRecord_String(t *testing.T) {
	// 连续3天打卡
	record := CheckinRecord(0b111)
	str := record.StringWithDays(3)

	expected := "✓✓✓"
	if str != expected {
		t.Errorf("String应该是 %s，实际为 %s", expected, str)
	}

	// 间隔打卡
	record2 := CheckinRecord(0b101)
	str2 := record2.StringWithDays(3)
	expected2 := "✓✗✓" // 前天打卡，昨天没打，今天打卡

	if str2 != expected2 {
		t.Errorf("String应该是 %s，实际为 %s", expected2, str2)
	}
}

func TestCheckinRecord_MaxContinuousDays(t *testing.T) {
	tests := []struct {
		name     string
		record   CheckinRecord
		expected int
	}{
		{
			name:     "无打卡",
			record:   0,
			expected: 0,
		},
		{
			name:     "全部连续",
			record:   0b1111111,
			expected: 7,
		},
		{
			name:     "中间断了，最长3天",
			record:   0b1110011,
			expected: 3,
		},
		{
			name:     "分段连续，最长4天",
			record:   0b11110110,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.MaxContinuousDays()
			if got != tt.expected {
				t.Errorf("MaxContinuousDays() = %d, want %d (binary: %b)", got, tt.expected, tt.record)
			}
		})
	}
}

func TestCheckinRecord_CheckinRate(t *testing.T) {
	// 最近7天打了5天
	record := CheckinRecord(0b1011101)

	rate := record.CheckinRate(7)
	expected := 5.0 / 7.0

	if rate != expected {
		t.Errorf("CheckinRate(7) = %f, want %f", rate, expected)
	}
}

// 基准测试
func BenchmarkCheckinRecord_Checkin(b *testing.B) {
	record := NewCheckinRecord()
	for i := 0; i < b.N; i++ {
		record = record.Checkin()
	}
}

func BenchmarkCheckinRecord_ContinuousDays(b *testing.B) {
	record := CheckinRecord(0b1111111111111111)
	for i := 0; i < b.N; i++ {
		_ = record.ContinuousDays()
	}
}

func BenchmarkCheckinRecord_TotalDaysInPeriod(b *testing.B) {
	record := CheckinRecord(0b1010101010101010)
	for i := 0; i < b.N; i++ {
		_ = record.TotalDaysInPeriod(30)
	}
}
