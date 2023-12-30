package main

import (
	"context"
	"encoding/json"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/redisdb"
	"github.com/google/uuid"
	"github.com/phuslu/log"
)

func main() {
	conf := &config.Config{
		RedisDBConfig: &config.RedisDBConfig{},
	}

	redisConn := redisdb.New(conf)
	ctx := context.Background()

	log.Debug().Msg("Salvando registros")
	for i := 0; i < 10; i++ {
		id, _ := uuid.NewUUID()
		val, _ := uuid.NewUUID()
		ok := redisConn.SaveHSetData(ctx, "test-key", "key-"+id.String(), val.String())
		if ok {
			log.Info().Msg("Registro salvo")
		}
	}

	log.Debug().Msg("Lendo registro")
	data, err := redisConn.ReadHSetData(ctx, "test-key")
	if err != nil {
		log.Error().Msg(err.Error())
	}

	j, _ := json.MarshalIndent(data, "", "  ")
	log.Debug().Msg(string(j))

	log.Debug().Msg("Deletando registro")
	ok := redisConn.DeleteAllHSetData(ctx, "test-key")
	if ok {
		log.Info().Msg("todos os registros deletados")
	}
}
