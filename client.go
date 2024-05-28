package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var (
	pongWait     = 5 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Message struct {
	message string
}
type Client struct {
	connection *websocket.Conn
	manager    *Manager
	eggres     chan Event
}

func (c *Client) readMessage() {

	defer func() {
		log.Println("removing client from manager ")
		c.manager.removeClient(c)
	}()
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("....pong is never coming....")
		return
	}
	c.connection.SetReadLimit(100)

	c.connection.SetPongHandler(c.handlePong)
	for {
		log.Println("about reading msg...")
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Println("error while reading ::: ", err)

			}
			break

		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			fmt.Printf("error unmarshall event ::%v", err)
			break
		}

		if err := c.manager.routeHandler(request, c); err != nil {
			log.Println("error handling event ::: ", err)

		}
		//c.eggres <- Event{Type: "new_message", Payload: json.RawMessage(`"helo from server "`)}
	}
}

func (c *Client) writeMessage() {

	timer := time.NewTicker(pingInterval)

	defer func() {

		log.Println("Goodbye client!")
		c.manager.removeClient(c)
	}()

	for {

		select {
		case msg, ok := <-c.eggres:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed")
				}
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Println("Error while marshaling data ::: ", err)
				return

			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("\n Fail to send message !!! %v", err)
			}
			log.Println("Mesage sent.")

		case <-timer.C:
			log.Println("ping")
			err := c.connection.WriteMessage(websocket.PingMessage, []byte(``))
			if err != nil {
				log.Println("ping write error :::", err)

				return
			}

		}

	}
}

func (c *Client) handlePong(data string) error {

	log.Println(":::pong:::")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func NewClient(ws *websocket.Conn, m *Manager) *Client {
	return &Client{
		connection: ws,
		manager:    m,
		eggres:     make(chan Event),
	}
}
