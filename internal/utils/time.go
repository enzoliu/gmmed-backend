package utils

import (
	"fmt"
	"time"
)

func ParseTaiwanDateToUTC(dateStr string) (time.Time, error) {
	layout := "2006-01-02 15:04:05" // 若包含時間則用 "2006-01-02 15:04:05"
	if len(dateStr) == 10 {
		layout = "2006-01-02"
	}

	// 載入 Asia/Taipei 時區
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.Time{}, fmt.Errorf("load location error: %w", err)
	}

	// 先解析為 Asia/Taipei 時區的時間
	localTime, err := time.ParseInLocation(layout, dateStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time error: %w", err)
	}

	// 再轉為 UTC
	return localTime.UTC(), nil
}
