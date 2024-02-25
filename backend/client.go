package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type ClentList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	// egress is a channel which prevents concurrent writing i.e. it stores in a unbuffered channel which then sends the code
	// This is because gorilla can only send one message at a time so it will discard mssg if the client spams
	egress chan Event
}

func NewClient(conn *websocket.Conn, mang *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    mang,
		egress:     make(chan Event),
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		//cleanup connection: helps to remove the client from manager if connection closed by the client
		c.manager.removeClient(c)
	}()
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading the message: %v", err)
			}
			break
		}
		// this is for testing and before Events were added
		// for wsclient := range c.manager.clients {
		// 	wsclient.egress <- payload // adding the message to all the clients connected to this clients manager
		// }
		// println(messageType)
		// println(string(payload))

		// after changing egress from []byte to Events
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Println("error marshalling data: ", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println(err)
		}

	}
}
func (c *Client) WriteMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println("error marshalling data: ", err)
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Failed to send a message: ", err)
			}
			log.Println("Message Sent")
		}
	}
}
