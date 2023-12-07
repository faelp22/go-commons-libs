package rabbitmq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var OldConsumerName string

type ConsumerConfig struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

func (rbm *Rbm_pool) Consumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {

	if cc.Consumer == "" {
		cc.Consumer = fmt.Sprintf("worker-read-msg@%s", uuid.New().String()[:8])
	} else if cc.Consumer == OldConsumerName {
		name := strings.Split(OldConsumerName, "@")[0]
		cc.Consumer = fmt.Sprintf("%s@%s", name, uuid.New().String()[:8])
	} else {
		cc.Consumer = fmt.Sprintf("%s@%s", cc.Consumer, uuid.New().String()[:8])
	}

	OldConsumerName = cc.Consumer

	msgs, err := rbm.channel.Consume(
		cc.Queue,     // queue
		cc.Consumer,  // consumer
		cc.AutoAck,   // auto-ack
		cc.Exclusive, // exclusive
		cc.NoLocal,   // no-local
		cc.NoWait,    // no-wait
		cc.Args,      // args
	)

	if err != nil {
		log.Println("Failed to register a consumer")
		log.Println(err)
		return
	}

	go func() {
		log.Println("Start Consumer")
		for msg := range msgs {
			callback(&msg)
		}
		log.Println("Close Consumer")
	}()
}

func (rbm *Rbm_pool) StartConsumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {
	count := 0
	for {

		if rbm.connStatus {
			go rbm.Consumer(cc, callback)
		}

		if count >= rbm.conf.RMQ_MAXX_RECONNECT_TIMES {
			log.Println("Erro to reconnect 3 times in RabbitMQ")
			os.Exit(1)
		}

		if err := <-rbm.err; err != nil {

			log.Println("Connection is closed, trying to reconnect in RabbitMQ")

			rb_conn, err := rbm.Connect()
			if err != nil {
				go func() { rbm.err <- errors.New("connection closed re trying") }()
				count++
				log.Println("Waiting 30 seconds to try again")
				time.Sleep(30 * time.Second) // wait 30 seconds
			} else {
				count = 0
				rbm.conn = rb_conn.GetConnect().conn
				rbm.connStatus = rb_conn.GetConnect().connStatus
			}
		}
	}
}
