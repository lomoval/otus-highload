package main

import (
	_ "app/routers"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"net/http"
	"strings"
)

func init() {
	viper.SetEnvPrefix("OTUS_HIGHLOAD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASS")

	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8",
			viper.Get("DB_USER"),
			viper.Get("DB_PASS"),
			viper.Get("DB_HOST"),
			viper.Get("DB_PORT"),
			viper.Get("DB_NAME"),
		))
}

func main() {
	var FilterUser = func(ctx *context.Context) {
		if strings.HasPrefix(ctx.Input.URL(), "/login") {
			return
		}
		// if strings.HasPrefix(ctx.Input.URL(), "/favicon") {
		// 	return
		// }

		ok := ctx.Input.Session("user")
		if ok == nil && strings.HasPrefix(ctx.Input.URL(), "/registration") {
			return
		}

		if ok == nil {
			ctx.Redirect(http.StatusFound, "/login")
			return
		}

		log.Debug().Msgf("user is logged and go to %s", ctx.Input.URL())
	}

	beego.InsertFilter("/*", beego.BeforeRouter, FilterUser)
	beego.Run()
}
