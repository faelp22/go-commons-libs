package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (rbm *rbm_pool) Consumer(queue_name string, callback func(msg *amqp.Delivery)) {

	HOSTNAME := os.Getenv("HOSTNAME")
	if HOSTNAME == "" {
		HOSTNAME = "bot-user-session-control"
	}

	msgs, err := rbm.channel.Consume(
		queue_name, // queue
		HOSTNAME,   // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
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
