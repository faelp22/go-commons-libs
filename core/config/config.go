package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/phuslu/log"
)

var (
	VERSION    = "0.1.0-dev"
	COMMIT     = "ABCDEFG-dev"
	oldAppName string
)

const (
	DEVELOPER           = "developer"
	HOMOLOGATION        = "homologation"
	PRODUCTION          = "production"
	TARGET_DEPLOY_LOCAL = "local"
	TARGET_DEPLOY_NUVEM = "nuvem"
)

type Config struct {
	AppName           string `json:"app_name"`
	AppMode           string `json:"app_mode"`
	AppLogLevel       string `json:"app_log_level"`
	AppTargetDeploy   string `json:"app_target_deploy"`
	AppCommitShortSha string `json:"commit_short_sha"`
	AppVersion        string `json:"version"`
	globalLog         *log.Logger
	*HttpConfig
	*MongoDBConfig
	*RedisDBConfig
	*PGSQLConfig
	*RMQConfig
	*BlobStorage
}

type HttpConfig struct {
	PORT   string      `json:"port"`
	Logger *log.Logger `json:"-"`
}

type MongoDBConfig struct {
	MDB_URI                string `json:"mdb_uri"`
	MDB_NAME               string `json:"mdb_name"`
	MDB_DEFAULT_COLLECTION string `json:"mdb_default_collection"`
}

type RedisDBConfig struct {
	RDB_HOST string `json:"rdb_host"`
	RDB_PORT string `json:"rdb_port"`
	RDB_USER string `json:"rdb_user"`
	RDB_PASS string `json:"rdb_pass"`
	RDB_DB   int64  `json:"rdb_db"`
	RDB_DSN  string `json:"-"`
}

type PGSQLConfig struct {
	DB_DRIVE                  string `json:"db_drive"`
	DB_HOST                   string `json:"db_host"`
	DB_PORT                   string `json:"db_port"`
	DB_USER                   string `json:"db_user"`
	DB_PASS                   string `json:"db_pass"`
	DB_NAME                   string `json:"db_name"`
	DB_SSL_MODE               bool   `json:"db_ssl_mode"`
	DB_CONNECT_TIMEOUT        int    `json:"db_connect_timeout"`
	DB_DSN                    string `json:"-"`
	DB_SET_MAX_OPEN_CONNS     int    `json:"db_set_max_open_conns"`
	DB_SET_MAX_IDLE_CONNS     int    `json:"db_set_max_idle_conns"`
	DB_SET_CONN_MAX_LIFE_TIME int    `json:"db_set_conn_max_life_time"`
}

type RMQConfig struct {
	RMQ_URI                  string `json:"rmq_uri"`
	RMQ_MAXX_RECONNECT_TIMES int    `json:"rmq_maxx_reconnect_times"`
}

type BlobStorage struct {
	BS_ACCOUNT_NAME    string `json:"account_name"`
	BS_ACCOUNT_KEY     string `json:"account_key"`
	BS_SERVICE_URL     string `json:"service_url"`
	BS_URL_EXPIRY_TIME int64  `json:"bs_url_expiry_time"`
}

var default_conf *Config

func NewDefaultConf() *Config {

	if default_conf != nil {
		return default_conf
	}

	default_conf = &Config{
		AppVersion:        VERSION,
		AppCommitShortSha: COMMIT,
		globalLog: &log.Logger{
			Level:  log.InfoLevel,
			Caller: 1,
			// TimeField: "date",
			// TimeFormat: time.RFC3339,
			// Writer: &log.IOWriter{os.Stdout},
		},
	}

	SRV_APP_NAME := os.Getenv("SRV_APP_NAME")
	if SRV_APP_NAME != "" {
		default_conf.setAppName(SRV_APP_NAME)
	} else {
		default_conf.setAppName("")
	}

	SRV_APP_MODE := os.Getenv("SRV_APP_MODE")
	if SRV_APP_MODE != "" {
		default_conf.setAppMode(SRV_APP_MODE)
	} else {
		default_conf.setAppMode(PRODUCTION)
	}

	SRV_APP_LOG_LEVEL := os.Getenv("SRV_APP_LOG_LEVEL")
	if SRV_APP_LOG_LEVEL != "" {
		default_conf.setAppLogLevel(SRV_APP_LOG_LEVEL)
	} else {
		default_conf.setAppLogLevel("Info")
	}

	SRV_APP_TARGET_DEPLOY := os.Getenv("SRV_APP_TARGET_DEPLOY")
	if SRV_APP_TARGET_DEPLOY != "" {
		default_conf.setAppTargetDeploy(SRV_APP_TARGET_DEPLOY)
	} else {
		default_conf.setAppTargetDeploy(TARGET_DEPLOY_NUVEM)
	}

	return default_conf
}

func (c *Config) SetGlobalLogger(logger *log.Logger) {
	c.globalLog = logger
	log.DefaultLogger = *c.globalLog
}

func (c *Config) GetGlobalLogger() *log.Logger {
	return c.globalLog
}

func (c *Config) Reload() {
	c.setAppLogLevel(c.AppLogLevel)
	c.setAppMode(c.AppMode)
	c.setAppName(c.AppName)
	c.setAppTargetDeploy(c.AppTargetDeploy)
}

func (c *Config) setAppLogLevel(level string) {
	switch strings.ToUpper(level) {
	case "TRACE":
		c.globalLog.Level = log.TraceLevel
	case "DEBUG":
		c.globalLog.Level = log.DebugLevel
	case "INFO":
		c.globalLog.Level = log.InfoLevel
	case "WARN":
		c.globalLog.Level = log.WarnLevel
	case "ERROR":
		c.globalLog.Level = log.ErrorLevel
	case "FATAL":
		c.globalLog.Level = log.FatalLevel
	case "PANIC":
		c.globalLog.Level = log.PanicLevel
	default:
		c.globalLog.Level = log.InfoLevel
		log.DefaultLogger = *c.globalLog
		log.Info().Str("LogLevel", "Info").Msg(fmt.Sprintf("Attention, The value [%s] is not valid, see the available options: (Trace, Debug, Info, Warn, Error, Fatal and Panic). Setting the default logging level to [Info].", level))
	}

	c.AppLogLevel = c.globalLog.Level.String()
	log.DefaultLogger = *c.globalLog
}

func (c *Config) setAppMode(mode string) {
	switch strings.ToLower(mode) {
	case PRODUCTION:
		c.AppMode = PRODUCTION
	case HOMOLOGATION:
		c.AppMode = HOMOLOGATION
	case DEVELOPER:
		c.AppMode = DEVELOPER
	default:
		c.AppMode = PRODUCTION
		log.Info().Msg(fmt.Sprintf("Attention, The value [%s] is not valid, see the available options: (developer, homologation and production). Setting the default app mode to [production].", mode))
	}
}

func (c *Config) setAppName(name string) {

	if name != "" {
		c.AppName = name
	} else if name == "" && c.AppName == "" {
		c.AppName = fmt.Sprintf("App@%s", uuid.New().String()[:8])
	} else if c.AppName == oldAppName {
		old_tmp_name := strings.Split(oldAppName, "@")[0]
		c.AppName = fmt.Sprintf("%s@%s", old_tmp_name, uuid.New().String()[:8])
	} else {
		c.AppName = fmt.Sprintf("%s@%s", c.AppName, uuid.New().String()[:8])
	}

	oldAppName = c.AppName
}

func (c *Config) setAppTargetDeploy(target string) {
	switch strings.ToLower(target) {
	case TARGET_DEPLOY_LOCAL:
		c.AppTargetDeploy = TARGET_DEPLOY_LOCAL
	case TARGET_DEPLOY_NUVEM:
		c.AppTargetDeploy = TARGET_DEPLOY_NUVEM
	default:
		c.AppMode = TARGET_DEPLOY_NUVEM
		log.Info().Msg(fmt.Sprintf("Attention, The value [%s] is not valid, see the available options: (Local or Nuvem). Setting the default app mode to [production].", target))
	}
}
