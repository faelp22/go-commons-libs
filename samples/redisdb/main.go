package main

import (
	"context"
	"fmt"
	"log"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
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
		log.Println("Registro salvo")
	}

	data2, err := redisConn.ReadData(ctx, TESTE_KEY)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(data2))
}
