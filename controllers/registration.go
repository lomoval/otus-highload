package controllers

import (
	"app/service"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type RegistrationController struct {
	Base
}

func (c *RegistrationController) Get() {
	if c.GetSession("user") != nil {
		c.Redirect("/", http.StatusFound)
		return
	}

	c.TplName = "registration.tpl"
}

func (c *RegistrationController) Post() {
	user, err := c.fillUser(nil)
	p := c.GetString("password")

	if user.Login == "" || p == "" || p != c.GetString("repeat-password") ||
		!service.ValidateProfileData(user.Profile) {
		c.AbortBadRequest()
	}

	user, err = service.CreateUser(user, p)
	if err != nil {
		log.Err(err).Msgf("failed to create user")
		if errors.Is(err, service.ErrDuplicate) {
			c.AbortBadRequest()
		}
		c.AbortInternalError()
	}

	c.userToSession(user)
	c.userToViewModel(user)
	c.redirectToHome()
}
