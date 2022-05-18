package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type News_20220514_202717 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &News_20220514_202717{}
	m.Created = "20220514_202717"

	migration.Register("News_20220514_202717", m)
}

// Run the migrations
func (m *News_20220514_202717) Up() {
	m.SQL(
		`CREATE TABLE news (
	id BIGINT auto_increment NOT NULL,
	text varchar(1024) NOT NULL,
	creator_id BIGINT NOT NULL,
  create_timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT news_PK PRIMARY KEY (id),
	CONSTRAINT news_FK FOREIGN KEY (creator_id) REFERENCES user(id) ON DELETE RESTRICT
)`)
}

// Reverse the migrations
func (m *News_20220514_202717) Down() {
	m.SQL("DROP TABLE IF EXISTS news")
}
