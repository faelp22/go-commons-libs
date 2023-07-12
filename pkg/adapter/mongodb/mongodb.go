package mongodb

import (
	"context"
	"log"

	"github.com/faelp22/go-commons-libs/core/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBInterface interface {
	GetCollection() (DBCollection *mongo.Collection)
	GetCollectionByName(name string) (DBCollection *mongo.Collection)
}

type mongodb_pool struct {
	DB           *mongo.Client
	DBName       string
	DBCollection string
}

var mdbpool = &mongodb_pool{}
var ctx = context.TODO()

func New(conf *config.Config) MongoDBInterface {

	if mdbpool != nil && mdbpool.DB != nil && mdbpool.DBName != "" {

		return mdbpool

	} else {
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MDB_URI))
		if err != nil {
			log.Fatal("Erro to make Connect DB:", err.Error())
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatal("Erro to contact DB:", err.Error())
		}

		mdbpool = &mongodb_pool{
			DB:           client,
			DBName:       conf.MDB_NAME,
			DBCollection: conf.MDB_COLLECTION,
		}
	}

	return mdbpool
}

func (d *mongodb_pool) GetCollection() (DBCollection *mongo.Collection) {
	return d.DB.Database(d.DBName).Collection(d.DBCollection)
}

func (d *mongodb_pool) GetCollectionByName(name string) (DBCollection *mongo.Collection) {
	return d.DB.Database(d.DBName).Collection(name)
}

func ObjectIDFromHex(hex string) (objectID primitive.ObjectID, err error) {
	objectID, err = primitive.ObjectIDFromHex(hex)
	if err != nil {
		log.Println(err.Error())
		return objectID, err
	}
	return objectID, nil
}
