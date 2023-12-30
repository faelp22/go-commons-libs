package main

import (
	"context"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
	"github.com/phuslu/log"
)

func main() {
	conf := &config.Config{
		RedisDBConfig: &config.RedisDBConfig{},
	}

	redisConn := redisdb.New(conf)
	ctx := context.Background()

	const TESTE_KEY = "TESTE_DEV"

	data := []byte(`{"ok": "ok"}`)

	ok := redisConn.SaveData(ctx, TESTE_KEY, data, 0)
	if ok {
		log.Info().Msg("Registro salvo")
	}

	data2, err := redisConn.ReadData(ctx, TESTE_KEY)
	if err != nil {
		log.Error().Msg(err.Error())
	}

	log.Info().Msg(string(data2))
}
