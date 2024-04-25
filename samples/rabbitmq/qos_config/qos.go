package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/rabbitmq"
	"github.com/google/uuid"
	"github.com/phuslu/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	VERSION = "0.1.0-dev"
	COMMIT  = "ABCDEFG-dev"
)

const (
	QUEUE_NAME   = "TESTE_SAMPLE_QUEUE"
	CONTENT_TYPE = "application/json; charset=utf-8"
)

type MessageResponseStatus struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

var queues []rabbitmq.Queue = []rabbitmq.Queue{
	{
		Name:       QUEUE_NAME,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
	},
}

func main() {
	conf := &config.Config{
		AppMode:         config.DEVELOPER,
		AppTargetDeploy: config.TARGET_DEPLOY_LOCAL,
		RMQConfig:       &config.RMQConfig{},
	}

	rbmqConn := rabbitmq.New(conf)
	taskService := newTaskService(rbmqConn, conf)

	done := make(chan bool)
	go taskService.Run()
	log.Info().Str("Mode", conf.AppMode).Str("Version", conf.AppVersion).Str("Commit", conf.AppCommitShortSha).Msg("Worker Running")
	<-done
}

type taskService struct {
	rbmq rabbitmq.RabbitInterface
	conf *config.Config
}

func newTaskService(rbmq rabbitmq.RabbitInterface, conf *config.Config) *taskService {
	return &taskService{
		rbmq: rbmq,
		conf: conf,
	}
}

func (ts *taskService) consumerCallback(msg *amqp.Delivery) {
	log.Debug().Str("msgBody", string(msg.Body)).Msg("New message received")

	if err := msg.Ack(false); err != nil {
		log.Error().Str("FunctionName", "consumerCallback").Msg("Error acknowledging message status")
	} else {
		log.Debug().Msg("Message status updated successfully")
	}
}

func (ts *taskService) anotherProcess() {
	log.Debug().Msg("Producing data")

	msgStatus := &MessageResponseStatus{
		Id:     uuid.New().String(),
		Status: "START",
	}

	data, err := json.Marshal(msgStatus)
	if err != nil {
		log.Error().Str("ERROR_TEST_RMQ", "Failed to parse MessageResponseStatus to JSON").Msg(err.Error())
	}

	msg := &rabbitmq.Message{
		Data:        data,
		ContentType: CONTENT_TYPE,
	}

	pc := &rabbitmq.ProducerConfig{
		Key: QUEUE_NAME,
	}

	err = ts.rbmq.Producer(context.Background(), pc, msg)
	if err != nil {
		log.Error().Str("ERROR_TEST_RMQ", "Failed to produce a message").Str("data", string(msg.Data)).Msg(err.Error())
	}
}

func (ts *taskService) Run() {
	conn, err := ts.rbmq.Connect()
	if err != nil {
		log.Fatal().Str("ERROR_TEST_RMQ", "Failed to connect to RabbitMQ").Msg(err.Error())
	}

	// -----------------------------------------------------

	errs := conn.CompleteQueueDeclare(queues)
	if len(errs) >= 1 {
		for _, err := range errs {
			log.Error().Msg(err.Error())
		}
		log.Fatal().Msg("Failed to run CompleteQueueDeclare in main")
	}

	// -----------------------------------------------------

	cc := &rabbitmq.ConsumerConfig{
		Queue: QUEUE_NAME,
		ControlQosConfig: rabbitmq.ControlQosConfig{
			PrefetchCount: 1,
			PrefetchSize:  0,
			Global:        false,
		},
	}

	go conn.StartConsumer(cc, ts.consumerCallback)

	for {
		if ts.rbmq.GetConnectStatus() {
			log.Debug().Msg("Starting new internal process")
			ts.anotherProcess()
		} else {
			log.Debug().Msg("Stopped without running internal processes")
		}
		time.Sleep(3 * time.Second)
	}
}
