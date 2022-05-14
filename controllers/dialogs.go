package controllers

import (
	"app/service"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

type DialogController struct {
	Base
}

// @router /dialogs [get]
func (c *DialogController) Dialogs() {
	dialogs, err := service.Dialogs()
	if err != nil {
		log.Err(err).Msgf("failed to get dialogs")
		c.AbortInternalError()
	}

	c.Data["Dialogs"] = dialogs
	c.TplName = "dialogs.tpl"
}

// @router /dialogs [post]
func (c *ProfileController) AddDialog() {
	text := c.GetString("name")
	if text == "" {
		log.Error().Msgf("dialog should have name")
		c.AbortBadRequest()
		return
	}
	err := service.AddDialog(c.user().Id, text)
	if err != nil {
		log.Err(err).Msgf("failed to add dialog")
		c.AbortInternalError()
	}

	c.Redirect("/dialogs", http.StatusFound)
}

// @router /dialogs/:id [get]
func (c *DialogController) DialogAnswers(id int64) {
	dialog, err := service.Dialog(id)
	if err != nil {
		log.Err(err).Msgf("failed to get dialog '%d' answers", id)
		c.AbortInternalError()
	}

	answers, err := service.DialogAnswers(id)
	if err != nil {
		log.Err(err).Msgf("failed to get dialog '%d' answers", id)
		c.AbortInternalError()
	}

	c.Data["Dialog"] = dialog
	c.Data["Answers"] = answers
	c.TplName = "answers.tpl"
}

// @router /dialogs/:id [post]
func (c *ProfileController) AddDialogAnswer(id int64) {
	text := c.GetString("text")
	if text == "" {
		log.Error().Msgf("dialog answer should have value")
		c.AbortBadRequest()
		return
	}
	err := service.AddDialogAnswer(id, c.user().Id, text)
	if err != nil {
		log.Err(err).Msgf("failed to add dialog answer")
		c.AbortInternalError()
	}

	c.Redirect(fmt.Sprintf("/dialogs/%d", id), http.StatusFound)
}
