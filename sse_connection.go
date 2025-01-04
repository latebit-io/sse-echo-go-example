package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

// SSEConnection this type is responsible for handing the SSE connection
type SSEConnection struct {
	// Id identifier for the connection
	Id string
	// Context the web framework context used to write the response
	Context echo.Context
	// CloseChannel This channel is used to manage a single connection per Id
	CloseChannel chan struct{}
	// MessageChannel used to send events to clients
	MessageChannel chan string
	// ConnectionService a pointer to the connection service that manages client channels
	ConnectionService ConnectionService
}

// NewSSEConnection creates a new connection
func NewSSEConnection(context echo.Context, id string, closed chan struct{}, message chan string,
	connectionService ConnectionService) *SSEConnection {
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
			//context is done when the client disconnects
			log.Printf("SSE client disconnected, ip: %v", conn.Context.RealIP())
			conn.ConnectionService.Remove(conn.Id)
			conn.ConnectionService.Send("offline", fmt.Sprintf("offline - %s", conn.Id))
			return nil

		case <-conn.CloseChannel:
			//this is used to close an existing connection, if another connection is opened with the same id
			_, err := fmt.Fprintf(r, "data: Message: %s\n\n", "connection closed")
			if err != nil {
				log.Println(err.Error())
			}
			r.Flush()
			conn.ConnectionService.Close(conn.Id)
			return nil
		case msg := <-conn.MessageChannel:
			//receives messages from broadcaster and sends to the client
			_, err := fmt.Fprintf(r, msg)
			if err != nil {
				log.Println(err.Error())
			}
			r.Flush()
		}
	}
}