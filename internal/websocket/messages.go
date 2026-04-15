package websocket

import "encoding/json"

type MessageType string

const (
	MsgInit           MessageType = "init"
	MsgTerminalInput  MessageType = "term_input"
	MsgTerminalOutput MessageType = "term_output"
	MsgTerminalResize MessageType = "resize"
	MsgErr            MessageType = "error"
)

type WSMessage struct {
	Type      MessageType     `json:"type"`
	ClientID  string          `json:"client_id"`
	SessionID string          `json:"session_id"`
	Payload   json.RawMessage `json:"payload"`
}

type WSTermInputMsg struct {
	Input string `json:"input"`
}

type WSTermOutputMsg struct {
	Output string `json:"output"`
}
