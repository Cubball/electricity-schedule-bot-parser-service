package models

import "time"

type ScheduleEntry struct {
    QueueNumber string    `json:"queueNumber"`
    Date        time.Time `json:"date"`
    StartTime   time.Time `json:"startTime"`
    EndTime     time.Time `json:"endTime"`
}
