package connectors

import "time"

type Downtime struct {
	Author    string
	Comment   string
	StartTime time.Time
	EndTime   time.Time
}
