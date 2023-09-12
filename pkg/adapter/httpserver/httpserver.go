package httpserver

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func New(r *mux.Router, conf *config.Config, opts ...cors.Options) *http.Server {
	var handler http.Handler
	handler = r

	if opts != nil {
		handler = cors.New(opts[0]).Handler(r)
	}

	SRV_PORT := os.Getenv("SRV_PORT")
	if SRV_PORT != "" {
		conf.PORT = SRV_PORT
	} else {
		conf.PORT = "3000"
	}

	srv := &http.Server{
		ReadTimeout:  10 * time.Second, // Aguarda 10 segundos
		WriteTimeout: 10 * time.Second, // Responde em 10 segundos
		Addr:         ":" + conf.PORT,
		Handler:      handler,
		ErrorLog:     log.New(os.Stderr, "logger: ", log.Lshortfile),
	}

	return srv
}
