package models

type ScheduleEntry struct {
	QueueNumber string   `json:"queueNumber"`
	Date        DateOnly `json:"date"`
	StartTime   TimeOnly `json:"startTime"`
	EndTime     TimeOnly `json:"endTime"`
}
