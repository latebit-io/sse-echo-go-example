package main

type ConnectionService interface {
	Send(eventType string, data interface{})
	Add(handle string, message chan string, closed chan struct{}, onclose func())
	Get(handle string) (func(), bool)
	Remove(handle string) bool
}