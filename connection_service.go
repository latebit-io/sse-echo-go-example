package main

type ConnectionService interface {
	Send(eventType string, data string)
	Add(id string, message chan string, closed chan struct{}, onclose func())
	Get(id string) (func(), bool)
	Remove(id string) bool
	Close(id string)
}