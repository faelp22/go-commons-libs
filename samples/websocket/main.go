package main

import (
	"log"
	"net/http"
	"os"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/httpserver"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket CheckOrigin")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	defer ws.Close()

	log.Println("New terminal connection")

	// Echo server simples
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Printf("Received: %s", message)

		err = ws.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

func main() {
	os.Setenv("SRV_HTTP_PORT", "8080")

	conf := config.NewDefaultConf()
	router := mux.NewRouter()

	router.HandleFunc("/ws", handleWebSocket)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := httpserver.New(router, conf, nil)

	log.Printf("Server starting on port %s", conf.PORT)
	log.Fatal(srv.ListenAndServe())
}
