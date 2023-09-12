package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"

	_ "github.com/lib/pq"
)

type DatabaseInterface interface {
	GetDB() *sql.DB
	GetConn() *sql.Conn
	CloseConnection() error
}

type dabase_pool struct {
	db   *sql.DB
	conn *sql.Conn
}

var dbpool = &dabase_pool{}

func New(conf *config.Config) *dabase_pool {

	SRV_DB_DRIVE := os.Getenv("SRV_DB_DRIVE")
	if SRV_DB_DRIVE != "" {
		conf.DB_DRIVE = SRV_DB_DRIVE
	} else {
		conf.DB_DRIVE = "postgres"
	}

	SRV_DB_HOST := os.Getenv("SRV_DB_HOST")
	if SRV_DB_HOST != "" {
		conf.DB_HOST = SRV_DB_HOST
	} else {
		log.Println("A variável SRV_DB_HOST é obrigatória!")
		os.Exit(1)
	}

	SRV_DB_PORT := os.Getenv("SRV_DB_PORT")
	if SRV_DB_PORT != "" {
		conf.DB_PORT = SRV_DB_PORT
	} else {
		conf.DB_PORT = "5432"
	}

	SRV_DB_USER := os.Getenv("SRV_DB_USER")
	if SRV_DB_USER != "" {
		conf.DB_USER = SRV_DB_USER
	} else {
		log.Println("A variável SRV_DB_USER é obrigatória!")
		os.Exit(1)
	}

	SRV_DB_PASS := os.Getenv("SRV_DB_PASS")
	if SRV_DB_PASS != "" {
		conf.DB_PASS = SRV_DB_PASS
	} else {
		log.Println("A variável SRV_DB_PASS é obrigatória!")
		os.Exit(1)
	}

	SRV_DB_NAME := os.Getenv("SRV_DB_NAME")
	if SRV_DB_NAME != "" {
		conf.DB_NAME = SRV_DB_NAME
	} else {
		log.Println("A variável SRV_DB_NAME é obrigatória!")
		os.Exit(1)
	}

	SRV_DB_SET_MAX_OPEN_CONNS := os.Getenv("SRV_DB_SET_MAX_OPEN_CONNS")
	if SRV_DB_SET_MAX_OPEN_CONNS != "" {
		conf.DB_SET_MAX_OPEN_CONNS, _ = strconv.Atoi(SRV_DB_SET_MAX_OPEN_CONNS)
	} else {
		conf.DB_SET_MAX_OPEN_CONNS = 10 // Max 10 Open Conns
	}

	SRV_DB_SET_MAX_IDLE_CONNS := os.Getenv("SRV_DB_SET_MAX_IDLE_CONNS")
	if SRV_DB_SET_MAX_IDLE_CONNS != "" {
		conf.DB_SET_MAX_IDLE_CONNS, _ = strconv.Atoi(SRV_DB_SET_MAX_IDLE_CONNS)
	} else {
		conf.DB_SET_MAX_IDLE_CONNS = 10 // Max 10 Idle Conns
	}

	SRV_DB_SET_CONN_MAX_LIFE_TIME := os.Getenv("SRV_DB_SET_CONN_MAX_LIFE_TIME")
	if SRV_DB_SET_CONN_MAX_LIFE_TIME != "" {
		conf.DB_SET_CONN_MAX_LIFE_TIME, _ = strconv.Atoi(SRV_DB_SET_CONN_MAX_LIFE_TIME)
	} else {
		conf.DB_SET_CONN_MAX_LIFE_TIME = 5 // Max Open Conn Interval is 5 minutes
	}

	SRV_DB_SSL_MODE := os.Getenv("SRV_DB_SSL_MODE")
	if SRV_DB_SSL_MODE != "" {
		conf.SRV_DB_SSL_MODE, _ = strconv.ParseBool(SRV_DB_SSL_MODE)
	} else {
		conf.SRV_DB_SSL_MODE = false // SSL Mode false by default
	}

	switch conf.DB_DRIVE {
	case "postgres":
		sslMode := "require"
		if conf.Mode != config.PRODUCTION {
			if !conf.SRV_DB_SSL_MODE {
				sslMode = "disable"
			}
		}

		conf.DB_DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			conf.DB_HOST, conf.DB_PORT, conf.DB_USER, conf.DB_PASS, conf.DB_NAME, sslMode)

		dbpool = pgConn(conf)
	default:
		panic("Drive não implementado")
	}

	return dbpool
}

func (d *dabase_pool) GetDB() *sql.DB {
	return d.db
}

func (d *dabase_pool) GetConn() *sql.Conn {
	return d.conn
}

func pgConn(conf *config.Config) *dabase_pool {
	if dbpool != nil && dbpool.db != nil && dbpool.conn != nil {
		return dbpool
	}

	db, err := sql.Open(conf.DB_DRIVE, conf.DB_DSN)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(conf.DB_SET_MAX_OPEN_CONNS)
	db.SetMaxIdleConns(conf.DB_SET_MAX_IDLE_CONNS)
	db.SetConnMaxLifetime(time.Duration(conf.DB_SET_CONN_MAX_LIFE_TIME) * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	dbpool = &dabase_pool{
		db:   db,
		conn: conn,
	}

	return dbpool
}

func (d *dabase_pool) CloseConnection() error {
	if err := d.conn.Close(); err != nil {
		log.Println("error closing pgsql connection", err)
		return err
	}

	log.Println("PGSQL connection closed successfully")
	return nil
}
