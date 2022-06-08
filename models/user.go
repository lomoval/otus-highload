package models

import "time"

type User struct {
	Id        int64 `msgpack:"id"`
	Login     string
	Profile   Profile    `msgpack:"profile"`
	Interests []Interest `msgpack:"interests"`
}

func (u User) Valid() bool {
	return u.Id > 0
}

type Profile struct {
	Id        int64     `msgpack:"id"`
	Name      string    `msgpack:"name"`
	Surname   string    `msgpack:"surname"`
	BirthDate time.Time `msgpack:"birth_date"`
	City      string    `msgpack:"city"`
	Sex       Sex       `msgpack:"sex"`
}

type Sex struct {
	Id   int    `msgpack:"id"`
	Name string `msgpack:"name"`
}

type Interest struct {
	Id   int    `msgpack:"id"`
	Name string `msgpack:"name"`
}
