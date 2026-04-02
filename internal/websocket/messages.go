package websocket

import "encoding/json"

type MessageType string

const (
	MsgInit           MessageType = "init"
	MsgTerminalInput  MessageType = "terminal_input"
	MsgTerminalOutput MessageType = "terminal_output"
	MsgExec           MessageType = "exec"
	MsgErr            MessageType = "error"
)

type WSMessage struct {
	Type      MessageType     `json:"type"`
	ClientID  string          `json:"client_id"`
	SessionID string          `json:"session_id"`
	Payload   json.RawMessage `json:"payload"`
}
