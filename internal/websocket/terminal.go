package websocket

import (
	"context"

	"github.com/docker/docker/api/types"
)

type TerminalSession struct {
	// Terminal state
	TerminalResp types.HijackedResponse
	IsConnected  bool

	Input  chan WSTermInputMsg  // stdin to docker
	Output chan WSTermOutputMsg // stdout from docker

	Ctx    context.Context
	Cancel context.CancelFunc
}

func (t *TerminalSession) Start() {
	go t.TermInputPump()
	go t.TermOutputPump()
}

func (t *TerminalSession) TermInputPump() {
	for {
		select {
		case msg, ok := <-t.Input:
			if !ok {
				return
			}

			// push to container
			_, err := t.TerminalResp.Conn.Write([]byte(msg.Input))
			if err != nil {
				return
			}

		case <-t.Ctx.Done():
			return
		}
	}
}

func (t *TerminalSession) TermOutputPump() {
	buf := make([]byte, 4096)

	for {
		select {
		case <-t.Ctx.Done():
			return
		default:
			n, err := t.TerminalResp.Reader.Read(buf)
			if err != nil {
				return
			}

			if n == 0 {
				continue
			}

			output := string(buf[:n])

			msg := WSTermOutputMsg{
				Output: output,
			}

			// non blocking send
			select {
			case t.Output <- msg:
			case <-t.Ctx.Done():
				return
			default:
				
			}
		}
	}
}
