package controllers

import (
	"app/service"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

type PrivateDialogController struct {
	Base
}

// @router /privatedialogs [get]
func (c *DialogController) PrivateDialogs() {
	dialogs, err := service.PrivateDialogs(c.ReqCtx(), c.user().Id)
	if err != nil {
		log.Err(err).Msgf("failed to get dialogs")
		c.AbortInternalError()
	}

	c.Data["Dialogs"] = dialogs
	c.TplName = "privatedialogs.tpl"
}

// @router /privatedialog [post]
func (c *ProfileController) AddPrivateDialog() {
	userID, err := c.GetInt64("user_id")
	if err != nil {
		log.Error().Msgf("failed to post private answers")
		c.AbortBadRequest()
		return
	}
	dialogID, err := service.AddPrivateDialog(c.ReqCtx(), c.user().Id, userID)
	if err != nil {
		log.Err(err).Msgf("failed to add dialog")
		c.AbortInternalError()
	}

	c.Redirect(fmt.Sprintf("/privatedialogs/%d", dialogID), http.StatusFound)
}

// @router /privatedialogs/:id [get]
func (c *DialogController) PrivateDialogAnswers(id int64) {
	dialog, err := service.PrivateDialog(c.ReqCtx(), c.user().Id, id)
	if err != nil {
		log.Err(err).Msgf("failed to get dialog '%d' answers", id)
		c.AbortInternalError()
	}

	answers, err := service.PrivateDialogAnswers(c.ReqCtx(), id)
	if err != nil {
		log.Err(err).Msgf("failed to get dialog '%d' answers", id)
		c.AbortInternalError()
	}

	c.Data["Dialog"] = dialog
	c.Data["Answers"] = answers
	c.TplName = "privateanswers.tpl"
}

// @router /privatedialogs/:id [post]
func (c *ProfileController) AddPrivateDialogAnswer(id int64) {
	text := c.GetString("text")
	if text == "" {
		log.Error().Msgf("dialog answer should have value")
		c.AbortBadRequest()
		return
	}
	err := service.AddPrivateDialogAnswer(c.ReqCtx(), id, c.user().Id, text)
	if err != nil {
		log.Err(err).Msgf("failed to add dialog answer")
		c.AbortInternalError()
	}

	c.Redirect(fmt.Sprintf("/privatedialogs/%d", id), http.StatusFound)
}
