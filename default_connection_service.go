package main

import (
	"fmt"
	"log"
	"sync"
)

// DefaultConnectionService implements ConnectionService
type DefaultConnectionService struct {
	// Message channel to broadcast messages
	Message map[string]chan string
	// Connection used to limit one connection per client
	Connections sync.Map
}

// NewDefaultConnectionService creates a new DefaultConnectionService
func NewDefaultConnectionService() *DefaultConnectionService {
	return &DefaultConnectionService{
		Message:     make(map[string]chan string),
		Connections: sync.Map{},
	}
}

// Add a new connection to be tracked it accepts a function to implement an onclose callback
func (c *DefaultConnectionService) Add(id string, message chan string, closed chan struct{}, onclose func()) {
	c.Message[id] = message
	c.Connections.Store(id, onclose)
}

// Remove a connection so it is no longer tracked
func (c *DefaultConnectionService) Remove(id string) bool {
	if _, exists := c.Connections.Load(id); exists {
		c.Connections.Delete(id)
		delete(c.Message, id)
		return true
	}

	return false
}

// Get a connection and return its onclose callback
func (c *DefaultConnectionService) Get(id string) (func(), bool) {
	if conn, exists := c.Connections.Load(id); exists {
		if closeFunc, ok := conn.(func()); ok {
			return closeFunc, true
		}
	}
	return nil, false
}

// Send a message with event type to all clients
func (c *DefaultConnectionService) Send(eventType, message string) {
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, message)
	log.Println(event)

	for _, ch := range c.Message {
		select {
		case ch <- event:
			log.Println("Broadcaster sent message")
		default:
			log.Println("Channel is full, message not sent")
		}
	}

}

// Close a connection that will be reopened by same client different device or browser window
func (c *DefaultConnectionService) Close(id string) {
	c.Connections.Delete(id)
}
