package controllers

import (
	"app/service"
	"github.com/rs/zerolog/log"
	"net/http"
)

type LoginController struct {
	Base
}

func (c *LoginController) Get() {
	if c.GetSession("user") != nil {
		c.Redirect("/", http.StatusFound)
		return
	}
	c.TplName = "login.tpl"
}

func (c *LoginController) Post() {
	user, err := service.GetUserLoginInfo(c.GetString("login"), c.GetString("password"))
	if err != nil {
		log.Err(err).Msgf("failed login")
		c.AbortInternalError()
	}

	if user != nil {
		c.userToSession(*user)
		c.redirectToHome()
		return
	}
	c.TplName = "login.tpl"
}
