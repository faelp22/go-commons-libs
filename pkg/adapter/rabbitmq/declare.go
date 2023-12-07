package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	Name       string     // name
	Durable    bool       // durable
	AutoDelete bool       // delete when unused
	Exclusive  bool       // exclusive
	NoWait     bool       // no-wait
	Arguments  amqp.Table // arguments
	Binds      *[]Bind    // bind to exchange and route with queue bind
}

type Bind struct {
	ExchangeName string
	BindingKey   string
}

type Exchange struct {
	Name       string     // name
	Kind       string     // kind of exchange. ex: 'direct' | 'topic' | 'fanout'
	Durable    bool       // durable
	AutoDelete bool       // delete when unused
	Internal   bool       // internal exchange
	NoWait     bool       // no-wait
	Arguments  amqp.Table // arguments
}

func (rbm *Rbm_pool) SimpleQueueDeclare(sq Queue) (queue amqp.Queue, err error) {
	queue, err = rbm.channel.QueueDeclare(
		sq.Name,       // name
		sq.Durable,    // durable
		sq.AutoDelete, // delete when unused
		sq.Exclusive,  // exclusive
		sq.NoWait,     // no-wait
		sq.Arguments,  // arguments
	)

	if err != nil {
		log.Println("Erro to QueueDeclare Queue in RabbitMQ")
		return queue, err
	}

	return queue, nil
}

func (rbm *Rbm_pool) CompleteQueueDeclare(cq []Queue) []error {
	var listErrors []error
	for _, queue := range cq {
		if _, err := rbm.channel.QueueDeclare(
			queue.Name,       // name
			queue.Durable,    // durable
			queue.AutoDelete, // delete when unused
			queue.Exclusive,  // exclusive
			queue.NoWait,     // no-wait
			queue.Arguments,  // arguments
		); err != nil {
			log.Println("Erro to QueueDeclare Queue in RabbitMQ")
			listErrors = append(listErrors, err)
		}

		if queue.Binds != nil {
			for _, bind := range *queue.Binds {
				if err := rbm.channel.QueueBind(
					queue.Name,
					bind.BindingKey,
					bind.ExchangeName,
					queue.NoWait,
					queue.Arguments,
				); err != nil {
					log.Println("Erro to QueueBind in RabbitMQ")
					listErrors = append(listErrors, err)
				}
			}
		}
	}

	return listErrors
}

func (rbm *Rbm_pool) SimpleExchangeDeclare(se Exchange) error {
	if err := rbm.channel.ExchangeDeclare(
		se.Name,       // name
		se.Kind,       // kind of exchange. ex: 'direct' | 'topic' | 'fanout'
		se.Durable,    // durable
		se.AutoDelete, // delete when unused
		se.Internal,   // internal exchange
		se.NoWait,     // no-wait
		se.Arguments,  // arguments
	); err != nil {
		log.Println("Erro to ExchangeDeclare in RabbitMQ")
		return err
	}

	return nil
}

func (rbm *Rbm_pool) CompleteExchangeDeclare(ce []Exchange) []error {
	var listErrors []error
	for _, exchange := range ce {
		if err := rbm.channel.ExchangeDeclare(
			exchange.Name,       // name
			exchange.Kind,       // kind of exchange. ex: 'direct' | 'topic' | 'fanout'
			exchange.Durable,    // durable
			exchange.AutoDelete, // delete when unused
			exchange.Internal,   // internal exchange
			exchange.NoWait,     // no-wait
			exchange.Arguments,  // arguments
		); err != nil {
			log.Println("Erro to ExchangeDeclare in RabbitMQ")
			listErrors = append(listErrors, err)
		}
	}

	return listErrors
}

func (rbm *Rbm_pool) CompleteDeclare(cq []Queue, ce []Exchange) []error {
	var listErrors []error
	if err := rbm.CompleteExchangeDeclare(ce); err != nil {
		listErrors = append(listErrors, err...)
	}

	if err := rbm.CompleteQueueDeclare(cq); err != nil {
		listErrors = append(listErrors, err...)
	}

	return listErrors
}
