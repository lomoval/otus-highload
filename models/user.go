package models

import "time"

type User struct {
	Id        int64
	Login     string
	Profile   Profile
	Interests []Interest
}

type Profile struct {
	Id        int64
	Name      string
	Surname   string
	BirthDate time.Time
	City      string
	Sex       Sex
}

type Sex struct {
	Id   int
	Name string
}

type Interest struct {
	Id   int
	Name string
}
