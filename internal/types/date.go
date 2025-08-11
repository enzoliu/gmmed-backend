package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Date 自定義日期類型，只處理日期部分（不包含時間）
type Date struct {
	time.Time
}

// DateFormat 支援的日期格式
var DateFormats = []string{
	"2006-01-02",           // ISO 8601 日期格式
	"2006/01/02",           // 斜線分隔格式
	"2006-1-2",             // 不補零格式
	"2006/1/2",             // 斜線不補零格式
	"2006-01-02T15:04:05Z", // 完整 ISO 8601 格式
	"2006-01-02 15:04:05",  // 空格分隔格式
}

// NewDate 建立新的日期
func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// ParseDate 解析日期字串
func ParseDate(s string) (Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Date{}, fmt.Errorf("empty date string")
	}

	for _, format := range DateFormats {
		if t, err := time.Parse(format, s); err == nil {
			// 只保留日期部分，時間設為 00:00:00 UTC
			return Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}, nil
		}
	}

	return Date{}, fmt.Errorf("unable to parse date: %s", s)
}

// UnmarshalJSON 實現 JSON 反序列化
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// MarshalJSON 實現 JSON 序列化
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

// String 字串表示
func (d Date) String() string {
	return d.Format("2006-01-02")
}

// IsZero 檢查是否為零值
func (d Date) IsZero() bool {
	return d.Time.IsZero()
}

// Before 檢查是否在指定日期之前
func (d Date) Before(other Date) bool {
	return d.Time.Before(other.Time)
}

// After 檢查是否在指定日期之後
func (d Date) After(other Date) bool {
	return d.Time.After(other.Time)
}

// Equal 檢查是否等於指定日期
func (d Date) Equal(other Date) bool {
	return d.Year() == other.Year() && d.Month() == other.Month() && d.Day() == other.Day()
}

// AddDays 添加天數
func (d Date) AddDays(days int) Date {
	return Date{d.Time.AddDate(0, 0, days)}
}

// AddMonths 添加月數
func (d Date) AddMonths(months int) Date {
	return Date{d.Time.AddDate(0, months, 0)}
}

// AddYears 添加年數
func (d Date) AddYears(years int) Date {
	return Date{d.Time.AddDate(years, 0, 0)}
}

// ToTime 轉換為 time.Time
func (d Date) ToTime() time.Time {
	return d.Time
}

// Today 取得今天的日期
func Today() Date {
	now := time.Now().UTC()
	return Date{time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)}
}

// Yesterday 取得昨天的日期
func Yesterday() Date {
	return Today().AddDays(-1)
}

// Tomorrow 取得明天的日期
func Tomorrow() Date {
	return Today().AddDays(1)
}
