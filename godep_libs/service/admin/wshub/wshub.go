package wshub

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer with this period.
	pingPeriod = 60 * time.Second
)

type WSHub struct {
	Broadcast   chan []byte
	connections map[*connection]bool
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	conn *websocket.Conn
	send chan []byte
}

func NewWSHub() *WSHub {
	hub := &WSHub{
		connections: make(map[*connection]bool),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		Broadcast:   make(chan []byte),
	}

	go func() {
		for {
			select {
			case c := <-hub.register:
				hub.connections[c] = true
			case c := <-hub.unregister:
				if _, ok := hub.connections[c]; ok {
					delete(hub.connections, c)
					close(c.send)
				}
			case m := <-hub.Broadcast:
				for c := range hub.connections {
					select {
					case c.send <- m:
					default:
						close(c.send)
						delete(hub.connections, c)
					}
				}
			}
		}
	}()

	return hub
}

func (h *WSHub) ProcessWSConnection(conn *websocket.Conn, helloMessage []byte) {
	connection := &connection{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- connection

	if helloMessage != nil {
		connection.conn.WriteMessage(websocket.TextMessage, helloMessage)
	}

	// Write messages
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			connection.conn.Close()
			ticker.Stop()
		}()

		for {
			select {
			case message, ok := <-connection.send:
				connection.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					connection.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := connection.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			case <-ticker.C:
				connection.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := connection.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}()

	// Read messages
	defer func() {
		h.unregister <- connection
		connection.conn.Close()
	}()

	for {
		_, _, err := connection.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
