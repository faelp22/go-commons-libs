package redisdb

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/go-redis/redis/v8"
	"github.com/phuslu/log"
)

type RedisClientInterface interface {
	GetClient() *redis.Client
	ReadData(ctx context.Context, key string) (data []byte, err error)
	SaveData(ctx context.Context, key string, data []byte, timer time.Duration) (ok bool)
	SaveHSetData(ctx context.Context, key, field string, value interface{}) (ok bool)
	ReadHSetData(ctx context.Context, key string) (data map[string]string, err error)
	DeleteAllHSetData(ctx context.Context, key string) (ok bool)
	Publish(ctx context.Context, message []byte) error
	Subscriber(ctx context.Context, callback func(msg *redis.Message))
}

type redis_client struct {
	rdb               *redis.Client
	modifyLock        sync.RWMutex
	pubSubChannelName string
}

func New(conf *config.Config) RedisClientInterface {

	SRV_RDB_HOST := os.Getenv("SRV_RDB_HOST")
	if SRV_RDB_HOST != "" {
		conf.RDB_HOST = SRV_RDB_HOST
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_RDB_HOST é obrigatória!")
	}

	SRV_RDB_PORT := os.Getenv("SRV_RDB_PORT")
	if SRV_RDB_PORT != "" {
		conf.RDB_PORT = SRV_RDB_PORT
	} else {
		conf.RDB_PORT = "6379"
	}

	SRV_RDB_USER := os.Getenv("SRV_RDB_USER")
	if SRV_RDB_USER != "" {
		conf.RDB_USER = SRV_RDB_USER
	} else {
		log.Info().Msg("Se o Redis precisa de [usuário] a variável SRV_RDB_USER é obrigatória!")
	}

	SRV_RDB_PASS := os.Getenv("SRV_RDB_PASS")
	if SRV_RDB_PASS != "" {
		conf.RDB_PASS = SRV_RDB_PASS
	} else {
		log.Info().Msg("Se o Redis precisa de [senha] a variável SRV_RDB_PASS é obrigatória!")
	}

	SRV_RDB_DB := os.Getenv("SRV_RDB_DB")
	if SRV_RDB_DB != "" {
		conf.RDB_DB, _ = strconv.ParseInt(SRV_RDB_DB, 10, 64)
	} else {
		conf.RDB_DB = 0
	}

	if len(conf.RDB_HOST) > 3 {

		// "redis://<user>:<pass>@localhost:6379/<db>"
		// https://redis.uptrace.dev/guide/go-redis.html#connecting-to-redis-server

		conf.RDB_DSN = fmt.Sprintf("redis://%s:%s@%s:%s/%v",
			conf.RDB_USER, conf.RDB_PASS, conf.RDB_HOST, conf.RDB_PORT, conf.RDB_DB)
	}

	opt, err := redis.ParseURL(conf.RDB_DSN)
	if err != nil {
		log.Fatal().Str("ERRO_REDIS_CON", "Erro ao tentar fazer o Parse da DSN").Msg(err.Error())
	}

	rc := &redis_client{
		rdb: redis.NewClient(opt),
	}

	SRV_RDB_PUBSUB_CHANNEL, ok := os.LookupEnv("SRV_RDB_PUBSUB_CHANNEL")
	if !ok {
		log.Info().Msg("Se o Redis usa pubsub a variável SRV_RDB_PUBSUB_CHANNEL é necessária!")
	} else {
		rc.pubSubChannelName = SRV_RDB_PUBSUB_CHANNEL
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()

	status := rc.rdb.Ping(ctx)
	if status.String() != "ping: PONG" {
		log.Fatal().Str("ERRO_REDIS_CON_PIN", "Erro ao conectar no Redis").Str("Status", status.String()).Msg(status.Err().Error())
	}

	return rc
}

func (rs *redis_client) GetClient() *redis.Client {
	return rs.rdb
}

func (rs *redis_client) ReadData(ctx context.Context, key string) (data []byte, err error) {

	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()

	data, err = rs.rdb.Get(ctx, key).Bytes()
	if err != nil {
		log.Error().Str("FunctionName", "ReadData").Str("ERRO_REDIS", "Erro ao tentar ler uma informação").Msg(err.Error())
		return
	}

	return
}

func (rs *redis_client) SaveData(ctx context.Context, key string, data []byte, timer time.Duration) (ok bool) {

	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()

	if timer <= 0 {
		timer = time.Duration(15 * time.Minute)
	}

	result := rs.rdb.Set(ctx, key, data, timer)
	if result.Val() == "1" || result.Val() == "OK" {
		ok = true
	}

	return
}

// SaveHSetData salva um hashset
func (rs *redis_client) SaveHSetData(ctx context.Context, key, datakey string, value interface{}) (ok bool) {
	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()
	result := rs.rdb.HSet(ctx, key, datakey, value)
	if result.Err() != nil {
		log.Error().Str("FunctionName", "SaveHSetData").Str("ERRO_REDIS", "Erro ao tentar salvar uma informação").Msg(result.Err().Error())
		return
	}
	return true
}

// ReadHSetData lê todos os dados de um hashset
func (rs *redis_client) ReadHSetData(ctx context.Context, key string) (data map[string]string, err error) {
	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()

	data, err = rs.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		log.Error().Str("FunctionName", "ReadHSetData").Str("ERRO_REDIS", "Erro ao tentar Ler uma informação").Msg(err.Error())
		return nil, err
	}

	return
}

// DeleteAllHSetData deleta todos os dados de um hashset
func (rs *redis_client) DeleteAllHSetData(ctx context.Context, key string) (ok bool) {
	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()

	result := rs.rdb.Del(ctx, key)
	if result.Err() != nil {
		log.Error().Str("FunctionName", "DeleteAllHSetData").Str("ERRO_REDIS", "Erro ao tentar Deletar uma informação").Msg(result.Err().Error())
		return
	}
	return true
}

// Publish envia uma mensagem para um canal específico no Redis.
//
// Esta função recebe um contexto (ctx), um nome de canal (channel) e uma
// mensagem (message) como parâmetros. A mensagem é do tipo []byte, permitindo
// a publicação de dados em vários formatos, incluindo strings convertidas para
// slice de bytes.
//
// Args:
//
//	ctx (context.Context): O contexto para controlar a execução e o cancelamento
//	    desta função. Pode ser usado para controlar timeouts e cancelamentos.
//	message ([]byte): A mensagem a ser publicada no canal especificado. Deve ser
//	    um slice de bytes, proporcionando flexibilidade para o formato da mensagem.
//
// Returns:
//
//	error: Retorna um erro caso a publicação da mensagem falhe. Isso pode acontecer
//	    devido a problemas de conexão com o servidor Redis ou outros erros de rede.
//	    Em caso de sucesso, retorna nil.
//
// A função tenta publicar a mensagem no canal especificado através do cliente Redis
// (rs.rdb). Em caso de falha na publicação, um erro é registrado e retornado.
// Se a publicação for bem-sucedida, a função retorna nil, indicando que a operação
// foi realizada sem erros.
//
// Exemplo de Uso:
//
//	err := redisClient.Publish(ctx, []byte("minha mensagem"))
//	if err != nil {
//	    // Tratar erro
//	}
func (rs *redis_client) Publish(ctx context.Context, message []byte) error {
	err := rs.rdb.Publish(ctx, rs.pubSubChannelName, message).Err()
	if err != nil {
		log.Error().Str("Error to publish message: ", err.Error())
		return err
	}

	return nil
}

// Subscriber cria uma inscrição em um canal específico no Redis e processa
// mensagens recebidas através de um callback fornecido.
//
// Esta função configura uma inscrição em um canal do Redis. Quando mensagens
// são publicadas no canal especificado, a função de callback fornecida é chamada
// para cada mensagem recebida. Esta função permite o processamento assíncrono de
// mensagens utilizando goroutines.
//
// Args:
//
//	ctx (context.Context): O contexto para controlar a execução e o cancelamento
//	    da inscrição e processamento de mensagens. Pode ser usado para gerenciar
//	    timeouts e cancelamentos.
//	callback (func(msg *redis.Message)): Uma função de callback que será chamada
//	    para cada mensagem recebida no canal. A função recebe uma mensagem do tipo
//	    *redis.Message como argumento.
//
// A função cria uma inscrição no canal especificado utilizando o cliente Redis (rs.rdb).
// Após a inscrição, a função entra em um laço, ouvindo mensagens. Cada mensagem recebida
// é processada em uma nova goroutine, utilizando a função de callback fornecida.
//
// Após o término do laço (quando o canal de mensagens é fechado), a função registra
// um log indicando que a inscrição no canal especificado foi encerrada.
//
// É importante garantir que o contexto passado para esta função seja adequadamente
// gerenciado, pois ele controla o ciclo de vida da inscrição e o processamento das mensagens.
//
// Exemplo de Uso:
//
//	callback := func(msg *redis.Message) {
//	    fmt.Println(msg.Channel, msg.Payload)
//	}
//	redisClient.Subscriber(ctx, callback)
func (rs *redis_client) Subscriber(ctx context.Context, callback func(msg *redis.Message)) {
	pubsub := rs.rdb.Subscribe(ctx, rs.pubSubChannelName)
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Info().Msg(fmt.Sprintf("subscribed to channel: %q", rs.pubSubChannelName))
	for msg := range ch {
		go callback(msg)
	}

	log.Info().Msg(fmt.Sprintf("subscribed to channel: %q is closed", rs.pubSubChannelName))
}
