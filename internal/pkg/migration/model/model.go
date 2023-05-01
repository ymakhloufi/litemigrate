package model

import "time"

type Migration struct {
	ID          uint
	Filename    string
	StartedAt   time.Time
	CompletedAt *time.Time
}
