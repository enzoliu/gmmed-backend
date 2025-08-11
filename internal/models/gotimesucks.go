package models

import (
	"fmt"
	"strings"
	"time"
)

type GoTimeSucks struct {
	time.Time
}

var timeFormats = []string{
	time.RFC3339, // "2024-10-10T10:30:00Z"
	"2006-01-02", // "2024-10-10"
}

// 自訂 JSON 解析邏輯
func (ft *GoTimeSucks) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("invalid time format: %s", s)
}

// 範例結構
type Payload struct {
	Date GoTimeSucks `json:"date"`
}
