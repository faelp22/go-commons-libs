package mongodb

import (
	"context"
	"errors"
	"os"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBInterface interface {
	GetCollection() (*mongo.Collection, error)
	GetCollectionByName(name string) *mongo.Collection
}

type mongodb_pool struct {
	DB                  *mongo.Client
	DBName              string
	DBDefaultCollection string
}

var mdbpool = &mongodb_pool{}
var ctx = context.TODO()

func New(conf *config.Config) MongoDBInterface {

	SRV_MDB_URI := os.Getenv("SRV_MDB_URI")
	if SRV_MDB_URI != "" {
		conf.MDB_URI = SRV_MDB_URI
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_MDB_URI é obrigatória!")
	}

	SRV_MDB_NAME := os.Getenv("SRV_MDB_NAME")
	if SRV_MDB_NAME != "" {
		conf.MDB_NAME = SRV_MDB_NAME
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_MDB_NAME é obrigatória!")
	}

	SRV_MDB_DEFAULT_COLLECTION := os.Getenv("SRV_MDB_DEFAULT_COLLECTION")
	if SRV_MDB_DEFAULT_COLLECTION != "" {
		conf.MDB_DEFAULT_COLLECTION = SRV_MDB_DEFAULT_COLLECTION
	}

	if mdbpool != nil && mdbpool.DB != nil && mdbpool.DBName != "" {

		return mdbpool

	} else {

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MDB_URI))
		if err != nil {
			log.Fatal().Str("ERRO_CONECTION", "Erro to make Connect DB").Msg(err.Error())
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatal().Str("ERRO_CONECTION_PING", "Erro to Ping Connect DB").Msg(err.Error())
		}

		mdbpool = &mongodb_pool{
			DB:                  client,
			DBName:              conf.MDB_NAME,
			DBDefaultCollection: conf.MDB_DEFAULT_COLLECTION,
		}
	}

	return mdbpool
}

func (mdbp *mongodb_pool) GetCollection() (*mongo.Collection, error) {

	if mdbp.DBDefaultCollection == "" {
		return nil, errors.New("para usar esse método a variável SRV_MDB_DEFAULT_COLLECTION precisa ser informada ou use GetCollectionByName")
	}

	return mdbp.DB.Database(mdbp.DBName).Collection(mdbp.DBDefaultCollection), nil
}

func (mdbp *mongodb_pool) GetCollectionByName(name string) *mongo.Collection {
	return mdbp.DB.Database(mdbp.DBName).Collection(name)
}

func ObjectIDFromHex(hex string) (objectID primitive.ObjectID, err error) {
	objectID, err = primitive.ObjectIDFromHex(hex)
	if err != nil {
		log.Error().Str("FunctionName", "ObjectIDFromHex").Msg(err.Error())
		return objectID, err
	}
	return objectID, nil
}
