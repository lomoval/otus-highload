package controllers

import (
	"app/service"
	"github.com/beego/beego/v2/client/orm"
	"github.com/rs/zerolog/log"
)

type UsersController struct {
	Base
}

func (c *UsersController) Get() {
	u := c.user()
	paging, err := getPageParameters(&c.Base.Controller)
	if err != nil {
		log.Err(err).Msg("incorrect paging parameters")
		c.AbortBadRequest()
	}
	name := c.GetString("searchName")
	surname := c.GetString("searchSurname")
	if (name == "" && surname != "") || (name != "" && surname == "") {
		log.Error().Msgf("incorrect searching parameters")
		c.AbortBadRequest()
	}

	var users []orm.Params
	switch {
	case name != "" && surname != "":
		users, err = service.FindUsers(*u, paging.Limit, paging.Offset, name, surname)
		log.Debug().Msgf("founded users'%d'", len(users))
	default:
		users, err = service.Users(*u, paging.Limit, paging.Offset)
	}

	c.TplName = "users.tpl"
	if err != nil {
		log.Err(err).Msgf("failed to get users")
		return
	}
	c.Data["Users"] = users
	c.Data["Paging"] = paging
	c.Data["SearchName"] = name
	c.Data["SearchSurname"] = surname
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
