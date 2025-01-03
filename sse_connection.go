package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

type SSEConnection struct {
	Handle            string
	Context           echo.Context
	CloseChannel      chan struct{}
	MessageChannel    chan string
	ConnectionService *DefaultConnectionService
}

func NewSSEConnection(context echo.Context, handle string, closed chan struct{}, message chan string,
	connectionService *DefaultConnectionService) *SSEConnection {
	return &SSEConnection{
		Context:           context,
		Handle:            handle,
		CloseChannel:      closed,
		MessageChannel:    message,
		ConnectionService: connectionService,
	}
}

func (conn *SSEConnection) Run() error {
	log.Printf("SSE client connected, ip: %v", conn.Context.RealIP())
	r := conn.Context.Response()
	r.Header().Set("Content-Type", "text/event-stream")
	r.Header().Set("Cache-Control", "no-cache")
	r.Header().Set("Connection", "keep-alive")
	conn.ConnectionService.Send("online", fmt.Sprintf("%s", conn.Handle))
	for {
		select {
		case <-conn.Context.Request().Context().Done():
			log.Printf("SSE client disconnected, ip: %v", conn.Context.RealIP())
			conn.ConnectionService.Remove(conn.Handle)
			conn.ConnectionService.Send("offline", fmt.Sprintf("offline - %s", conn.Handle))
			return nil
		case <-conn.CloseChannel:
			_, err := fmt.Fprintf(r, "data: Message: %s\n\n", "connection closed")
			if err != nil {
				log.Println(err.Error())
			}
			r.Flush()
			conn.ConnectionService.Connections.Delete(conn.Handle)
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