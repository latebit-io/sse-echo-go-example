package main

import (
	"fmt"
	"log"
	"sync"
)

type DefaultConnectionService struct {
	Message     map[string]chan string
	Connections sync.Map
}

func NewDefaultConnectionService() *DefaultConnectionService {
	return &DefaultConnectionService{
		Message:     make(map[string]chan string),
		Connections: sync.Map{},
	}
}

func (c *DefaultConnectionService) Add(handle string, message chan string, closed chan struct{}, onclose func()) {
	c.Message[handle] = message
	c.Connections.Store(handle, onclose)
}

func (c *DefaultConnectionService) Remove(handle string) bool {
	if _, exists := c.Connections.Load(handle); exists {
		c.Connections.Delete(handle)
		delete(c.Message, handle)
		return true
	}

	return false
}

func (c *DefaultConnectionService) Get(handle string) (func(), bool) {
	if conn, exists := c.Connections.Load(handle); exists {
		if closeFunc, ok := conn.(func()); ok {
			return closeFunc, true
		}
	}
	return nil, false
}

func (c *DefaultConnectionService) Send(eventType, message string) {
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, message)
	log.Println(event)

	for _, ch := range c.Message {
		select {
		case ch <- event:
		default:
			log.Println("Channel is full, message not sent")
		}
	}

	log.Println("Broadcaster sent message")
}