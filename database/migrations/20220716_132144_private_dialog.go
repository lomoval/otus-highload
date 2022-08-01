package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type PrivateDialog_20220716_132144 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &PrivateDialog_20220716_132144{}
	m.Created = "20220716_132144"

	migration.Register("PrivateDialog_20220716_132144", m)
}

// Run the migrations
func (m *PrivateDialog_20220716_132144) Up() {
	m.SQL(`CREATE TABLE private_dialog (
	id BIGINT auto_increment NOT NULL,
	user_id_1 BIGINT NOT NULL,
	user_id_2 BIGINT NOT NULL,
	creation_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  UNIQUE KEY private_dialog_UN (user_id_1,user_id_2),
	CONSTRAINT private_dialog_PK PRIMARY KEY (id),
	CONSTRAINT private_dialog_FK FOREIGN KEY (user_id_1) REFERENCES user(id) ON DELETE CASCADE,
	CONSTRAINT private_dialog_FK_1 FOREIGN KEY (user_id_2) REFERENCES user(id) ON DELETE CASCADE`)

	m.SQL(`CREATE TABLE db.private_dialog_answer (
		id BIGINT auto_increment NOT NULL,
		dialog_id BIGINT NOT NULL,
		create_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		creator_id BIGINT NOT NULL,
		text varchar(4000) NOT NULL,
    processed tinyint(1) DEFAULT NULL,
		CONSTRAINT private_dialog_answer_FK FOREIGN KEY (creator_id) REFERENCES db.user(id) ON DELETE CASCADE,
		CONSTRAINT private_dialog_answer_FK_1 FOREIGN KEY (dialog_id) REFERENCES private_dialog (id) ON DELETE CASCADE`)
}

// Reverse the migrations
func (m *PrivateDialog_20220716_132144) Down() {
	m.SQL("DROP TABLE IF EXISTS private_message")
}
