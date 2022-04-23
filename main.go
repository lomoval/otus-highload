package main

import (
	_ "app/routers"
	"app/service"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strconv"
	"time"

	"net/http"
	"strings"
)

var globalSessions *session.Manager

const (
	maxIdle = 50
	maxConn = 100
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	viper.SetEnvPrefix("OTUS_HIGHLOAD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASS")
	viper.BindEnv("DB_MAX_IDLE")
	viper.BindEnv("DB_MAX_CONN")
	viper.BindEnv("DB_SLAVES")

	viper.SetDefault("DB_MAX_IDLE", maxIdle)
	viper.SetDefault("DB_MAX_CONN", maxConn)

	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8",
			viper.Get("DB_USER"),
			viper.Get("DB_PASS"),
			viper.Get("DB_HOST"),
			viper.Get("DB_PORT"),
			viper.Get("DB_NAME"),
		),
		orm.MaxIdleConnections(viper.GetInt("DB_MAX_IDLE")),
		orm.MaxOpenConnections(viper.GetInt("DB_MAX_CONN")),
	)

	slavesString := viper.GetString("DB_SLAVES")
	if slavesString != "" {
		slaves := strings.Split(slavesString, ",")
		service.SlavesCount = len(slaves)
		for i, slave := range slaves {
			parts := strings.SplitN(slave, ":", 2)
			host := parts[0]
			if len(parts) > 1 {
				host += ":" + parts[1]
			}
			err := orm.RegisterDataBase("slave"+strconv.Itoa(i), "mysql",
				fmt.Sprintf(
					"%s:%s@tcp(%s)/%s?charset=utf8",
					viper.Get("DB_USER"),
					viper.Get("DB_PASS"),
					host,
					viper.Get("DB_NAME"),
				),
			)
			if err != nil {
				log.Err(err).Msgf("failed to init slave database")
				os.Exit(1)
			}
		}
	}

	globalSessions, _ = session.NewManager("memory",
		&session.ManagerConfig{
			CookieName:      "gosessionid",
			EnableSetCookie: true,
			Gclifetime:      3600 * 10,
			Maxlifetime:     3600 * 10,
			DisableHTTPOnly: false,
			Secure:          false,
			CookieLifeTime:  3600 * 10,
		})
	go globalSessions.GC()
}

func main() {
	var FilterUser = func(ctx *context.Context) {
		if strings.HasPrefix(ctx.Input.URL(), "/login") {
			return
		}
		// if strings.HasPrefix(ctx.Input.URL(), "/favicon") {
		// 	return
		// }

		logged := ctx.Input.Session("user")
		if logged == nil && strings.HasPrefix(ctx.Input.URL(), "/registration") {
			return
		}

		if logged == nil {
			ctx.Redirect(http.StatusFound, "/login")
			return
		}

		log.Debug().Msgf("user is logged and go to %s", ctx.Input.URL())
	}

	beego.InsertFilter("/*", beego.BeforeRouter, FilterUser)
	beego.Run()
}
