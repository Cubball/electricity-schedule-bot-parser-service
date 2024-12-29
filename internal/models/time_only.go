package models

import (
	"encoding/json"
	"time"
)

type TimeOnly time.Time

const TimeOnlyLayout = "15:04"

func (t TimeOnly) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Time(t).Format(TimeOnlyLayout))
}
