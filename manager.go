package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	clients ClientList
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

func (m *Manager) serverWS(w http.ResponseWriter, r *http.Request) {

	log.Println("Inner conenction ")

	conn, err := websocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error while upgrade upgrader :::", err)

	}

	client := NewClient(conn, m)

	m.addClient(client)

	//	start reading
	go client.readMessage()
	go client.writeMessage()

}

func (m *Manager) addClient(client *Client) {

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client] = true

}
func (m *Manager) removeClient(client *Client) {

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
	m.clients[client] = true

}
