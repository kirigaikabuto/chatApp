package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

func main() {
	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("public"))
	router.Methods("GET").Path("/").Handler(fs)
	router.Methods("GET").Path("/ws").HandlerFunc(handleConnection)
	go handleMessages()
	log.Println("http server started on port :8000")
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal("Listen and serve:", err)
	}
}
func handleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	clients[ws] = true
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error :%v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
