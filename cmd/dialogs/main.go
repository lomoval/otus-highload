package main

import (
	internalgrpc "app/internal/server/grpc"
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/spf13/viper"
	"os/signal"
	"syscall"
)

var configFile string

var server *internalgrpc.Dialogs

func init() {
	viper.SetEnvPrefix("OTUS_HIGHLOAD")
	viper.BindEnv("DIALOGS_HOST")
	viper.BindEnv("DIALOGS_PORT")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASS")

	viper.SetDefault("DIALOGS_HOST", "0.0.0.0")
	viper.SetDefault("DIALOGS_PORT", 8005)
	viper.BindEnv("ZIPKIN_URL")

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

	server = internalgrpc.NewServer(viper.GetString("DIALOGS_HOST"), viper.GetInt("DIALOGS_PORT"), viper.GetString("ZIPKIN_URL"))
}

func main() {
	server.Start(context.Background())
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		server.Stop()
	}()
}
