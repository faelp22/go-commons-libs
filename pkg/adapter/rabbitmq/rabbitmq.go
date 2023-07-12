package rabbitmq

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitInterface interface {
	Publish(ctx context.Context, queue_name string, msg *Message) error
	Consumer(queue_name string, callback func(msg *amqp.Delivery))
	Connect() error
	Start(queue_name string, callback func(msg *amqp.Delivery))
}

type Message struct {
	Data        []byte
	ContentType string
}

type Fila struct {
	Name       string     // name
	Durable    bool       // durable
	AutoDelete bool       // delete when unused
	Exclusive  bool       // exclusive
	NoWait     bool       // no-wait
	Arguments  amqp.Table // arguments
}

type rbm_pool struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queues  []Fila
	err     chan error
	conf    *config.Config
}

var rbmpool = &rbm_pool{
	err: make(chan error),
}

func New(lista_filas []Fila, conf *config.Config) RabbitInterface {
	rbmpool = &rbm_pool{
		queues: lista_filas,
		conf:   conf,
		err:    make(chan error),
	}
	return rbmpool
}

func (rbm *rbm_pool) Connect() error {

	var err error

	rbm.conn, err = amqp.Dial(rbm.conf.RMQ_URI)
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ")
		return err
	}

	go func() {
		<-rbm.conn.NotifyClose(make(chan *amqp.Error)) //Listen to NotifyClose
		rbm.err <- errors.New("connection closed")
	}()

	rbm.channel, err = rbm.conn.Channel()
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ Channel")
		return err
	}

	for _, fl := range rbm.queues {
		_, err = rbm.channel.QueueDeclare(
			fl.Name,       // name
			fl.Durable,    // durable
			fl.AutoDelete, // delete when unused
			fl.Exclusive,  // exclusive
			fl.NoWait,     // no-wait
			fl.Arguments,  // arguments
		)
		if err != nil {
			log.Println("Erro to QueueDeclare Queue in RabbitMQ")
			return err
		}
	}

	log.Println("New RabbitMQ Connect Success")
	return nil
}

func (rbm *rbm_pool) Start(queue_name string, callback func(msg *amqp.Delivery)) {
	isClosed := false
	count := 0
	MAXX_RECONNECT_TIMES := 3
	for {

		if !isClosed {
			go rbm.Consumer(queue_name, callback)
		}

		if count >= MAXX_RECONNECT_TIMES {
			log.Println("Erro to reconnect 3 times in RabbitMQ")
			os.Exit(1)
		}

		if err := <-rbm.err; err != nil {
			if !isClosed {
				log.Println("Connection is closed, trying to reconnect in RabbitMQ")
			}
			err2 := rbm.Connect()
			if err2 != nil {
				go func() { rbm.err <- errors.New("connection closed") }()
				count++
				isClosed = true
				log.Println("Waiting 30 seconds to try again")
				time.Sleep(time.Duration(30) * time.Second) // wait 30 seconds
			} else {
				count = 0
				isClosed = false
			}
		}
	}
}
