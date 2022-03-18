package controllers

import (
	"app/service"
	"github.com/rs/zerolog/log"
)

type UsersController struct {
	Base
}

func (c *UsersController) Get() {
	u := c.user()
	users, err := service.Users(*u)

	c.TplName = "users.tpl"
	if err != nil {
		log.Err(err).Msgf("failed to get users")
		return
	}
	c.Data["Users"] = users
}

func (c *UsersController) Post() {
	friendID, err := c.GetInt64("friend_id")
	if err != nil {
		log.Err(err).Msgf("incorrect friend id parameter [%s]", c.GetString("friend_id"))
	}
	err = service.AddFriend(c.user().Id, friendID)
	if err != nil {
		log.Err(err).Msgf("failed to add friend")
		c.AbortInternalError()
	}
	c.Get()
}
