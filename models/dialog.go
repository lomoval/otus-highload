package models

type Dialog struct {
	ID      int64
	Name    string
	Creator User
}

type DialogAnswer struct {
	ID      int64
	Creator User
	Text    string
}
