package main

import "net/http"

func main() {

	setupAPI()
	http.ListenAndServe(":9000", nil)

}

func setupAPI() {
	manager := NewManager()
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serverWS)

}
