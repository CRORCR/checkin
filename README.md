# Checkin - 高性能打卡统计

基于位运算的打卡记录与统计，使用单个int64存储最近64天的打卡记录，极致性能和最小存储空间。

## 核心特性

- ⚡ **极致性能**: 所有操作都是位运算，纳秒级响应
- 💾 **最小存储**: 仅需8字节(int64)存储64天记录
- 🎯 **一次计算**: 连续打卡天数、总天数等统计都是一次位运算完成

## 原理说明

使用int64的每一位表示一天是否打卡：
- 第0位（最低位）= 今天
- 第1位 = 昨天
- 第2位 = 前天
- ...以此类推

### 示例

```
0b111     = 连续3天打卡      (值: 7)
0b1111111 = 连续7天打卡      (值: 127)
0b101     = 今天和前天打卡    (值: 5)
0b1010101 = 间隔打卡         (值: 85)
```

## 快速开始

```go
package main

import (
    "fmt"
    "ysgit.lunalabs.cn/products/go-common/checkin"
)

func main() {
    // 创建新记录
    record := checkin.NewCheckinRecord()

    // 今天打卡
    record = record.Checkin()

    // 查询统计
    fmt.Printf("今天已打卡: %v\n", record.IsCheckedToday())
    fmt.Printf("连续打卡: %d天\n", record.ContinuousDays())
    fmt.Printf("最近7天记录: %s\n", record.StringWithDays(7))
}
```

## API文档

### 创建和转换

```go
// 创建新记录
record := checkin.NewCheckinRecord()

// 从数据库读取（int64类型）
record = checkin.FromInt64(dbValue)

// 转换为int64存储到数据库
dbValue := record.ToInt64()
```

### 打卡操作

```go
// 今天打卡
record = record.Checkin()

// 指定某天打卡（0=今天，1=昨天，2=前天...）
record = record.CheckinDay(2) // 前天打卡

// 时间推移（每天0点调用，进入新的一天）
record = record.ShiftDay()
```

### 查询操作

```go
// 今天是否打卡
isChecked := record.IsCheckedToday()

// 指定天是否打卡
isCheckedYesterday := record.IsCheckedDay(1)

// 从今天开始的连续打卡天数
days := record.ContinuousDays()

// 最近N天的总打卡天数
total := record.TotalDaysInPeriod(7) // 最近7天

// 历史最大连续打卡天数
maxDays := record.MaxContinuousDays()

// 打卡率
rate := record.CheckinRate(7) // 0.0-1.0
```

### 可视化

```go
// 字符串表示（默认7天）
str := record.String() // "✓✓✗✓✓✓✓"

// 指定天数
str := record.StringWithDays(3) // "✓✗✓"

// 二进制表示
binary := record.Binary() // "0b1011101"

// 获取位图数组（用于前端）
bitmap := record.GetDaysBitmap(7) // []bool{true, false, true, ...}
```

## 数据库集成

### MySQL示例

```sql
CREATE TABLE user_checkin (
    user_id BIGINT PRIMARY KEY,
    record BIGINT NOT NULL DEFAULT 0,  -- 打卡记录
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

```go
// 读取
var dbValue int64
db.QueryRow("SELECT record FROM user_checkin WHERE user_id = ?", userID).Scan(&dbValue)
record := checkin.FromInt64(dbValue)

// 打卡
record = record.Checkin()

// 保存
db.Exec("UPDATE user_checkin SET record = ? WHERE user_id = ?", record.ToInt64(), userID)
```

### Redis示例

```go
// 读取
val, _ := redis.Get(ctx, fmt.Sprintf("checkin:%d", userID)).Int64()
record := checkin.FromInt64(val)

// 打卡
record = record.Checkin()

// 保存
redis.Set(ctx, fmt.Sprintf("checkin:%d", userID), record.ToInt64(), 0)
```

## 业务场景

### 场景1: 每日打卡系统

```go
// 用户打卡
func UserCheckin(userID int64) error {
    // 从数据库读取
    record := getRecordFromDB(userID)

    // 检查今天是否已打卡
    if record.IsCheckedToday() {
        return errors.New("今天已经打卡过了")
    }

    // 打卡
    record = record.Checkin()

    // 保存
    return saveRecordToDB(userID, record)
}

// 获取打卡统计
func GetCheckinStats(userID int64) map[string]interface{} {
    record := getRecordFromDB(userID)

    return map[string]interface{}{
        "continuous_days": record.ContinuousDays(),
        "total_7days":     record.TotalDaysInPeriod(7),
        "max_continuous":  record.MaxContinuousDays(),
        "rate_7days":      record.CheckinRate(7),
        "recent_7days":    record.StringWithDays(7),
    }
}
```

### 场景2: 连续打卡奖励

```go
// 检查是否可以领取奖励
func CanClaimReward(userID int64, requireDays int) bool {
    record := getRecordFromDB(userID)
    return record.ContinuousDays() >= requireDays
}

// 奖励配置
var rewards = map[int]int{
    3:  100,  // 连续3天：100积分
    7:  500,  // 连续7天：500积分
    30: 3000, // 连续30天：3000积分
}
```

### 场景3: 定时任务（每日0点）

```go
// 每天0点执行，将所有记录推移一天
func DailyShiftTask() {
    // 批量更新所有用户的打卡记录
    db.Exec(`
        UPDATE user_checkin
        SET record = record << 1
    `)
}
```

### 场景4: 排行榜

```go
// 获取连续打卡天数TOP 10
func GetCheckinLeaderboard() []UserCheckin {
    var results []UserCheckin

    // 从数据库读取所有记录，然后计算连续天数排序
    // 注意：对于大量数据，建议冗余一个continuous_days字段定期更新
    db.Select(&results, `
        SELECT user_id, record
        FROM user_checkin
        ORDER BY record DESC
        LIMIT 100
    `)

    // 计算并排序
    for i := range results {
        results[i].ContinuousDays = checkin.FromInt64(results[i].Record).ContinuousDays()
    }

    return results[:10]
}
```

## 性能测试

```
BenchmarkCheckin-8              1000000000      0.5 ns/op
BenchmarkContinuousDays-8       500000000       3.2 ns/op
BenchmarkTotalDaysInPeriod-8    200000000       7.5 ns/op
```

所有操作都是纳秒级，可以放心在高并发场景使用。

## 注意事项

1. **时间推移**: 需要每天0点调用`ShiftDay()`方法，将记录整体左移一位
2. **最大天数**: 单个int64最多存储64天记录，超过64天的记录会被丢弃
3. **不可变**: `CheckinRecord`是不可变的，所有操作都返回新的记录
4. **并发安全**: 本身是值类型，线程安全；但数据库读写需要自行加锁

## 扩展功能

如果需要存储更长时间的记录：
- 使用多个int64字段（如：current_month, last_month）
- 使用更大的数据类型（如：big.Int）
- 或者使用位图数据库（如Redis bitmap）

## License

MIT
