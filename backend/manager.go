package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024, //setting up the maximum read and write limit
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { // allow message from anywhere
			return true
		},
	}
)

type Manager struct {
	clients      ClentList
	sync.RWMutex // synchronization so that it can handle if multiple people in the connection send a message simultaneously

	handlers map[string]EventHandler // mapping the events that you want to handle using this manager
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClentList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// This function is called when we need to upgrade an http connection to a ws connection
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New Connection")

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, m)
	m.addClient(client)

	go client.ReadMessages() // allow the new client to send and receive messages
	go client.WriteMessage()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = true // add the client to the client list
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok { // if client is alredy present
		client.connection.Close() // Close connection from manager side
		delete(m.clients, client) // remove connection by deleting from the manager
	}
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println(string(event.Payload))
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	// check if the event type is present or not if present then handle else return err
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("There is no such event type.")
	}
}
