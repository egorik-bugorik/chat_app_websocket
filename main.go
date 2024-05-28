package main

import (
	"context"
	"net/http"
)

func main() {

	setupAPI()
	http.ListenAndServe(":9000", nil)

}

func setupAPI() {
	ctx := context.Background()
	manager := NewManager(ctx)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serverWS)
	http.HandleFunc("/login", manager.loginHandler)

}
