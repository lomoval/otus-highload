package main

import (
	internalgrpc "app/internal/server/grpc"
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	consul "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	configFile           string
	server               *internalgrpc.Dialogs
	dialogServerConsulId = ""
)

const serviceName = "dialogs"

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

	viper.BindEnv("CONSUL_ADDRESS")
	viper.BindEnv("DIALOGS_CONSUL_HOST")
	viper.BindEnv("DIALOGS_CONSUL_PORT")

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

	dialogServerConsulId = "dialogs" + strconv.Itoa(viper.GetInt("DIALOGS_CONSUL_PORT"))
	server = internalgrpc.NewServer(viper.GetString("DIALOGS_HOST"), viper.GetInt("DIALOGS_PORT"), viper.GetString("ZIPKIN_URL"))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	cfg := consul.DefaultConfig()
	cfg.Address = viper.GetString("CONSUL_ADDRESS")
	c, err := consul.NewClient(cfg)
	if err != nil {
		panic(err.Error())
	}
	consulAgent := c.Agent()
	serviceDef := &consul.AgentServiceRegistration{
		ID:      dialogServerConsulId,
		Name:    serviceName,
		Port:    viper.GetInt("DIALOGS_CONSUL_PORT"),
		Address: viper.GetString("DIALOGS_CONSUL_HOST"),
		Check: &consul.AgentServiceCheck{
			TTL: (10 * time.Second).String(),
		},
	}

	if err := consulAgent.ServiceRegister(serviceDef); err != nil {
		panic(err.Error())
	}

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				consulAgent.UpdateTTL("service:dialogs"+strconv.Itoa(viper.GetInt("DIALOGS_PORT")), "", "pass")
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		<-ctx.Done()
		server.Stop()
	}()
	server.Start(context.Background())
}
