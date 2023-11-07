package main

import (
	"context"
	"encoding/json"
	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
	"github.com/google/uuid"
	"log"
)

func main() {
	conf := &config.Config{
		RedisDBConfig: &config.RedisDBConfig{},
	}

	redisConn := redisdb.New(conf)
	ctx := context.Background()

	log.Println("Salvando registros")
	for i := 0; i < 10; i++ {
		id, _ := uuid.NewUUID()
		val, _ := uuid.NewUUID()
		ok := redisConn.SaveHSetData(ctx, "test-key", "key-"+id.String(), val.String())
		if ok {
			log.Println("Registro salvo")
		}
	}

	log.Println("Lendo registro")
	data, err := redisConn.ReadHSetData(ctx, "test-key")
	if err != nil {
		log.Println(err)
	}

	j, _ := json.MarshalIndent(data, "", "  ")
	log.Println(string(j))

	log.Println("Deletando registro")
	ok := redisConn.DeleteAllHSetData(ctx, "test-key")
	if ok {
		log.Println("todos os registros deletados")
	}
}
