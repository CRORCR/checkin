package checkin

import "fmt"

// Example_Basic 基本使用示例
func Example_Basic() {
	// 创建新的打卡记录
	record := NewCheckinRecord()

	// 今天打卡
	record = record.Checkin()
	fmt.Printf("今天已打卡: %v\n", record.IsCheckedToday())
	fmt.Printf("连续打卡: %d天\n", record.ContinuousDays())
	fmt.Printf("最近7天记录: %s\n", record.StringWithDays(7))

	// 输出:
	// 今天已打卡: true
	// 连续打卡: 1天
	// 最近7天记录: ✗✗✗✗✗✗✓
}

// Example_Database 数据库存储示例
func Example_Database() {
	// 从数据库读取（假设字段类型为 bigint）
	var dbValue int64 = 0b111 // 连续3天打卡

	// 转换为打卡记录
	record := FromInt64(dbValue)

	// 今天又打卡了
	record = record.Checkin()

	// 保存回数据库
	newDBValue := record.ToInt64()
	fmt.Printf("保存到数据库: %d (二进制: %b)\n", newDBValue, newDBValue)

	// 输出:
	// 保存到数据库: 7 (二进制: 111)
}

// Example_ContinuousCheckin 连续打卡统计示例
func Example_ContinuousCheckin() {
	// 模拟7天的打卡记录
	// 0b1111101 表示：今天、昨天、前天、3天前、4天前、5天前打卡，6天前没打
	record := CheckinRecord(0b1111101)

	fmt.Printf("打卡记录: %s\n", record.StringWithDays(7))
	fmt.Printf("从今天开始连续打卡: %d天\n", record.ContinuousDays())
	fmt.Printf("历史最长连续打卡: %d天\n", record.MaxContinuousDays())
	fmt.Printf("最近7天打卡: %d天\n", record.TotalDaysInPeriod(7))
	fmt.Printf("最近7天打卡率: %.1f%%\n", record.CheckinRate(7)*100)

	// 输出:
	// 打卡记录: ✗✓✓✓✓✓✓
	// 从今天开始连续打卡: 5天
	// 历史最长连续打卡: 5天
	// 最近7天打卡: 6天
	// 最近7天打卡率: 85.7%
}

// Example_DayShift 每日时间推移示例
func Example_DayShift() {
	fmt.Println("=== 模拟3天打卡流程 ===")

	// 第1天
	record := NewCheckinRecord()
	record = record.Checkin()
	fmt.Printf("第1天打卡后: %s (值:%d)\n", record.String(), record.ToInt64())

	// 时间推移到第2天
	record = record.ShiftDay()
	fmt.Printf("第2天0点推移: %s (值:%d)\n", record.String(), record.ToInt64())

	// 第2天打卡
	record = record.Checkin()
	fmt.Printf("第2天打卡后: %s (值:%d)\n", record.String(), record.ToInt64())

	// 时间推移到第3天
	record = record.ShiftDay()
	record = record.Checkin() // 第3天打卡
	fmt.Printf("第3天打卡后: %s (值:%d)\n", record.String(), record.ToInt64())

	fmt.Printf("连续打卡: %d天\n", record.ContinuousDays())

	// 输出:
	// === 模拟3天打卡流程 ===
	// 第1天打卡后: ✗✗✗✗✗✗✓ (值:1)
	// 第2天0点推移: ✗✗✗✗✗✓✗ (值:2)
	// 第2天打卡后: ✗✗✗✗✗✓✓ (值:3)
	// 第3天打卡后: ✗✗✗✗✓✓✓ (值:7)
	// 连续打卡: 3天
}

// Example_BusinessLogic 实际业务逻辑示例
func Example_BusinessLogic() {
	fmt.Println("=== 用户打卡系统 ===")

	// 用户A: 连续打卡7天
	userA := CheckinRecord(0b1111111)
	fmt.Printf("用户A: %s - 连续%d天\n", userA.String(), userA.ContinuousDays())

	// 用户B: 断断续续打卡
	userB := CheckinRecord(0b1010101)
	fmt.Printf("用户B: %s - 连续%d天，总共%d天\n",
		userB.String(),
		userB.ContinuousDays(),
		userB.TotalDaysInPeriod(7))

	// 用户C: 前3天连续，后面断了
	userC := CheckinRecord(0b0000111)
	fmt.Printf("用户C: %s - 连续%d天，历史最长%d天\n",
		userC.String(),
		userC.ContinuousDays(),
		userC.MaxContinuousDays())

	// 判断是否可以领取连续打卡奖励（需要连续7天）
	if userA.ContinuousDays() >= 7 {
		fmt.Println("🎉 用户A可以领取7天连续打卡奖励！")
	}

	// 输出:
	// === 用户打卡系统 ===
	// 用户A: ✓✓✓✓✓✓✓ - 连续7天
	// 用户B: ✓✗✓✗✓✗✓ - 连续1天，总共4天
	// 用户C: ✗✗✗✗✓✓✓ - 连续3天，历史最长3天
	// 🎉 用户A可以领取7天连续打卡奖励！
}

// Example_GetBitmap 获取位图数组示例（用于前端展示）
func Example_GetBitmap() {
	record := CheckinRecord(0b1011101)

	// 获取最近7天的打卡位图
	bitmap := record.GetDaysBitmap(7)

	fmt.Println("最近7天打卡情况（数组形式）:")
	for i := len(bitmap) - 1; i >= 0; i-- {
		status := "✗"
		if bitmap[i] {
			status = "✓"
		}
		fmt.Printf("  %d天前: %s\n", i, status)
	}

	// 输出:
	// 最近7天打卡情况（数组形式）:
	//   0天前: ✓  (今天)
	//   1天前: ✗  (昨天)
	//   2天前: ✓
	//   3天前: ✓
	//   4天前: ✓
	//   5天前: ✗
	//   6天前: ✓
}
