package config

const (
	DEVELOPER    = "developer"
	HOMOLOGATION = "homologation"
	PRODUCTION   = "production"
)

type Config struct {
	Mode string `json:"mode"`
	HttpConfig
	MongoDBConfig
	RedisDBConfig
	PGSQLConfig
	RMQConfig
}

type HttpConfig struct {
	PORT string `json:"port"`
}

type MongoDBConfig struct {
	MDB_URI        string `json:"mdb_uri"`
	MDB_NAME       string `json:"mdb_name"`
	MDB_COLLECTION string `json:"mdb_collection"`
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
	DB_DRIVE string `json:"db_drive"`
	DB_HOST  string `json:"db_host"`
	DB_PORT  string `json:"db_port"`
	DB_USER  string `json:"db_user"`
	DB_PASS  string `json:"db_pass"`
	DB_NAME  string `json:"db_name"`
	DB_DSN   string `json:"-"`
}

type RMQConfig struct {
	RMQ_URI string `json:"rmq_uri"`
}
