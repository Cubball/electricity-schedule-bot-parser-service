package models

import (
	"encoding/json"
	"time"
)

type DateOnly time.Time

const DateOnlyLayout = "2006-01-02"

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(DateOnlyLayout))
}
