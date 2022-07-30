package service

import (
	"app/models"
	"context"
	"errors"
	"github.com/beego/beego/v2/client/orm"
)

func PrivateDialogs(_ context.Context, userID int64) ([]models.Dialog, error) {
	o := getReadOrm()

	var ids []int64
	var usersIds []int64
	var names []string
	var surnames []string
	_, err := o.Raw(`
	SELECT dialog_id, p.user_id, name, surname 
	FROM (SELECT id dialog_id, IF(user_id_1 = ?, user_id_2, user_id_1) user_id 
				FROM private_dialog 
				WHERE user_id_1 = ? OR user_id_2 = ?) d
	JOIN profile p ON p.user_id = d.user_id 
	`, userID, userID, userID).
		QueryRows(&ids, &usersIds, &names, &surnames)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	dialogs := make([]models.Dialog, 0, len(ids))
	for i, id := range ids {
		dialogs = append(dialogs, models.Dialog{ID: id, Creator: models.User{
			Id:        usersIds[i],
			Profile:   models.Profile{Name: names[i], Surname: surnames[i]},
			Interests: nil,
		}})
	}

	return dialogs, nil
}

func PrivateDialog(_ context.Context, userID int64, id int64) (models.Dialog, error) {
	o := getReadOrm()

	var dialog models.Dialog
	err := o.Raw("SELECT id, IF(user_id_1 = ?, user_id_2, user_id_1) user_id FROM private_dialog "+
		"WHERE id=? ORDER BY id ASC;", userID, id).
		QueryRow(&dialog.ID, &dialog.Creator.Id)

	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return dialog, nil
		}
		return dialog, err
	}

	return dialog, nil
}

func AddPrivateDialog(_ context.Context, userID1 int64, userID2 int64) (int64, error) {
	if userID1 > userID2 {
		t := userID1
		userID1 = userID2
		userID2 = t
	}
	var dialogId int64
	err := orm.NewOrm().Raw("SELECT id FROM private_dialog WHERE user_id_1=? AND user_id_2=?", userID1, userID2).QueryRow(&dialogId)
	if err != nil {
		return 0, err
	}
	if dialogId == 0 {
		r, err := orm.NewOrm().Raw("INSERT INTO private_dialog(user_id_1, user_id_2) VALUES(?, ?) ", userID1, userID2).Exec()
		dialogId, err = r.LastInsertId()
		if err != nil {
			return 0, err
		}
	}
	return dialogId, err
}

func PrivateDialogAnswers(_ context.Context, dialogID int64) ([]models.DialogAnswer, error) {
	o := getReadOrm()

	var ids []int64
	var cretorsIDs []int64
	var texts []string
	_, err := o.Raw("SELECT id as \"ID\", text, creator_id FROM private_dialog_answer WHERE dialog_id=? ORDER BY create_timestamp ASC;", dialogID).
		QueryRows(&ids, &texts, &cretorsIDs)

	answers := make([]models.DialogAnswer, 0, len(ids))
	for i, id := range ids {
		answers = append(answers, models.DialogAnswer{ID: id, Text: texts[i], Creator: models.User{Id: cretorsIDs[i]}})
	}

	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return answers, nil
}

func AddPrivateDialogAnswer(_ context.Context, dialogID int64, creatorID int64, text string) error {
	_, err := orm.NewOrm().Raw("INSERT INTO private_dialog_answer(dialog_id, creator_id, text) VALUES(?, ?, ?) ",
		dialogID, creatorID, text).Exec()
	return err
}
