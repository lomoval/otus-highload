package service

import (
	"app/models"
	"errors"
	"github.com/tarantool/go-tarantool"
)

var ErrTarantoolNotAvailable = errors.New("tarantool not available")
var conn *tarantool.Connection

func SetupTarantool(server string, user string, password string) error {
	if server == "" {
		return nil
	}

	opts := tarantool.Opts{User: user, Pass: password}
	var err error
	conn, err = tarantool.Connect(server, opts)
	return err
}

func ShutdownTarantool() {
	if conn != nil {
		conn.Close()
	}
}

func ProfileFromTarantool(id int64) (models.User, error) {
	if conn == nil {
		return models.User{}, ErrTarantoolNotAvailable
	}
	user := models.User{}
	users := [][]*models.User{{&user}}
	if err := conn.CallTyped("load_profile", []interface{}{id}, &users); err != nil {
		return models.User{}, err
	}

	if user.Id > 0 {
		return user, nil
	}
	return models.User{}, nil
}
