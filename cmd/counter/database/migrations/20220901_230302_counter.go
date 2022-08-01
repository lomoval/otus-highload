package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type Dialog_20220508_230302 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &Dialog_20220508_230302{}
	m.Created = "20220508_230302"

	migration.Register("Dialog_20220508_230302", m)
}

// Run the migrations
func (m *Dialog_20220508_230302) Up() {
	m.SQL(
		`
CREATE TABLE private_message_counter (
  id bigint NOT NULL AUTO_INCREMENT,
  dialog_id bigint NOT NULL,
  check_timestamp timestamp NULL DEFAULT NULL,
  creator_id bigint NOT NULL,
  count bigint NOT NULL DEFAULT '1',
  PRIMARY KEY (id),
  UNIQUE KEY private_message_counter_UN (dialog_id,creator_id)
)`)

	m.SQL(
		`CREATE TABLE dialog_answer (
	id BIGINT auto_increment NOT NULL,
	` + "`text` varchar(4000) NOT NULL," +
			`creator_id BIGINT NOT NULL,
	dialog_id BIGINT NOT NULL,
  create_timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT dialog_answer_PK PRIMARY KEY (id),
	CONSTRAINT dialog_answer_FK_1 FOREIGN KEY (dialog_id) REFERENCES dialog(id) ON DELETE CASCADE
)`)

}

// Reverse the migrations
func (m *Dialog_20220508_230302) Down() {
	m.SQL("DROP TABLE IF EXISTS dialog_answer")
	m.SQL("DROP TABLE IF EXISTS dialog")
}
