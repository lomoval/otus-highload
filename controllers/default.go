package controllers

import "github.com/rs/zerolog/log"

type MainController struct {
	Base
}

func (c *MainController) Get() {
	log.Info().Msgf("%v", c.user())
	c.userToViewModelFromSession()
	c.Data["Website"] = "beego!!.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}
