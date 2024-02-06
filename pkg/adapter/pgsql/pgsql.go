package pgsql

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	_ "github.com/lib/pq"
	"github.com/phuslu/log"
)

type DatabaseInterface interface {
	GetDB() *sql.DB
	CloseConnection() error
}

type dabase_pool struct {
	db *sql.DB
}

var dbpool = &dabase_pool{}

func New(conf *config.Config) *dabase_pool {

	conf.DB_DRIVE = "postgres"

	SRV_DB_HOST := os.Getenv("SRV_DB_HOST")
	if SRV_DB_HOST != "" {
		conf.DB_HOST = SRV_DB_HOST
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_DB_HOST é obrigatória!")
	}

	SRV_DB_PORT := os.Getenv("SRV_DB_PORT")
	if SRV_DB_PORT != "" {
		conf.DB_PORT = SRV_DB_PORT
	} else if conf.DB_PORT == "" {
		conf.DB_PORT = "5432"
	}

	SRV_DB_USER := os.Getenv("SRV_DB_USER")
	if SRV_DB_USER != "" {
		conf.DB_USER = SRV_DB_USER
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_DB_USER é obrigatória!")
	}

	SRV_DB_PASS := os.Getenv("SRV_DB_PASS")
	if SRV_DB_PASS != "" {
		conf.DB_PASS = SRV_DB_PASS
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_DB_PASS é obrigatória!")
	}

	SRV_DB_NAME := os.Getenv("SRV_DB_NAME")
	if SRV_DB_NAME != "" {
		conf.DB_NAME = SRV_DB_NAME
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável SRV_DB_NAME é obrigatória!")
	}

	SRV_DB_CONNECT_TIMEOUT := os.Getenv("SRV_DB_CONNECT_TIMEOUT")
	if SRV_DB_CONNECT_TIMEOUT != "" {

		var err error
		conf.DB_CONNECT_TIMEOUT, err = strconv.Atoi(SRV_DB_CONNECT_TIMEOUT)
		if err != nil {
			log.Error().Str("SRV_DB_CONNECT_TIMEOUT", "Invalid value").Str("SetDefaultValue", "10s").Msg(err.Error())
			conf.DB_CONNECT_TIMEOUT = 10 // 10s
		}

	} else if conf.DB_CONNECT_TIMEOUT == 0 {
		conf.DB_CONNECT_TIMEOUT = 10 // 10s
	}

	SRV_DB_SET_MAX_OPEN_CONNS := os.Getenv("SRV_DB_SET_MAX_OPEN_CONNS")
	if SRV_DB_SET_MAX_OPEN_CONNS != "" {

		var err error
		conf.DB_SET_MAX_OPEN_CONNS, err = strconv.Atoi(SRV_DB_SET_MAX_OPEN_CONNS)
		if err != nil {
			log.Error().Str("SRV_DB_SET_MAX_OPEN_CONNS", "Invalid value").Str("SetDefaultValue", "Max 10 Open Conns").Msg(err.Error())
			conf.DB_SET_MAX_OPEN_CONNS = 10 // Max 10 Open Conns
		}

	} else if conf.DB_SET_MAX_OPEN_CONNS == 0 {
		conf.DB_SET_MAX_OPEN_CONNS = 10 // Max 10 Open Conns
	}

	SRV_DB_SET_MAX_IDLE_CONNS := os.Getenv("SRV_DB_SET_MAX_IDLE_CONNS")
	if SRV_DB_SET_MAX_IDLE_CONNS != "" {

		var err error
		conf.DB_SET_MAX_IDLE_CONNS, err = strconv.Atoi(SRV_DB_SET_MAX_IDLE_CONNS)
		if err != nil {
			log.Error().Str("SRV_DB_SET_MAX_IDLE_CONNS", "Invalid value").Str("SetDefaultValue", "Max 10 Idle Conns").Msg(err.Error())
			conf.DB_SET_MAX_IDLE_CONNS = 10 // Max 10 Idle Conns
		}

	} else if conf.DB_SET_MAX_IDLE_CONNS == 0 {
		conf.DB_SET_MAX_IDLE_CONNS = 10 // Max 10 Idle Conns
	}

	SRV_DB_SET_CONN_MAX_LIFE_TIME := os.Getenv("SRV_DB_SET_CONN_MAX_LIFE_TIME")
	if SRV_DB_SET_CONN_MAX_LIFE_TIME != "" {

		var err error
		conf.DB_SET_CONN_MAX_LIFE_TIME, err = strconv.Atoi(SRV_DB_SET_CONN_MAX_LIFE_TIME)
		if err != nil {
			log.Error().Str("SRV_DB_SET_CONN_MAX_LIFE_TIME", "Invalid value").Str("SetDefaultValue", "5 minutes").Msg(err.Error())
			conf.DB_SET_CONN_MAX_LIFE_TIME = 5 // Max Open Conn Interval is 5 minutes
		}

	} else if conf.DB_SET_CONN_MAX_LIFE_TIME == 0 {
		conf.DB_SET_CONN_MAX_LIFE_TIME = 5 // Max Open Conn Interval is 5 minutes
	}

	SRV_DB_SSL_MODE := os.Getenv("SRV_DB_SSL_MODE")
	if SRV_DB_SSL_MODE != "" {

		var err error
		conf.DB_SSL_MODE, err = strconv.ParseBool(SRV_DB_SSL_MODE)
		if err != nil {
			log.Error().Str("SRV_DB_SSL_MODE", "Invalid value").Str("SetDefaultValue", "false").Msg(err.Error())
			conf.DB_SSL_MODE = false
		}

	}

	sslMode := "disable"
	if conf.DB_SSL_MODE {
		sslMode = "require"
	}

	conf.DB_DSN = fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' connect_timeout='%d' fallback_application_name='%s' sslmode='%s'",
		conf.DB_HOST, conf.DB_PORT, conf.DB_USER, conf.DB_PASS, conf.DB_NAME, conf.DB_CONNECT_TIMEOUT, conf.AppName, sslMode)

	dbpool = pgConn(conf)

	return dbpool
}

func (d *dabase_pool) GetDB() *sql.DB {
	return d.db
}

func pgConn(conf *config.Config) *dabase_pool {
	if dbpool != nil && dbpool.db != nil {
		return dbpool
	}

	db, err := sql.Open(conf.DB_DRIVE, conf.DB_DSN)
	if err != nil {
		log.Fatal().Str("FunctionName", "pgConn").Str("ERROR_CONNECTION", "Falha ao criar objecto de conexão do banco de dados").Msg(err.Error())
	}

	db.SetMaxOpenConns(conf.DB_SET_MAX_OPEN_CONNS)
	db.SetMaxIdleConns(conf.DB_SET_MAX_IDLE_CONNS)
	db.SetConnMaxLifetime(time.Duration(conf.DB_SET_CONN_MAX_LIFE_TIME) * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal().Str("FunctionName", "pgConn").Str("ERROR_CONNECTION_PING", "Falha ao tentar fazer o ping da conexão").Msg(err.Error())
	}

	dbpool = &dabase_pool{
		db: db,
	}

	log.Info().Str("FunctionName", "pgConn").Str("CONNECTION_STATUS", "Open").Msg("PGSQL connection Open successfully")

	return dbpool
}

func (d *dabase_pool) CloseConnection() error {
	if err := d.db.Close(); err != nil {
		log.Fatal().Str("FunctionName", "pgConn").Str("ERROR_CONNECTION_CLOSE", "Falha ao tentar fechar a conexão com o banco de dados").Msg(err.Error())
		return err
	}

	log.Info().Str("FunctionName", "pgConn").Str("CONNECTION_STATUS", "closed").Msg("PGSQL connection closed successfully")
	return nil
}
