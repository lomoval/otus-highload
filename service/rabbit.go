package service

import (
	"app/models"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"math/rand"
	"strconv"
	"time"
)

var (
	rabbitConn             *amqp.Connection
	rabbitCh               *amqp.Channel
	rabbitUrl              string
	rabbitConsumerConn     *amqp.Connection
	rabbitConsumerCh       *amqp.Channel
	rabbitUserNewsDelivery map[int64]<-chan amqp.Delivery
	rabbitQueueNamePrefix  = "news-"
)

func StartRabbit(url string) error {
	rand.Seed(time.Now().UnixNano())
	rabbitQueueNamePrefix += strconv.Itoa(rand.Int()) + "-"

	rabbitUserNewsDelivery = make(map[int64]<-chan amqp.Delivery)
	var err error
	rabbitUrl = url
	rabbitConn, err = amqp.Dial(url)

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		return err
	}

	rabbitCh.ExchangeDeclare("user.news", "direct", false, false, false, false, nil)

	if err != nil {
		return err
	}

	return err
}

func StopRabbit() {
	if rabbitConn != nil && !rabbitConn.IsClosed() {
		rabbitConn.Close()
	}
}

func SendRabbitNewsMessage(userID int64, news models.News) {
	var err error
	if rabbitConn == nil || rabbitConn.IsClosed() {
		rabbitConn, err = amqp.Dial(rabbitUrl) // "amqp://guest:guest@localhost:5672/")
		if err != nil {
			log.Err(err).Msgf("failed to dial rabbit`")
			return
		}
		rabbitCh, err = rabbitConn.Channel()
		if err != nil {
			log.Err(err).Msgf("failed to send rabbit message")
			return
		}
	}

	b, err := json.Marshal(news)
	if err != nil {
		log.Err(err).Msgf("failed to send rabbit message")
		return
	}

	friendsIDs, err := FriendsIDs(userID)
	if err != nil {
		log.Err(err).Msgf("failed to get user ids")
		return
	}

	for _, id := range friendsIDs {
		err = rabbitCh.Publish(
			"user.news",               // exchange
			strconv.FormatInt(id, 10), // routing key
			false,                     // mandatory
			false,                     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        b,
			})

		if err != nil {
			log.Err(err).Msgf("failed to publish message")
			return
		}
	}
}

func GetRabbitNews(userID int64) ([]models.News, error) {
	delivery, err := consumerDelivery(userID)
	if err != nil {
		return nil, err
	}

	var news []models.News
	for {
		select {
		case msg, ok := <-delivery:
			if !ok {
				deleteDelivery(userID)
				return nil, nil
			}
			var n models.News
			json.Unmarshal(msg.Body, &n)
			news = append(news, n)
		default:
			return news, nil
		}
	}
	return nil, nil
}

func consumerDelivery(userID int64) (<-chan amqp.Delivery, error) {
	var err error
	if rabbitConsumerConn == nil || rabbitConsumerConn.IsClosed() {
		rabbitConsumerCh = nil
		rabbitConsumerConn, err = amqp.Dial(rabbitUrl)
		if rabbitConsumerCh != nil {
			rabbitConsumerCh.Close()
		}
	}
	if err != nil {
		return nil, err
	}

	if rabbitConsumerCh == nil {
		rabbitConsumerCh, err = rabbitConsumerConn.Channel()
		if err != nil {
			return nil, err
		}
	}

	if del, ok := rabbitUserNewsDelivery[userID]; ok {
		return del, nil
	}

	rabbitQueue, err := rabbitConsumerCh.QueueDeclare(
		fmt.Sprintf("%s%d", rabbitQueueNamePrefix, userID),
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("bind queue for %d", userID)
	err = rabbitConsumerCh.QueueBind(
		fmt.Sprintf("%s%d", rabbitQueueNamePrefix, userID),
		strconv.FormatInt(userID, 10),
		"user.news",
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	delivery, err := rabbitConsumerCh.Consume(
		rabbitQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	addDelivery(userID, delivery)
	return delivery, nil
}

func addDelivery(userID int64, delivery <-chan amqp.Delivery) {
	rabbitUserNewsDelivery[userID] = delivery
}

func deleteDelivery(userID int64) {
	delete(rabbitUserNewsDelivery, userID)
}

func RemoveRabbitUser(userID int64) {
	delete(rabbitUserNewsDelivery, userID)
}
