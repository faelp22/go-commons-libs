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

var filas []rabbitmq.Queue = []rabbitmq.Queue{
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

	rbmq_conn := rabbitmq.New(conf)
	task_service := newTaskService(rbmq_conn, conf)

	done := make(chan bool)
	go task_service.Run()
	log.Info().Str("Mode", conf.AppMode).Str("Version", conf.AppVersion).Str("Commit", conf.AppCommitShortSha).Msg("Worker Running")
	<-done
}

type task_service struct {
	rbmq rabbitmq.RabbitInterface
	conf *config.Config
}

func newTaskService(rbmq rabbitmq.RabbitInterface, conf *config.Config) *task_service {
	return &task_service{
		rbmq: rbmq,
		conf: conf,
	}
}

func (ts *task_service) consumerCallback(msg *amqp.Delivery) {

	log.Debug().Str("msgBody", string(msg.Body)).Msg("New MSG received")

	if err := msg.Ack(false); err != nil {
		log.Error().Str("FunctionName", "consumerCallback").Msg("Erro to ACK MSG Status")
	} else {
		log.Debug().Msg("MSG Status update success")
	}
}

func (ts *task_service) anotherProccess() {
	log.Debug().Msg("Produzindo dados")

	msg_status := &MessageResponseStatus{
		Id:     uuid.New().String(),
		Status: "START",
	}

	data, err := json.Marshal(msg_status)
	if err != nil {
		log.Error().Str("ERRO_TEST_RMQ", "Failed to parse MessageResponseStatus to JSON").Msg(err.Error())
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
		log.Error().Str("ERRO_TEST_RMQ", "Failed to Producer a MSG").Str("data", string(msg.Data)).Msg(err.Error())
	}
}

func (ts *task_service) Run() {
	conn, err := ts.rbmq.Connect()
	if err != nil {
		log.Fatal().Str("ERRO_TEST_RMQ", "Erro to connect in rabbitmq").Msg(err.Error())
	}

	// -----------------------------------------------------

	errs := conn.CompleteQueueDeclare(filas)
	if len(errs) >= 1 {
		for _, err := range errs {
			log.Error().Msg(err.Error())
		}
		log.Fatal().Msg("Erro to run CompleteQueueDeclare in main")
	}

	// -----------------------------------------------------

	cc := &rabbitmq.ConsumerConfig{
		Queue: QUEUE_NAME,
	}

	go conn.StartConsumer(cc, ts.consumerCallback)

	for {
		if ts.rbmq.GetConnectStatus() {
			log.Debug().Msg("Iniciando novo processo interno")
			ts.anotherProccess()
		} else {
			log.Debug().Msg("Parado sem fazer processos internos")
		}
		time.Sleep(3 * time.Second)
	}
}

// go build samples/rabbitmq/main.go && SRV_RMQ_URI="amqp://admin:supersenha@localhost:5672/" ./main.exe
