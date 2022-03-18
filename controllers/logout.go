package controllers

import (
	"net/http"
)

type LogoutController struct {
	Base
}

func (c *LogoutController) Get() {
	c.DestroySession()
	c.Redirect("/", http.StatusFound)
}
