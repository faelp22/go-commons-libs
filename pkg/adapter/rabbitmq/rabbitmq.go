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

const DEFAULT_MAX_RECONNECT_TIMES = 3 // 3 attempts

type RabbitInterface interface {
	// Connect creates a new connection and returns RabbitInterface to access functions and error
	Connect() (RabbitInterface, error)
	// GetConnect gets the active connection
	GetConnect() *rbm_pool
	// GetConnectStatus get the status of connection
	GetConnectStatus() bool
	// CloseConnection closes the active connection
	CloseConnection() error

	// SimpleQueueDeclare used to declare a single Queue into RabbitMQ and returns it or an error
	SimpleQueueDeclare(sq Queue) (queue amqp.Queue, err error)
	// CompleteQueueDeclare used to declare a multiple Queue into RabbitMQ and returns a list of errors if happens.
	//
	// NOTE: If you run this function defining the Bind field contained in Queue, you must have to had defined
	// an Exchange first and then passing it to the field.
	CompleteQueueDeclare(sq []Queue) []error

	// SimpleExchangeDeclare used to declare a single Exchange into RabbitMQ and returns an error if happens
	SimpleExchangeDeclare(se Exchange) error
	// CompleteExchangeDeclare used to declare a multiple Exchange into RabbitMQ and returns a list of errors if happens
	CompleteExchangeDeclare(ce []Exchange) []error

	// CompleteDeclare used to fully declare multiple Queue and Exchange into
	// RabbitMQ and returns a list of errors if happens.
	//
	// NOTE: You can pass empty arrays to this function if not present in your project. If your project
	// doesn't contain binds, just don't set the Bind field contained in Queue struct.
	CompleteDeclare(cq []Queue, ce []Exchange) []error

	// Producer publishes a Message to RabbitMQ following the configuration passed on ProducerConfig
	Producer(ctx context.Context, pc *ProducerConfig, msg *Message) error
	// Consumer consumes a Queue on RabbitMQ following the configuration passed on ConsumerConfig
	Consumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery))
	// StartConsumer starts a consumer routine listening to a Queue of RabbitMQ
	// following the configuration passed on ConsumerConfig.
	//
	// There is a DEFAULT_MAX_RECONNECT_TIMES variable that defines on 3 the number of retries to reconnect to the
	// RabbitMQ service currently running. You can define this number by setting an env variable called
	// SRV_RMQ_MAXX_RECONNECT_TIMES
	StartConsumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery))
}

type rbm_pool struct {
	conn                 *amqp.Connection
	channel              *amqp.Channel
	conf                 *config.Config
	err                  chan error
	MAXX_RECONNECT_TIMES int
	connStatus           bool
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
		conf.RMQ_MAXX_RECONNECT_TIMES = DEFAULT_MAX_RECONNECT_TIMES
	}

	rbmpool := &rbm_pool{
		conf:       conf,
		err:        make(chan error),
		connStatus: false,
	}
	return rbmpool
}

func (rbm *rbm_pool) Connect() (RabbitInterface, error) {
	var err error

	if rbm.connStatus {
		log.Println("There is already a connection, returning Conn")
		return rbm, nil
	}

	rbm.conn, err = amqp.Dial(rbm.conf.RMQ_URI)
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ")
		return rbm, err
	}

	notifyConnClose := make(chan *amqp.Error)
	rbm.conn.NotifyClose(notifyConnClose) // Listen to Connection NotifyClose

	rbm.channel, err = rbm.conn.Channel()
	if err != nil {
		log.Println("Erro to Connect in RabbitMQ Channel")
		return rbm, err
	}

	notifyChanClose := make(chan *amqp.Error)
	rbm.channel.NotifyClose(notifyChanClose) // Listen to Channel NotifyClose

	go func() {
		select {
		case <-notifyConnClose:
			log.Println("connection closed")
			rbm.err <- errors.New("connection closed")
			rbm.connStatus = false
		case <-notifyChanClose:
			rbm.err <- errors.New("channel closed")
			rbm.connStatus = false
		}
	}()

	rbm.connStatus = true
	log.Println("New RabbitMQ Connect Success")

	return rbm, nil
}

func (rbm *rbm_pool) GetConnect() *rbm_pool {
	return rbm
}

func (rbm *rbm_pool) GetConnectStatus() bool {
	return rbm.connStatus
}

func (rbm *rbm_pool) CloseConnection() error {
	if err := rbm.conn.Close(); err != nil {
		log.Println("error closing rabbit connection", err)
		return err
	}

	log.Println("RabbitMQ connection closed successfully")
	return nil
}
