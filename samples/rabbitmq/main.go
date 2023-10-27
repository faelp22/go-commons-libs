package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/rabbitmq"
	"github.com/google/uuid"
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
		Mode:      config.DEVELOPER,
		RMQConfig: &config.RMQConfig{},
	}

	rbmq_conn := rabbitmq.New(conf)
	task_service := newTaskService(rbmq_conn, conf)

	done := make(chan bool)
	go task_service.Run()
	log.Printf("Worker Running [Mode: %s], [Version: %s], [Commit: %s]", conf.Mode, VERSION, COMMIT)
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

	log.Println("New MSG received")

	log.Println(string(msg.Body))

	if err := msg.Ack(false); err != nil {
		log.Println("Erro to ACK MSG Status")
	} else {
		log.Println("MSG Status update success")
	}
}

func (ts *task_service) anotherProccess() {
	log.Println("Produzindo dados")

	msg_status := &MessageResponseStatus{
		Id:     uuid.New().String(),
		Status: "START",
	}

	data, err := json.Marshal(msg_status)
	if err != nil {
		log.Println("Failed to parse MessageResponseStatus to JSON")
		log.Println(err)
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
		log.Println("Failed to Producer a MSG")
		log.Println(err)
		log.Println(string(msg.Data))
	}
}

func (ts *task_service) Run() {
	conn, err := ts.rbmq.Connect()
	if err != nil {
		log.Println("Erro to connect in rabbitmq")
		log.Println(err)
		os.Exit(1)
	}

	// -----------------------------------------------------

	errs := conn.CompleteQueueDeclare(filas)
	if len(errs) >= 1 {
		log.Println("Erro to run CompleteQueueDeclare in main")
		for err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	// -----------------------------------------------------

	cc := &rabbitmq.ConsumerConfig{
		Queue: QUEUE_NAME,
	}

	go conn.StartConsumer(cc, ts.consumerCallback)

	for {
		if ts.rbmq.GetConnectStatus() {
			log.Println("Iniciando novo processo interno")
			ts.anotherProccess()
		} else {
			log.Println("Parado sem fazer processos internos")
		}
		time.Sleep(3 * time.Second)
	}
}

// go build samples/rabbitmq/main.go && SRV_RMQ_URI="amqp://admin:supersenha@localhost:5672/" ./main.exe
