package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	ClientID  string
	SessionID string
	Send      chan WSMessage // send buffer
	Closed    bool
}

func NewClient(conn *websocket.Conn, sessionID string) *Client {
	return &Client{
		Conn:      conn,
		ClientID:  generateClinetID(),
		SessionID: sessionID,
		Send:      make(chan WSMessage, 256),
		Closed:    false,
	}
}

func (c *Client) WriteToClient(msg *WSMessage) {
	c.Send <- *msg
}

func (c *Client) WriteErrToClient(err error) {
	payload, _ := json.Marshal(map[string]string{
		"err": err.Error(),
	})

	c.WriteToClient(&WSMessage{
		Type:      MsgErr,
		ClientID:  c.ClientID,
		SessionID: c.SessionID,
		Payload:   payload,
	})
}

func (c *Client) ReadPump(ws *WSManager) {
	defer func() {
		ws.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.WriteErrToClient(fmt.Errorf("inavlid msg format"))
			continue
		}

		ws.Incoming <- msg
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		data, _ := json.Marshal(msg)
		c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := c.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			break
		}
	}
}

const CLIENT_ID_LEN = 6

func generateClinetID() string {
	id := uuid.New()
	return id.String()[:CLIENT_ID_LEN]
}
