package rabbitmq

import (
	"errors"
	"time"

	"github.com/phuslu/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerConfig struct {
	Queue            string
	Consumer         string
	AutoAck          bool
	Exclusive        bool
	NoLocal          bool
	NoWait           bool
	Args             amqp.Table
	ControlQosConfig ControlQosConfig
}

type ControlQosConfig struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

func (rbm *Rbm_pool) Consumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {

	if cc.ControlQosConfig.PrefetchCount > 0 {
		err := rbm.channel.Qos(cc.ControlQosConfig.PrefetchCount, cc.ControlQosConfig.PrefetchSize, cc.ControlQosConfig.Global)
		if err != nil {
			log.Error().Str("FunctionName", "Consumer").Str("ERRO_CONSUMER", "Failed to set QoS").Msg(err.Error())
			return
		}
	}

	if cc.Consumer == "" {
		cc.Consumer = rbm.conf.AppName
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
		log.Error().Str("FunctionName", "Consumer").Str("ERRO_CONSUMER", "Failed to register a consumer").Msg(err.Error())
		return
	}

	go func() {
		log.Info().Str("FunctionName", "Consumer").Msg("Close Consumer")
		for msg := range msgs {
			callback(&msg)
		}
		log.Info().Str("FunctionName", "Consumer").Msg("Close Consumer")
	}()
}

func (rbm *Rbm_pool) StartConsumer(cc *ConsumerConfig, callback func(msg *amqp.Delivery)) {
	count := 0
	for {

		if rbm.connStatus {
			go rbm.Consumer(cc, callback)
		}

		if count >= rbm.conf.RMQ_MAXX_RECONNECT_TIMES {
			log.Fatal().Str("FunctionName", "StartConsumer").Msg("Erro to reconnect 3 times in RabbitMQ")
		}

		if err := <-rbm.err; err != nil {

			log.Warn().Str("FunctionName", "StartConsumer").Msg("Connection is closed, trying to reconnect in RabbitMQ")

			rb_conn, err := rbm.Connect()
			if err != nil {
				go func() { rbm.err <- errors.New("connection closed re trying") }()
				count++
				log.Warn().Str("FunctionName", "StartConsumer").Msg("Waiting 30 seconds to try again")
				time.Sleep(30 * time.Second) // wait 30 seconds
			} else {
				count = 0
				rbm.conn = rb_conn.GetConnect().conn
				rbm.connStatus = rb_conn.GetConnect().connStatus
			}
		}
	}
}
