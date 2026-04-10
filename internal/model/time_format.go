package model

import (
	"fmt"
	"time"
)

// FormatFriendlyTime 将时间转换为友好格式
// 示例：刚刚、1分钟前、1小时前、昨天、2024-01-15
func FormatFriendlyTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// 计算时间差（秒）
	seconds := int64(diff.Seconds())

	switch {
	case seconds < 0:
		return "刚刚"
	case seconds < 60:
		return "刚刚"
	case seconds < 3600:
		minutes := seconds / 60
		return fmt.Sprintf("%d分钟前", minutes)
	case seconds < 86400:
		hours := seconds / 3600
		return fmt.Sprintf("%d小时前", hours)
	case seconds < 172800: // 2天内
		return "昨天"
	case seconds < 259200: // 3天内
		return "前天"
	default:
		// 超过3天显示具体日期
		return t.Format("2006-01-02")
	}
}

// FormatFriendlyTimeWithTime 带时间的友好格式（用于更精确的显示）
func FormatFriendlyTimeWithTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	seconds := int64(diff.Seconds())

	switch {
	case seconds < 0:
		return "刚刚"
	case seconds < 60:
		return "刚刚"
	case seconds < 3600:
		minutes := seconds / 60
		return fmt.Sprintf("%d分钟前", minutes)
	case seconds < 86400:
		hours := seconds / 3600
		return fmt.Sprintf("%d小时前", hours)
	case seconds < 604800: // 7天内
		days := seconds / 86400
		return fmt.Sprintf("%d天前", days)
	default:
		// 超过7天显示完整日期
		return t.Format("2006-01-02")
	}
}

// FormatDateTime 标准日期时间格式
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
