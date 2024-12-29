package models

import "time"

type Schedule struct {
    Entries   []ScheduleEntry `json:"entries"`
    FetchTime time.Time       `json:"fetchTime"`
}
