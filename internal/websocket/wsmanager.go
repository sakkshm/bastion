package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WSManager struct {
	SessionID string

	Clients map[string]*Client

	Broadcast chan WSMessage // outgoing messages
	Incoming  chan WSMessage // incoming messages (from all clients)

	Register   chan *Client
	Unregister chan *Client

	Ctx    context.Context
	Cancel context.CancelFunc

	mu sync.RWMutex
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// allow all origins (dev only)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSManager(sessionID string) *WSManager {

	ctx, cancel := context.WithCancel(context.Background())

	return &WSManager{
		SessionID:  sessionID,
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan WSMessage),
		Incoming:   make(chan WSMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),

		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (ws *WSManager) Run() {
	for {
		select {
		case <-ws.Ctx.Done():

			// cleanup all clients
			for _, client := range ws.Clients {
				if !client.Closed {
					client.Closed = true
					close(client.Send)
				}
				client.Conn.Close()
			}

			return

		case client := <-ws.Register:
			ws.mu.Lock()
			ws.Clients[client.ClientID] = client
			ws.mu.Unlock()

			payload, _ := json.Marshal(map[string]string{
				"msg": "connected",
			})

			client.WriteToClient(&WSMessage{
				Type:      MsgInit,
				SessionID: ws.SessionID,
				ClientID:  client.ClientID,
				Payload:   payload,
			})

		case client := <-ws.Unregister:
			ws.mu.RLock()
			_, ok := ws.Clients[client.ClientID]
			ws.mu.RUnlock()

			if ok {
				ws.mu.Lock()
				delete(ws.Clients, client.ClientID)
				ws.mu.Unlock()

				if !client.Closed {
					client.Closed = true
					close(client.Send)
				}
			}

		case msg := <-ws.Broadcast:
			var dead []string

			ws.mu.RLock()
			clients := make([]*Client, 0, len(ws.Clients))
			for _, c := range ws.Clients {
				clients = append(clients, c)
			}
			ws.mu.RUnlock()

			for _, client := range clients {
				select {
				case client.Send <- msg:
				default:
					if !client.Closed {
						client.Closed = true
						close(client.Send)
					}
					dead = append(dead, client.ClientID)
				}
			}

			ws.mu.Lock()
			for _, id := range dead {
				delete(ws.Clients, id)
			}
			ws.mu.Unlock()

		case msg := <-ws.Incoming:
			go ws.handleIncoming(msg)
		}
	}
}

func (ws *WSManager) handleIncoming(msg WSMessage) {

	// check sessionID and clientID
	ws.mu.RLock()
	client, ok := ws.Clients[msg.ClientID]
	ws.mu.RUnlock()

	if !ok { return }
	if msg.SessionID != ws.SessionID {
		client.WriteErrToClient(fmt.Errorf("inavlid session id"))
		return
	}

	switch msg.Type {

	case MsgTerminalInput:
		var data struct {
			Input string `json:"input"`
		}
		_ = json.Unmarshal(msg.Payload, &data)

		fmt.Println("terminal:", data.Input)

	case MsgExec:
		var data struct {
			Cmd string `json:"cmd"`
		}

		_ = json.Unmarshal(msg.Payload, &data)

		fmt.Println("exec:", data.Cmd)

	default:
		ws.mu.RLock()
		client, ok := ws.Clients[msg.ClientID]
		ws.mu.RUnlock()
		if !ok {
			return
		}
		client.WriteErrToClient(fmt.Errorf("inavlid msg type"))
	}
}
