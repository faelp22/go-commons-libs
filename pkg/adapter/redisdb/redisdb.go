package redisdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/go-redis/redis/v8"
)

type RedisClientInterface interface {
	ReadData(ctx context.Context, key string) (data []byte, err error)
	SaveData(ctx context.Context, key string, data []byte, timer time.Duration) (ok bool)
	SaveHSetData(ctx context.Context, key, field string, value interface{}) (ok bool)
	ReadHSetData(ctx context.Context, key string) (data map[string]string, err error)
	DeleteAllHSetData(ctx context.Context, key string) (ok bool)
}

type redis_client struct {
	rdb        *redis.Client
	modifyLock sync.RWMutex
}

func New(conf *config.Config) RedisClientInterface {

	SRV_RDB_HOST := os.Getenv("SRV_RDB_HOST")
	if SRV_RDB_HOST != "" {
		conf.RDB_HOST = SRV_RDB_HOST
	} else {
		log.Println("A variável SRV_RDB_HOST é obrigatória!")
		os.Exit(1)
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
		log.Println("Se o Redis precisa de [usuário] a variável SRV_RDB_USER é obrigatória!")
	}

	SRV_RDB_PASS := os.Getenv("SRV_RDB_PASS")
	if SRV_RDB_PASS != "" {
		conf.RDB_PASS = SRV_RDB_PASS
	} else {
		log.Println("Se o Redis precisa de [senha] a variável SRV_RDB_PASS é obrigatória!")
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
		log.Fatal(err)
	}

	rc := &redis_client{
		rdb: redis.NewClient(opt),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()

	status := rc.rdb.Ping(ctx)
	if status.String() != "ping: PONG" {
		log.Println("Erro ao conectar no Redis")
		log.Fatal(status)
	}

	return rc
}

func (rs *redis_client) ReadData(ctx context.Context, key string) (data []byte, err error) {

	rs.modifyLock.Lock()
	defer rs.modifyLock.Unlock()

	data, err = rs.rdb.Get(ctx, key).Bytes()
	if err != nil {
		log.Println(err.Error())
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
		log.Println(result.Err().Error())
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
		log.Println(result.Err().Error())
		return
	}
	return true
}
