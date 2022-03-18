package controllers

import (
	"app/models"
	beego "github.com/beego/beego/v2/server/web"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"reflect"
)

const (
	templateHome  = "/"
	templateLogin = "/login"
)

type Base struct {
	beego.Controller
}

func (c *Base) userToSession(user models.User) {
	c.SetSession("user", user)
}

func (c *Base) user() *models.User {
	v := c.GetSession("user")
	if v == nil {
		return nil
	}
	u, ok := v.(models.User)
	if !ok {
		log.Error().Msgf("incorrect user structure in session, type: %s", reflect.TypeOf(u))
		return nil
	}
	return &u
}

func (c *Base) isUserLogged() bool {
	v := c.GetSession("user")
	if v == nil {
		return false
	}
	u, ok := v.(models.User)
	if !ok {
		log.Error().Msgf("incorrect user structure in session, type: %s", reflect.TypeOf(u))
		return false
	}
	return true
}

func (c *Base) AbortInternalError() {
	c.Abort(strconv.Itoa(http.StatusInternalServerError))
}

func (c *Base) AbortBadRequest() {
	c.Abort(strconv.Itoa(http.StatusBadRequest))
}

func (c *Base) userToViewModel(user models.User) {
	c.Data["User"] = user
}

func (c *Base) userToViewModelFromSession() {
	u := c.user()
	if u != nil {
		c.Data["User"] = *u
	}
}

func (c *Base) homeTpl() {
	c.TplName = templateHome
}

func (c *Base) loginTpl() {
	c.TplName = templateLogin
}

func (c *Base) redirectToHome() {
	c.Redirect("/", http.StatusFound)
}

func (c *Base) fillUser(user *models.User) (models.User, error) {
	u := models.User{}
	if user != nil {
		u = *user
	}
	birthDate := c.GetString("birthdate")
	t, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return u, err
	}
	sex, err := c.GetInt("sex")
	if err != nil {
		return u, err
	}

	interestsStrings := strings.Split(c.GetString("interests"), "\n")
	interests := make([]models.Interest, 0, len(interestsStrings))
	for _, interest := range interestsStrings {
		interest = strings.TrimSpace(interest)
		if interest != "" {
			interests = append(interests, models.Interest{Name: interest})
		}
	}

	u.Login = c.GetString("login")
	u.Profile.Name = c.GetString("name")
	u.Profile.Surname = c.GetString("surname")
	u.Profile.BirthDate = t
	u.Profile.City = c.GetString("city")
	u.Profile.Sex = models.Sex{Id: sex, Name: ""}
	u.Interests = interests
	return u, nil
}
