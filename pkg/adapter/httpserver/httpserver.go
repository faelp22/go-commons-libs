package httpserver

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	// "github.com/rs/cors"
)

func NewHTTPServer(r *mux.Router, port string) *http.Server {

	// handler := cors.Default().Handler(r)
	// handler := cors.AllowAll().Handler(r)
	handler := r

	srv := &http.Server{
		ReadTimeout:  10 * time.Second, // Aguarda 10 segundos
		WriteTimeout: 10 * time.Second, // Responde em 10 segundos
		Addr:         ":" + port,
		Handler:      handler,
		ErrorLog:     log.New(os.Stderr, "logger: ", log.Lshortfile),
	}

	return srv
}
