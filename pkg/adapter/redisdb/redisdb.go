package redisdb

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/go-redis/redis/v8"
)

type RedisClientInterface interface {
	ReadData(ctx context.Context, key string) (data []byte, err error)
	SaveData(ctx context.Context, key string, data []byte, timer time.Duration) (ok bool)
}

type redis_client struct {
	rdb        *redis.Client
	modifyLock sync.RWMutex
}

func NewRedisClient(conf *config.Config) RedisClientInterface {
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
	if result.Val() == "1" {
		ok = true
	}

	return
}
