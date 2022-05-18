package controllers

import (
	"app/service"
	"github.com/rs/zerolog/log"
	"net/http"
)

type NewsController struct {
	Base
}

// @router /news [get]
func (c *NewsController) Get() {
	news, err := service.News(c.user().Id)
	if err != nil {
		log.Err(err).Msgf("failed to get news of user '%d'", c.user().Id)
		c.AbortInternalError()
		return
	}

	c.Data["News"] = news
	c.TplName = "news.tpl"
}

// @router /friends/news [get]
func (c *NewsController) GetFriendsNews() {
	news := service.CachedNews(c.user().Id)
	if news == nil {
		log.Debug().Msgf("news from db")
		var err error
		news, err = service.GetFriendsNews(c.user().Id)
		if err != nil {
			log.Err(err).Msgf("failed to get friends news")
			c.AbortInternalError()
			return
		}
	}

	c.Data["News"] = news
	c.TplName = "friendsnews.tpl"
}

// @router /news/ [post]
func (c *NewsController) Post() {
	text := c.GetString("text")
	if text == "" {
		log.Error().Msgf("news can not be empty")
		c.AbortBadRequest()
		return
	}

	err := service.AddNews(c.user().Id, text)
	if err != nil {
		log.Err(err).Msgf("failed to add news")
		c.AbortInternalError()
		return
	}

	c.Redirect("/news", http.StatusFound)
}
