package service

import (
	"app/models"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
)

const (
	newsTopic                  = "user.news"
	privateMessageTopic        = "user.private-message"
	privateMessageConfirmTopic = "user.private-message-confirm"
)

var newsProducer *kafka.Producer
var privateMessageProducer *kafka.Producer
var privateMessageConfirmationProducer *kafka.Producer
var done chan struct{}

func StartNewsProducer(bootStrapServers string) error {
	var err error
	newsProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootStrapServers})
	if err != nil {
		return err
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range newsProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Error().Msgf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Debug().Msgf("Delivered message to %v\n", ev.TopicPartition)
				}
			case kafka.Error:
				log.Error().Msgf("kafka producer error %v", e)
			}
		}
	}()
	return nil
}

func StartPrivateMessageProducer(bootStrapServers string) error {
	var err error
	privateMessageProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootStrapServers})
	if err != nil {
		return err
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range privateMessageProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Error().Msgf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Debug().Msgf("Delivered message to %v\n", ev.TopicPartition)
				}
			case kafka.Error:
				log.Error().Msgf("kafka producer error %v", e)
			}
		}
	}()
	return nil
}

func StartPrivateMessageConfirmationProducer(bootStrapServers string) error {
	var err error
	privateMessageConfirmationProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootStrapServers})
	if err != nil {
		return err
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range privateMessageConfirmationProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Error().Msgf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Debug().Msgf("Delivered message to %v\n", ev.TopicPartition)
				}
			case kafka.Error:
				log.Error().Msgf("kafka producer error %v", e)
			}
		}
	}()
	return nil
}

func StartNewsConsumer(bootstrapServer string) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id":          "group.news",
		"auto.offset.reset": "earliest"})

	if err != nil {
		return err
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics([]string{newsTopic}, nil)

	done = make(chan struct{})
	go func() {
		defer c.Close()
		for {
			select {
			case <-done:
				return
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					dec := gob.NewDecoder(bytes.NewBuffer(e.Value))
					var news models.News
					err := dec.Decode(&news)
					if err != nil {
						log.Err(err).Msgf("failed to decode Kafka message")
					}
					log.Debug().Msgf("news messages: %v", news)
					err = AddFriendsNews(news)
					if err != nil {
						log.Err(err).Msgf("failed to add friends news")
					}
				case kafka.Error:
					log.Error().Msgf("kafka error %v", e)
					if e.Code() == kafka.ErrAllBrokersDown {
						return
					}
				default:
					log.Debug().Msgf("Ignored %v\n", e)
				}
			}
		}
	}()
	return nil
}

func StopProducers() {
	close(done)
	newsProducer.Flush(15 * 1000)
	newsProducer.Close()
	privateMessageProducer.Flush(15 * 1000)
	privateMessageProducer.Close()
	privateMessageConfirmationProducer.Flush(15 * 1000)
	privateMessageConfirmationProducer.Close()
}

func SendNewsMessage(news models.News) {
	var message bytes.Buffer
	var key bytes.Buffer
	enc := gob.NewEncoder(&message)
	err := enc.Encode(news)
	if err != nil {
		log.Err(err).Msgf("failed to encode news message")
		return
	}

	keyEnc := gob.NewEncoder(&key) // Will write to network.
	err = keyEnc.Encode(news.Creator.Id)
	if err != nil {
		log.Err(err).Msgf("failed to encode news message")
		return
	}

	t := newsTopic
	newsProducer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &t, Partition: kafka.PartitionAny},
		Value:          message.Bytes(),
		Key:            key.Bytes(),
	}
}

func SendNewPrivateDialogMessage(privateMessage models.PrivateDialogMessage) {
	var message bytes.Buffer
	var key bytes.Buffer
	enc := gob.NewEncoder(&message)
	err := enc.Encode(privateMessage)
	if err != nil {
		log.Err(err).Msgf("failed to encode private dialog message")
		return
	}

	keyEnc := gob.NewEncoder(&key) // Will write to network.
	err = keyEnc.Encode(privateMessage.DialogId)
	if err != nil {
		log.Err(err).Msgf("failed to encode news message")
		return
	}

	t := privateMessageTopic
	privateMessageProducer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &t, Partition: kafka.PartitionAny},
		Value:          message.Bytes(),
		Key:            key.Bytes(),
	}
}

func SendPrivateMessageConfirmation(privateMessage models.PrivateDialogMessage) {
	var message bytes.Buffer
	var key bytes.Buffer
	enc := gob.NewEncoder(&message)
	err := enc.Encode(privateMessage)
	if err != nil {
		log.Err(err).Msgf("failed to encode private dialog message")
		return
	}

	keyEnc := gob.NewEncoder(&key) // Will write to network.
	err = keyEnc.Encode(privateMessage.DialogId)
	if err != nil {
		log.Err(err).Msgf("failed to encode news message")
		return
	}

	t := privateMessageConfirmTopic
	privateMessageConfirmationProducer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &t, Partition: kafka.PartitionAny},
		Value:          message.Bytes(),
		Key:            key.Bytes(),
	}
}

func StartPrivateMessageConfirmationConsumer(bootstrapServer string) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id":          "group.confirm.private.answer",
		"auto.offset.reset": "earliest"})

	if err != nil {
		return err
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics([]string{privateMessageConfirmTopic}, nil)

	done = make(chan struct{})
	go func() {
		defer c.Close()
		for {
			select {
			case <-done:
				return
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					dec := gob.NewDecoder(bytes.NewBuffer(e.Value))
					var m models.PrivateDialogMessage
					err := dec.Decode(&m)
					if err != nil {
						log.Err(err).Msgf("failed to decode Kafka message")
					}
					log.Debug().Msgf("confirm answer: %v", m)
					err = ConfirmPrivateDialog(m.Id)
					if err != nil {
						log.Err(err).Msgf("failed to confirm message")
					}
				case kafka.Error:
					log.Error().Msgf("kafka error %v", e)
					if e.Code() == kafka.ErrAllBrokersDown {
						return
					}
				default:
					log.Debug().Msgf("Ignored %v\n", e)
				}
			}
		}
	}()
	return nil
}

func StartPrivateMessageConsumer(bootstrapServer string) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id":          "group.private.message",
		"auto.offset.reset": "earliest"})

	if err != nil {
		return err
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics([]string{privateMessageTopic}, nil)

	done = make(chan struct{})
	go func() {
		defer c.Close()
		for {
			select {
			case <-done:
				return
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					dec := gob.NewDecoder(bytes.NewBuffer(e.Value))
					var m models.PrivateDialogMessage
					err := dec.Decode(&m)
					if err != nil {
						log.Err(err).Msgf("failed to decode Kafka message")
					}
					log.Debug().Msgf("confirm answer: %v", m)
					err = IncDialogAnswer(m.DialogId, m.ToUserId, m.Timestamp)
					if err != nil {
						log.Err(err).Msgf("failed inc counter")
					} else {
						SendPrivateMessageConfirmation(m)
					}
				case kafka.Error:
					log.Error().Msgf("kafka error %v", e)
					if e.Code() == kafka.ErrAllBrokersDown {
						return
					}
				default:
					log.Debug().Msgf("Ignored %v\n", e)
				}
			}
		}
	}()
	return nil
}
