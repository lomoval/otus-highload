package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type User_20220319_102647 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &User_20220319_102647{}
	m.Created = "20220319_102647"

	migration.Register("User_20220319_102647", m)
}

// Run the migrations
func (m *User_20220319_102647) Up() {
	m.SQL(
		`CREATE TABLE sex (
  id tinyint(3) unsigned NOT NULL,
  name varchar(50) NOT NULL,
  PRIMARY KEY (id)
)`)

	m.SQL(
		`CREATE TABLE user (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  login varchar(100) NOT NULL,
	password varchar(100) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY user_UN (login)
)`)

	m.SQL(
		`CREATE TABLE db.friend (
	id BIGINT auto_increment NOT NULL,
	user_id_1 BIGINT NOT NULL,
	user_id_2 BIGINT NOT NULL,
	CONSTRAINT friend_PK PRIMARY KEY (id),
  UNIQUE KEY friend_UN (user_id_1,user_id_2),
	CONSTRAINT friend_user_1_FK FOREIGN KEY (user_id_1) REFERENCES user(id) ON DELETE CASCADE,
	CONSTRAINT friend_user_2_FK FOREIGN KEY (user_id_1) REFERENCES user(id) ON DELETE CASCADE)
`)

	m.SQL(
		`CREATE TABLE profile (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  user_id bigint(20) NOT NULL,
  name varchar(100) NOT NULL,
  surname varchar(100) NOT NULL,
  birth_date DATE NOT NULL,
  sex_id tinyint(3) unsigned NOT NULL,
  city varchar(125) DEFAULT NULL,
	PRIMARY KEY (id),
  KEY profile_user_FK (user_id),
  KEY profile_sex_FK (sex_id),
  CONSTRAINT profile_user_FK FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE,
  CONSTRAINT profile_sex_FK FOREIGN KEY (sex_id) REFERENCES sex (id)
)`)

	m.SQL(
		`CREATE TABLE interest (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  name varchar(250) NOT NULL,
  user_id bigint(20) NOT NULL,
  PRIMARY KEY (id),
  KEY interest_FK (user_id),
  CONSTRAINT interest_user_FK FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE)
`)

	// Data
	m.SQL(
		`INSERT INTO db.sex (id,name) VALUES
	 (1,'Male'),
	 (2,'Female'),
	 (3,'Don''t know');
`)

}

// Reverse the migrations
func (m *User_20220319_102647) Down() {
	m.SQL("DROP TABLE IF EXISTS profile")
	m.SQL("DROP TABLE IF EXISTS interest")
	m.SQL("DROP TABLE IF EXISTS sex")
	m.SQL("DROP TABLE IF EXISTS friend")
	m.SQL("DROP TABLE IF EXISTS user")
}
