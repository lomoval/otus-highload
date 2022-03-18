package controllers

import (
	"app/service"
	"github.com/rs/zerolog/log"
)

type ProfileController struct {
	Base
}

// @router /profile [get]
func (c *ProfileController) Profile(id *int64) {
	userId := c.user().Id
	if id != nil {
		userId = *id
	}
	u, err := service.Profile(userId)
	if err != nil {
		log.Err(err).Msgf("failed to get profile by id [%d]", id)
		c.AbortInternalError()
	}

	c.Data["ReadOnly"] = userId != c.user().Id
	c.Data["Profile"] = u.Profile
	c.Data["Interests"] = u.Interests
	c.TplName = "profile.tpl"
}

// @router /profile [post]
func (c *ProfileController) Post() {
	u, err := c.fillUser(c.user())
	if err != nil {
		log.Err(err).Msgf("failed to fill user")
		c.AbortInternalError()
	}

	err = service.SaveUser(u)
	if err != nil {
		log.Err(err).Msgf("failed to save user")

	}

	c.Data["Profile"] = u.Profile
	c.Data["Interests"] = u.Interests
	c.TplName = "profile.tpl"
}
