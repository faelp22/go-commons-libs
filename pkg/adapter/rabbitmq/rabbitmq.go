package rabbitmq

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/faelp22/go-commons-libs/core/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const DEFAULT_MAXX_RECONNECT_TIMES = 3

type RabbitInterface interface {
	Connect() (RabbitInterface, error)
	GetConnect() *rbm_pool
	SimpleQueueDeclare(sq SimpleQueue) (queue amqp.Queue, err error)
	Producer(ctx context.Context, pc *ProducerConfig, msg *Message) error
	Consumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery))
	StartConsumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery))
}

type SimpleQueue struct {
	Name       string     // name
	Durable    bool       // durable
	AutoDelete bool       // delete when unused
	Exclusive  bool       // exclusive
	NoWait     bool       // no-wait
	Arguments  amqp.Table // arguments
}

type rbm_pool struct {
	conn                 *amqp.Connection
	channel              *amqp.Channel
	conf                 *config.Config
	err                  chan error
	MAXX_RECONNECT_TIMES int
}

var rbmpool = &rbm_pool{
	err: make(chan error),
}

func New(conf *config.Config) RabbitInterface {

	SRV_RMQ_URI := os.Getenv("SRV_RMQ_URI")
	if SRV_RMQ_URI != "" {
		conf.RMQ_URI = SRV_RMQ_URI
	} else {
		log.Println("A variável SRV_RMQ_URI é obrigatória!")
		os.Exit(1)
	}

	SRV_RMQ_MAXX_RECONNECT_TIMES := os.Getenv("SRV_RMQ_MAXX_RECONNECT_TIMES")
	if SRV_RMQ_MAXX_RECONNECT_TIMES != "" {
		conf.RMQ_MAXX_RECONNECT_TIMES, _ = strconv.Atoi(SRV_RMQ_MAXX_RECONNECT_TIMES)
	} else {
		conf.RMQ_MAXX_RECONNECT_TIMES = DEFAULT_MAXX_RECONNECT_TIMES
	}

	rbmpool = &rbm_pool{
		conf: conf,
		err:  make(chan error),
	}
	return rbmpool
}

func (rbm *rbm_pool) Connect() (RabbitInterface, error) {

	var err error

	rbm.conn, err = amqp.Dial(rbm.conf.RMQ_URI)
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ")
		return rbm, err
	}

	go func() {
		<-rbm.conn.NotifyClose(make(chan *amqp.Error)) // Listen to Connection NotifyClose
		rbm.err <- errors.New("connection closed")
	}()

	rbm.channel, err = rbm.conn.Channel()
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ Channel")
		return rbm, err
	}

	go func() {
		<-rbm.channel.NotifyClose(make(chan *amqp.Error)) // Listen to Channel NotifyClose
		rbm.err <- errors.New("channel closed")
	}()

	log.Println("New RabbitMQ Connect Success")

	return rbm, nil
}

func (rbm *rbm_pool) GetConnect() *rbm_pool {
	return rbm
}

func (rbm *rbm_pool) SimpleQueueDeclare(sq SimpleQueue) (queue amqp.Queue, err error) {
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
