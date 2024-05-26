package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	eggres     chan []byte
}

func (c *Client) readMessage() {

	defer func() {
		log.Println("removing client from manager ")
		c.manager.removeClient(c)
	}()
	for {
		log.Println("about reading msg...")
		messageType, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Println("error while reading ::: ", err)

			}
			break

		}

		for client, _ := range c.manager.clients {
			client.eggres <- payload

		}
		log.Println("MessageType from client --->>> ", messageType)
		log.Println("Message from client --->>> ", payload)

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
			if err := c.connection.WriteMessage(websocket.TextMessage, msg); err != nil {
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
		eggres:     make(chan []byte),
	}
}
