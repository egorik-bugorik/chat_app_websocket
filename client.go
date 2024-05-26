package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type ClientList map[*Client]bool

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

	}
}

func (c *Client) writeMessage() {

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

		}
	}
}

func NewClient(ws *websocket.Conn, m *Manager) *Client {
	return &Client{
		connection: ws,
		manager:    m,
		eggres:     make(chan Event),
	}
}
