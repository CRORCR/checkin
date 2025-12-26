package checkin

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ==================== MySQL 示例 ====================
//⏺ 每天0点必须执行这个 SQL：
//UPDATE user_checkin SET record = record << 1
//这条 SQL 会把所有用户的记录整体左移一位，让"今天"变成"昨天"。

// UserCheckinModel 用户打卡数据模型
type UserCheckinModel struct {
	UserID    int64     `db:"user_id"`
	Record    int64     `db:"record"` // 打卡记录（BIGINT类型）
	UpdatedAt time.Time `db:"updated_at"`
}

// CheckinService 打卡服务
type CheckinService struct {
	db *sql.DB
}

// NewCheckinService 创建打卡服务
func NewCheckinService(db *sql.DB) *CheckinService {
	return &CheckinService{db: db}
}

// UserCheckin 用户打卡
func (s *CheckinService) UserCheckin(ctx context.Context, userID int64) error {
	// 1. 从数据库读取记录
	var dbValue int64
	err := s.db.QueryRowContext(ctx,
		"SELECT record FROM user_checkin WHERE user_id = ?",
		userID,
	).Scan(&dbValue)

	if err == sql.ErrNoRows {
		// 首次打卡，创建记录
		dbValue = 0
	} else if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}

	// 2. 转换为 CheckinRecord 类型
	record := FromInt64(dbValue)

	// 3. 检查今天是否已打卡
	if record.IsCheckedToday() {
		return fmt.Errorf("今天已经打卡过了")
	}

	// 4. 打卡
	record = record.Checkin()

	// 5. 保存回数据库
	newValue := record.ToInt64()

	if dbValue == 0 {
		// 插入新记录
		_, err = s.db.ExecContext(ctx,
			"INSERT INTO user_checkin (user_id, record) VALUES (?, ?)",
			userID, newValue,
		)
	} else {
		// 更新现有记录
		_, err = s.db.ExecContext(ctx,
			"UPDATE user_checkin SET record = ?, updated_at = NOW() WHERE user_id = ?",
			newValue, userID,
		)
	}

	return err
}

// GetUserCheckinStats 获取用户打卡统计
func (s *CheckinService) GetUserCheckinStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	// 从数据库读取
	var dbValue int64
	err := s.db.QueryRowContext(ctx,
		"SELECT record FROM user_checkin WHERE user_id = ?",
		userID,
	).Scan(&dbValue)

	if err == sql.ErrNoRows {
		// 没有记录
		return map[string]interface{}{
			"continuous_days": 0,
			"total_7days":     0,
			"total_30days":    0,
			"max_continuous":  0,
			"rate_7days":      0.0,
			"recent_7days":    "✗✗✗✗✗✗✗",
		}, nil
	}

	if err != nil {
		return nil, err
	}

	// 转换并计算统计
	record := FromInt64(dbValue)

	return map[string]interface{}{
		"continuous_days": record.ContinuousDays(),
		"total_7days":     record.TotalDaysInPeriod(7),
		"total_30days":    record.TotalDaysInPeriod(30),
		"max_continuous":  record.MaxContinuousDays(),
		"rate_7days":      record.CheckinRate(7),
		"recent_7days":    record.StringWithDays(7),
		"bitmap_7days":    record.GetDaysBitmap(7),
	}, nil
}

// IsCheckedToday 检查今天是否已打卡
func (s *CheckinService) IsCheckedToday(ctx context.Context, userID int64) (bool, error) {
	var dbValue int64
	err := s.db.QueryRowContext(ctx,
		"SELECT record FROM user_checkin WHERE user_id = ?",
		userID,
	).Scan(&dbValue)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	record := FromInt64(dbValue)
	return record.IsCheckedToday(), nil
}

// DailyShiftTask 定时任务：每天0点执行，推移所有记录
func (s *CheckinService) DailyShiftTask(ctx context.Context) error {
	// 方式1：直接SQL操作（最高效）
	// 将所有 record 左移1位（相当于 record = record << 1）
	_, err := s.db.ExecContext(ctx,
		"UPDATE user_checkin SET record = record << 1",
	)
	return err

	// 方式2：逐条处理（如果需要额外逻辑）
	// rows, err := s.db.QueryContext(ctx, "SELECT user_id, record FROM user_checkin")
	// ... 循环处理每条记录
}

// GetCheckinLeaderboard 获取连续打卡排行榜 TOP N
func (s *CheckinService) GetCheckinLeaderboard(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// 从数据库读取记录
	rows, err := s.db.QueryContext(ctx,
		"SELECT user_id, record FROM user_checkin ORDER BY record DESC LIMIT ?",
		limit*2, // 多取一些，因为需要重新计算连续天数排序
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 计算连续打卡天数并排序
	type userRank struct {
		UserID         int64
		Record         int64
		ContinuousDays int
	}

	var ranks []userRank
	for rows.Next() {
		var r userRank
		if err := rows.Scan(&r.UserID, &r.Record); err != nil {
			return nil, err
		}

		// 转换并计算连续天数
		record := FromInt64(r.Record)
		r.ContinuousDays = record.ContinuousDays()

		ranks = append(ranks, r)
	}

	// 按连续天数排序（简单冒泡排序，生产环境建议用标准库sort）
	for i := 0; i < len(ranks)-1; i++ {
		for j := 0; j < len(ranks)-i-1; j++ {
			if ranks[j].ContinuousDays < ranks[j+1].ContinuousDays {
				ranks[j], ranks[j+1] = ranks[j+1], ranks[j]
			}
		}
	}

	// 取前N名
	if len(ranks) > limit {
		ranks = ranks[:limit]
	}

	// 转换为返回格式
	result := make([]map[string]interface{}, len(ranks))
	for i, r := range ranks {
		record := FromInt64(r.Record)
		result[i] = map[string]interface{}{
			"user_id":         r.UserID,
			"continuous_days": r.ContinuousDays,
			"total_7days":     record.TotalDaysInPeriod(7),
			"recent":          record.StringWithDays(7),
		}
	}

	return result, nil
}

// ==================== 完整示例 ====================

// Example_DatabaseUsage 数据库使用完整示例
func Example_DatabaseUsage() {
	// 1. 连接数据库
	db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/dbname?parseTime=true")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer db.Close()

	// 2. 创建打卡服务
	service := NewCheckinService(db)
	ctx := context.Background()

	// 3. 用户打卡
	userID := int64(12345)
	if err := service.UserCheckin(ctx, userID); err != nil {
		fmt.Printf("打卡失败: %v\n", err)
	} else {
		fmt.Println("打卡成功！")
	}

	// 4. 获取打卡统计
	stats, err := service.GetUserCheckinStats(ctx, userID)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("连续打卡: %d天\n", stats["continuous_days"])
	fmt.Printf("最近7天: %s\n", stats["recent_7days"])
	fmt.Printf("打卡率: %.1f%%\n", stats["rate_7days"].(float64)*100)

	// 5. 获取排行榜
	leaderboard, _ := service.GetCheckinLeaderboard(ctx, 10)
	fmt.Println("\n=== 连续打卡排行榜 ===")
	for i, user := range leaderboard {
		fmt.Printf("%d. 用户%d - 连续%d天 %s\n",
			i+1,
			user["user_id"],
			user["continuous_days"],
			user["recent"],
		)
	}
}

// ==================== SQL 语句参考 ====================

const (
	// 建表语句
	CreateTableSQL = `
CREATE TABLE user_checkin (
    user_id BIGINT PRIMARY KEY COMMENT '用户ID',
    record BIGINT NOT NULL DEFAULT 0 COMMENT '打卡记录',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_record (record) COMMENT '排行榜查询索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户打卡记录表';
`

	// 用户打卡
	CheckinSQL = `
INSERT INTO user_checkin (user_id, record) VALUES (?, 1) ON DUPLICATE KEY UPDATE record = record | 1
`

	// 查询打卡记录
	GetRecordSQL = `SELECT record FROM user_checkin WHERE user_id = ?`

	// 每日推移（定时任务）
	DailyShiftSQL = `UPDATE user_checkin SET record = record << 1
`
	// 清理过期数据（可选，删除长期未打卡的用户）
	CleanupSQL = `DELETE FROM user_checkin WHERE record = 0 AND updated_at < DATE_SUB(NOW(), INTERVAL 90 DAY)
`
)

// ==================== 使用说明 ====================

/*
完整使用流程：

1. 创建数据库表：
   执行 CreateTableSQL

2. 在代码中使用：
   - 读取：SELECT record FROM user_checkin WHERE user_id = ?
   - 转换：record := checkin.FromInt64(dbValue)
   - 操作：record = record.Checkin()
   - 保存：UPDATE user_checkin SET record = ? WHERE user_id = ?

3. 定时任务（每天0点）：
   执行 DailyShiftSQL，将所有记录左移1位

4. 重要提示：
   - 数据库字段必须用 BIGINT，不能用 INT
   - Go 代码中用 int64 类型接收和存储
   - 一次查询就能计算所有统计，不需要额外SQL
*/
