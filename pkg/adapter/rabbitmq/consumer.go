package rabbitmq

import (
	"errors"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerConfig struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

func (rbm *rbm_pool) Consumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {
	HOSTNAME := os.Getenv("HOSTNAME")
	if HOSTNAME == "" {
		cc.Consumer = "worker-read-msg"
	}

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
	}

	go func() {
		log.Println("Start Consumer")
		for msg := range msgs {
			callback(&msg)
		}
		log.Println("Close Consumer")
	}()
}

func (rbm *rbm_pool) StartConsumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {
	isClosed := false
	count := 0
	for {

		if !isClosed {
			go rbm.Consumer(cc, callback)
		}

		if count >= rbm.conf.RMQ_MAXX_RECONNECT_TIMES {
			log.Println("Erro to reconnect 3 times in RabbitMQ")
			os.Exit(1)
		}

		if err := <-rbm.err; err != nil {
			if !isClosed {
				log.Println("Connection is closed, trying to reconnect in RabbitMQ")
			}

			rb_conn, err2 := rbm.Connect()
			if err2 != nil {
				go func() { rbm.err <- errors.New("connection closed") }()
				count++
				isClosed = true
				log.Println("Waiting 30 seconds to try again")
				time.Sleep(time.Duration(30) * time.Second) // wait 30 seconds
			} else {
				count = 0
				isClosed = false
				rbm.conn = rb_conn.GetConnect().conn
			}
		}
	}
}
