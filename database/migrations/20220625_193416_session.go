package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type Session_20220625_193416 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &Session_20220625_193416{}
	m.Created = "20220625_193416"

	migration.Register("Session_20220625_193416", m)
}

// Run the migrations
func (m *Session_20220625_193416) Up() {
	m.SQL(
		`CREATE TABLE session (
		session_key varchar(64) NOT NULL,
		session_data varbinary(5000) DEFAULT NULL,
		session_expiry int NOT NULL,
		PRIMARY KEY (session_key)
	)`)
}

// Reverse the migrations
func (m *Session_20220625_193416) Down() {
	m.SQL("DROP TABLE IF EXISTS session")
}
