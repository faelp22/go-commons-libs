package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	Data        []byte
	ContentType string
}

type ProducerConfig struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
}

func (rbm *rbm_pool) Producer(ctx context.Context, pc *ProducerConfig, msg *Message) error {

	err := rbm.channel.PublishWithContext(ctx,
		pc.Exchange,  // exchange
		pc.Key,       // routing key
		pc.Mandatory, // mandatory
		pc.Immediate, // immediate
		amqp.Publishing{
			Body:        msg.Data,
			ContentType: msg.ContentType,
		})

	if err != nil {
		log.Println(err)
	}

	return err
}
