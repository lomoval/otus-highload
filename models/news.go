package models

import "time"

type News struct {
	ID        int64
	Text      string
	Creator   User
	Timestamp time.Time
}
