package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (rbm *rbm_pool) Publish(ctx context.Context, queue_name string, msg *Message) error {

	err := rbm.channel.PublishWithContext(ctx,
		"",         // exchange
		queue_name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Body:        msg.Data,
			ContentType: msg.ContentType,
		})

	if err != nil {
		log.Println(err)
	}

	return err
}
