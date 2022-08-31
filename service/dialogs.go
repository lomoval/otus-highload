//go:build dialogsservice
// +build dialogsservice

package service

import (
	"app/models"
	"context"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"strconv"
)

func Dialogs(_ context.Context) ([]models.Dialog, error) {
	o := getReadOrm()

	var ids []int64
	var names []string
	var creatorIDs []int64
	_, err := o.Raw(`SELECT id AS "ID", name AS "Name", creator_id AS "CreatorID" FROM dialog ORDER BY id ASC;`).
		QueryRows(&ids, &names, &creatorIDs)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	dialogs := make([]models.Dialog, 0, len(ids))
	for i, id := range ids {
		dialogs = append(dialogs, models.Dialog{ID: id, Name: names[i]})
	}

	return dialogs, nil
}

func Dialog(_ context.Context, id int64) (models.Dialog, error) {
	o := getReadOrm()

	var dialog models.Dialog
	err := o.Raw("SELECT id, name, creator_id FROM dialog WHERE id=? ORDER BY id ASC;", id).
		QueryRow(&dialog.ID, &dialog.Name, &dialog.Creator.Id)

	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return dialog, nil
		}
		return dialog, err
	}

	return dialog, nil
}

func AddDialog(_ context.Context, creatorID int64, name string) error {
	_, err := orm.NewOrm().Raw("INSERT INTO dialog(creator_id, name) VALUES(?, ?) ", creatorID, name).Exec()
	return err
}

func DialogAnswers(_ context.Context, dialogID int64) ([]models.DialogAnswer, error) {
	o := getReadOrm()

	var ids []int64
	var texts []string
	_, err := o.Raw(
		dialogShardComment(dialogID)+
			" SELECT id as \"AD\", text FROM dialog_answer WHERE dialog_id=? ORDER BY create_timestamp ASC;", dialogID).
		QueryRows(&ids, &texts)

	answers := make([]models.DialogAnswer, 0, len(ids))
	for i, id := range ids {
		answers = append(answers, models.DialogAnswer{ID: id, Text: texts[i]})
	}

	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return answers, nil
}

func AddDialogAnswer(_ context.Context, dialogID int64, creatorID int64, text string) error {
	_, err := orm.NewOrm().Raw(
		dialogShardComment(dialogID)+
			" INSERT INTO dialog_answer(dialog_id, creator_id, text) VALUES(?, ?, ?) ",
		dialogID, creatorID, text).Exec()
	return err
}

func dialogShardComment(dialogID int64) string {
	return "/* dialog_shard=" + strconv.Itoa(int(dialogID%3)) + " */"
}
