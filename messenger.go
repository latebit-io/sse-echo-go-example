package main

import (
	"fmt"
	"log"
)

type Broadcaster struct {
	Message map[string]chan string
}

func NewMessenger() *Broadcaster {
	return &Broadcaster{
		Message: make(map[string]chan string),
	}
}

func (m *Broadcaster) Send(eventType, message string) {
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, message)
	log.Println(event)

	for _, ch := range m.Message {
		select {
		case ch <- event:
			// Message successfully sent
		default:
			// Channel is full; consider logging or handling this case
			log.Println("Channel is full, message not sent")
		}
	}

	log.Println("Broadcaster sent message")
}

func (m *Broadcaster) Remove(user string) error {
	delete(m.Message, user)
	return nil
}

func (m *Broadcaster) Add(user string, channel chan string) error {
	m.Message[user] = channel
	return nil
}
