package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/httpserver"
	"github.com/gorilla/mux"
)

func main() {
	conf := config.NewDefaultConf()
	router := mux.NewRouter()

	// Exemplo 1: Ignorar logs de arquivos estáticos
	logConfig := &httpserver.LoggingMiddlewareConfig{
		Enabled: true,
		IgnorePaths: []string{
			"/assets/",     // Ignora todos os arquivos em /assets/
			"/static/",     // Ignora todos os arquivos em /static/
			"/favicon.ico", // Ignora favicon
			"/public/",     // Ignora arquivos públicos
		},
	}

	// Rotas de exemplo
	router.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Hello World"}`)
	}).Methods("GET")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy"}`)
	}).Methods("GET")

	// Simular rota de arquivos estáticos (normalmente seria http.FileServer)
	router.PathPrefix("/assets/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Static file content"))
	})

	// Criar servidor com configuração de log customizada
	srv := httpserver.NewWithLogConfig(router, conf, nil, logConfig)

	log.Printf("Server starting on port %s", conf.PORT)
	log.Printf("Log config: Enabled=%v, IgnorePaths=%v", logConfig.Enabled, logConfig.IgnorePaths)
	log.Printf("\nTeste os endpoints:")
	log.Printf("  - http://localhost:%s/api/hello (vai gerar log)", conf.PORT)
	log.Printf("  - http://localhost:%s/assets/app.js (NÃO vai gerar log)", conf.PORT)
	log.Printf("  - http://localhost:%s/health (vai gerar log)", conf.PORT)

	log.Fatal(srv.ListenAndServe())
}
