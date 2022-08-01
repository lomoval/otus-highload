package main

import (
	internalgrpc "app/internal/server/grpc"
	"app/service"
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	configFile           string
	server               *internalgrpc.Counter
	dialogServerConsulId = ""
)

const serviceName = "dialogs"

func init() {
	viper.SetEnvPrefix("OTUS_HIGHLOAD")
	viper.BindEnv("COUNTER_HOST")
	viper.BindEnv("COUNTER_PORT")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASS")
	viper.BindEnv("KAFKA_BOOTSTRAPS_SERVERS")

	viper.SetDefault("COUNTER_HOST", "0.0.0.0")
	viper.SetDefault("COUNTER_PORT", 8006)

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
	server = internalgrpc.NewCounterServer(viper.GetString("COUNTER_HOST"), viper.GetInt("COUNTER_PORT"))

	err := service.StartPrivateMessageConfirmationProducer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news producer")
		os.Exit(1)
	}
	err = service.StartPrivateMessageConsumer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news consumer")
		os.Exit(1)
	}
}

func main() {
	defer service.StopProducers()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		server.Stop()
	}()
	server.Start(context.Background())
}
