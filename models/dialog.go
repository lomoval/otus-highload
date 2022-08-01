package models

import "time"

type Dialog struct {
	ID      int64
	Name    string
	Creator User
	Count   int32
}

type DialogAnswer struct {
	ID      int64
	Creator User
	Text    string
}

type PrivateDialogMessage struct {
	Id         int64
	DialogId   int64
	FromUserId int64
	ToUserId   int64
	Timestamp  time.Time
}
