package controllers

import (
	"app/service"
	"github.com/rs/zerolog/log"
)

type FriendsController struct {
	Base
}

func (c *FriendsController) Get() {
	u := c.user()
	paging, err := getPageParameters(&c.Base.Controller)
	if err != nil {
		log.Err(err).Msg("incorrect paging parameters")
		c.AbortBadRequest()
	}
	friends, err := service.Friends(*u, paging.Limit, paging.Offset)

	c.TplName = "friends.tpl"
	if err != nil {
		log.Err(err).Msgf("failed to get friends")
		return
	}
	c.Data["Users"] = friends
	c.Data["Paging"] = paging
}

func (c *FriendsController) Post() {
	friendID, err := c.GetInt64("friend_id")
	if err != nil {
		log.Err(err).Msgf("incorrect friend id parameter [%s]", c.GetString("friend_id"))
	}
	err = service.RemoveFriend(c.user().Id, friendID)
	if err != nil {
		log.Err(err).Msgf("failed to remove friend")
		c.AbortInternalError()
	}
	c.Get()
}
