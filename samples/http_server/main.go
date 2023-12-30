package main

import (
	"net/http"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/httpserver"
	"github.com/gorilla/mux"
	"github.com/phuslu/log"
	"github.com/rs/cors"
)

func main() {

	logger := &log.Logger{
		Level:  log.InfoLevel,
		Caller: 0,
	}

	conf := config.NewDefaultConf()
	conf.AppTargetDeploy = config.TARGET_DEPLOY_LOCAL
	// conf.SetGlobalLogger(logger)

	conf.HttpConfig = &config.HttpConfig{
		Logger: logger,
	}

	r := mux.NewRouter()

	registerHealthCheckHandlers(r)

	corsOpts := &cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}

	srv := httpserver.New(r, conf, corsOpts)

	done := make(chan bool)
	go srv.ListenAndServe()
	log.Info().Str("Port", conf.PORT).Str("Mode", conf.AppMode).Str("Version", conf.AppVersion).Str("Commit", conf.AppCommitShortSha).Msg("Server Run")
	<-done
}

func registerHealthCheckHandlers(r *mux.Router) {
	r.Use(httpserver.ContentTypeJSONMiddleware)
	healthApi := r.PathPrefix("/api/v1").Subrouter()
	healthApi.Handle("/healthcheck", healthCheck()).Methods("GET", "OPTIONS")
}

func healthCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("FuntionName", "healthCheck").Msg("Ok Ok")
		SuccessHttpMsgHealthCheckOk.Write(w)
	})
}

var SuccessHttpMsgHealthCheckOk httpserver.HttpMsg = httpserver.HttpMsg{
	Msg:  "Server Ok",
	Code: http.StatusOK,
}
