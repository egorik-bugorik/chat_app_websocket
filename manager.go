package main

import (
	"context"
	"encoding/json"
	errors "errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://localhost:9000":
		return true
	default:
		return false

	}
}

type Manager struct {
	clients  ClientList
	mu       sync.RWMutex
	handlers map[string]EventHandler
	otps     RetentionMap
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, time.Second*4),
	}
	m.setupEventHandlers()

	return m
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	if req.Username == "gorik" && req.Password == "123" {
		type Response struct {
			Otp string `json:"otp"`
		}

		res := Response{Otp: m.otps.NewOtp().Key}
		data, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Println("error while writing data ::: ", err)
			return
		}

	}
	w.WriteHeader(http.StatusUnauthorized)

}

func (m *Manager) routeHandler(event Event, c *Client) error {

	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("Therer is no such event!!!")

	}

}
func (m *Manager) serverWS(w http.ResponseWriter, r *http.Request) {

	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return

	}

	if !m.otps.VerifyOtp(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return

	}

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

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = sendMessage
}

func sendMessage(event Event, c *Client) error {

	var sen SendMessageEvent

	err := json.Unmarshal(event.Payload, &sen)
	if err != nil {

		return fmt.Errorf("coudln't unmarshal payload into sendMesgEv :::", err)

	}

	var broad NewMessageEvent

	broad.Sent = time.Now()
	broad.Message = sen.Message
	broad.From = sen.From

	data, err := json.Marshal(broad)
	if err != nil {
		return fmt.Errorf("coudln't marshal  newMsgEv:::", err)

	}

	outgoingEvent := Event{
		Type:    EventNewMessage,
		Payload: data,
	}

	for client := range c.manager.clients {

		client.eggres <- outgoingEvent
	}

	return nil

}
