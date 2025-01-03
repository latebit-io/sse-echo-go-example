package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

// SSEConnection this type is responsible for handing the SSE connection
type SSEConnection struct {
	Id                string
	Context           echo.Context
	CloseChannel      chan struct{}
	MessageChannel    chan string
	ConnectionService *DefaultConnectionService
}

// NewSSEConnection creates a new connection
func NewSSEConnection(context echo.Context, id string, closed chan struct{}, message chan string,
	connectionService *DefaultConnectionService) *SSEConnection {
	return &SSEConnection{
		Context:           context,
		Id:                id,
		CloseChannel:      closed,
		MessageChannel:    message,
		ConnectionService: connectionService,
	}
}

// Run starts a SSE event loop, it detects connections and disconnects and relays messages
// to the client
func (conn *SSEConnection) Run() error {
	log.Printf("SSE client connected, ip: %v", conn.Context.RealIP())
	r := conn.Context.Response()
	r.Header().Set("Content-Type", "text/event-stream")
	r.Header().Set("Cache-Control", "no-cache")
	r.Header().Set("Connection", "keep-alive")
	conn.ConnectionService.Send("online", fmt.Sprintf("%s", conn.Id))
	for {
		select {
		case <-conn.Context.Request().Context().Done():
			log.Printf("SSE client disconnected, ip: %v", conn.Context.RealIP())
			conn.ConnectionService.Remove(conn.Id)
			conn.ConnectionService.Send("offline", fmt.Sprintf("offline - %s", conn.Id))
			return nil
		case <-conn.CloseChannel:
			_, err := fmt.Fprintf(r, "data: Message: %s\n\n", "connection closed")
			if err != nil {
				log.Println(err.Error())
			}
			r.Flush()
			conn.ConnectionService.Connections.Delete(conn.Id)
			return nil
		case msg := <-conn.MessageChannel:
			_, err := fmt.Fprintf(r, msg)
			if err != nil {
				log.Println(err.Error())
			}
			r.Flush()
		}
	}
}