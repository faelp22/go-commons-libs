package pgsql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"

	_ "github.com/lib/pq"
)

type DatabaseInterface interface {
	GetDB() (DB *sql.DB)
}

type dabase_pool struct {
	DB *sql.DB
}

var dbpool = &dabase_pool{}

func New(conf *config.Config) *dabase_pool {

	switch conf.DB_DRIVE {
	case "postgres":

		if conf.Mode != config.PRODUCTION {
			conf.DB_DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				conf.DB_HOST, conf.DB_PORT, conf.DB_USER, conf.DB_PASS, conf.DB_NAME)
		} else {
			conf.DB_DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
				conf.DB_HOST, conf.DB_PORT, conf.DB_USER, conf.DB_PASS, conf.DB_NAME)
		}

		dbpool = PGConn(conf)
	default:
		panic("Drive não implementado")
	}

	return dbpool
}

func (d *dabase_pool) GetDB() (DB *sql.DB) {
	return d.DB
}

func PGConn(conf *config.Config) *dabase_pool {

	if dbpool != nil && dbpool.DB != nil {

		return dbpool

	} else {

		db, err := sql.Open(conf.DB_DRIVE, conf.DB_DSN)
		if err != nil {
			log.Fatal(err)
		}
		// defer db.Close()

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(5 * time.Minute)

		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}

		dbpool = &dabase_pool{
			DB: db,
		}
	}

	return dbpool
}
