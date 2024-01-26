package main

import (
	"context"
	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
)

func main() {
	rdsConfig := config.Config{
		RedisDBConfig: &config.RedisDBConfig{},
	}

	redisClient := redisdb.New(&rdsConfig)

	ctx := context.Background()

	err := redisClient.Publish(ctx, []byte("test-message"))
	if err != nil {
		// handler your error
		return
	}
}
