package main

import (
	"app/models"
	"app/service"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/brianvoe/gofakeit"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrIncorrectGoroutinesCount = errors.New("incorrect number of goroutines")

var loginStartIndex uint64

// Run starts tasks in n goroutines.
func Run(n int) (<-chan struct{}, error) {
	if n <= 0 {
		return nil, ErrIncorrectGoroutinesCount
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	results := make(chan error)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					select {
					case results <- create(atomic.AddUint64(&loginStartIndex, 1)):
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	var errCount int
	for result := range results {
		if result != nil {
			errCount++
			log.Err(result).Msgf("error count %d", errCount)
		}

	}
	return ctx.Done(), nil
}

func create(i uint64) error {
	u := models.User{
		Login: fmt.Sprintf("u%d", i),
		Profile: models.Profile{
			Name:    gofakeit.FirstName(),
			Surname: gofakeit.LastName(),
			BirthDate: gofakeit.DateRange(
				time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Now().AddDate(-8, 0, 0),
			),
			City: gofakeit.City(),
			Sex:  models.Sex{Id: gofakeit.Number(1, 3)},
		},
	}
	_, err := service.CreateUser(u, fmt.Sprintf("u%d", i))
	return err
}

func main() {
	flag.Uint64Var(&loginStartIndex, "startIndex", 0, "")
	workers := flag.Int("workersCount", 10, "")
	flag.Parse()

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
	if loginStartIndex == 0 {
		log.Error().Msgf("incorrect login start index")
		return
	}

	done, _ := Run(*workers)
	<-done
}
