package main

import (
	"context"
	"fmt"
	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
	"github.com/go-redis/redis/v8"
)

func main() {
	rdsConfig := config.Config{
		RedisDBConfig: &config.RedisDBConfig{
			RDB_HOST: "localhost",
			RDB_PORT: "6379",
			RDB_DB:   0,
			RDB_DSN:  "redis://localhost:6379/0",
		},
	}

	redisClient := redisdb.New(&rdsConfig)

	ctx := context.Background()

	redisClient.Subscriber(ctx, worker)
}

func worker(msg *redis.Message) {
	// example to handle a message
	fmt.Println(fmt.Sprintf("Message received from channel [%s]: %s", msg.Channel, msg.Payload))
}
