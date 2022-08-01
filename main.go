//go:generate -command PROTOC protoc -I./api/proto ./api/proto/dialog.proto ./api/proto/dialog-service.proto ./api/proto/counter-service.proto
//go:generate PROTOC --go_out=./api/ --go-grpc_out=./api/
// :generate PROTOC --grpc-gateway_out ../../../api/ --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true

package main

import (
	"app/models"
	_ "app/routers"
	"app/service"
	"encoding/gob"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strconv"
	"time"

	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/beego/beego/v2/server/web/session/mysql"
)

var globalSessions *session.Manager
var zipkinTracer *zipkin.Tracer

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
	viper.BindEnv("KAFKA_BOOTSTRAPS_SERVERS")
	viper.BindEnv("TARANTOOL_SERVER")
	viper.BindEnv("TARANTOOL_SERVER_USER")
	viper.BindEnv("TARANTOOL_SERVER_PASS")
	viper.BindEnv("RABBIT_URL")
	viper.BindEnv("DIALOGS_HOST")
	viper.BindEnv("DIALOGS_PORT")
	viper.BindEnv("ZIPKIN_URL")
	viper.BindEnv("CONSUL_DIALOGS_URI")
	viper.BindEnv("COUNTERS_HOST")
	viper.BindEnv("COUNTERS_PORT")

	err := service.SetupTarantool(
		viper.GetString("TARANTOOL_SERVER"),
		viper.GetString("TARANTOOL_SERVER_USER"),
		viper.GetString("TARANTOOL_SERVER_PASS"),
	)

	if err != nil {
		log.Err(err).Msgf("failed to init tarantool")
		os.Exit(1)
	}

	gob.Register(models.User{})
	web.BConfig.WebConfig.Session.SessionProvider = "mysql"
	web.BConfig.WebConfig.Session.SessionProviderConfig = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8",
		viper.Get("DB_USER"),
		viper.Get("DB_PASS"),
		viper.Get("DB_HOST"),
		viper.Get("DB_PORT"),
		viper.Get("DB_NAME"),
	)

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

	reporter := zipkinreporter.NewReporter(viper.GetString("ZIPKIN_URL"))
	zipkinTracer, err = zipkin.NewTracer(reporter)
	if err != nil {
		log.Err(err).Msgf("failed to init zipkin")
		os.Exit(1)
	}
	err = service.SetupGrpcDialogs(viper.GetString("CONSUL_DIALOGS_URI"), zipkinTracer)
	if err != nil {
		log.Err(err).Msgf("failed to init dialogs service")
		os.Exit(1)
	}
	err = service.StartNewsProducer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news producer")
		os.Exit(1)
	}
	err = service.StartNewsConsumer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news consumer")
		os.Exit(1)
	}
	err = service.StartRabbit(viper.GetString("RABBIT_URL"))
	if err != nil {
		log.Err(err).Msgf("failed to init Rabbit news consumer")
		os.Exit(1)
	}

	err = service.SetupGrpcCounters(viper.GetString("COUNTERS_HOST") + ":" + viper.GetString("COUNTERS_PORT"))
	if err != nil {
		log.Err(err).Msgf("failed to init dialogs service")
		os.Exit(1)
	}

	err = service.StartPrivateMessageProducer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news producer")
		os.Exit(1)
	}
	err = service.StartPrivateMessageConfirmationConsumer(viper.GetString("KAFKA_BOOTSTRAPS_SERVERS"))
	if err != nil {
		log.Err(err).Msgf("failed to init Kafka news consumer")
		os.Exit(1)
	}

}

func main() {
	defer service.StopProducers()
	defer service.ShutdownTarantool()
	defer service.StopRabbit()
	defer service.StopGrpcDialogs()
	defer service.StopGrpcCounter()

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

		service.SetActiveUser(logged.(models.User).Id, time.Hour*10)
		log.Debug().Msgf("user is logged and go to %s", ctx.Input.URL())
	}

	beego.InsertFilter("/*", beego.BeforeRouter, FilterUser)

	serverMiddleware := zipkinhttp.NewServerMiddleware(
		zipkinTracer,
		zipkinhttp.TagResponseSize(true),
		zipkinhttp.SpanName("http-request"),
	)
	beego.RunWithMiddleWares("", serverMiddleware)
}
