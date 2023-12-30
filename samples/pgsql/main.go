package main

import (
	"context"

	"github.com/phuslu/log"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/pgsql"
)

func main() {

	conf := config.NewDefaultConf()
	conf.AppTargetDeploy = config.TARGET_DEPLOY_LOCAL
	conf.PGSQLConfig = &config.PGSQLConfig{
		DB_HOST: "localhost",
		DB_USER: "postgres",
		DB_PASS: "123456",
		DB_NAME: "postgres",
	}

	dbPool := pgsql.New(conf)
	app := NewAppService(dbPool)

	log.Info().Str("Status", "Ok").Msg(app.TestGetNow(context.Background()))
}

type app_service struct {
	dbp pgsql.DatabaseInterface
}

func NewAppService(database_pool pgsql.DatabaseInterface) *app_service {
	return &app_service{
		dbp: database_pool,
	}
}

func (ps *app_service) TestGetNow(ctx context.Context) string {
	stmt, err := ps.dbp.GetDB().PrepareContext(ctx, "select now()")
	if err != nil {
		log.Error().Str("ErroPrepareContext", "Erro ao fazer ao criar o stmt").Msg(err.Error())
	}

	defer stmt.Close()

	var now string

	if err := stmt.QueryRowContext(ctx).Scan(&now); err != nil {
		log.Error().Str("ErroQueryRowContext", "Erro ao fazer a consulta no banco de dados").Msg(err.Error())
	}

	return now
}
