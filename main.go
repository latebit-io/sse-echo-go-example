package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"sync"
)

func main() {
	messenger := NewMessenger()
	connections := sync.Map{}

	app := echo.New()
	online := NewOnline()

	app.GET("/", func(c echo.Context) error {
		log.Printf("SSE client connected, ip: %v", c.RealIP())
		var user string
		userCookie, err := c.Cookie("user")
		if err != nil {
			userCookie = new(http.Cookie)
			userCookie.Name = "user"
			newUUID := uuid.New()
			userCookie.Value = newUUID.String()
			c.SetCookie(userCookie)
		}
		user = userCookie.Value

		if conn, exists := connections.Load(user); exists {
			if closeFunc, ok := conn.(func()); ok {
				closeFunc()
			}
		}

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		online.AddUser(user)
		closeChan := make(chan struct{})
		connections.Store(user, func() {
			close(closeChan)
		})
		messageChan := make(chan string, 1)
		messenger.Add(user, messageChan)
		messenger.Send("online", fmt.Sprintf("%s", user))

		defer func() {
			connections.Delete(user)
			messenger.Remove(user)
			online.RemoveUser(user)
		}()

		for {
			select {
			case <-c.Request().Context().Done():
				log.Printf("SSE client disconnected, ip: %v", c.RealIP())
				messenger.Send("offline", fmt.Sprintf("offline - %s", user))
				return nil
			case <-closeChan:
				_, err := fmt.Fprintf(w, "data: Message: %s\n\n", "connection closed")
				if err != nil {
					log.Println(err.Error())
				}
				w.Flush()
				return nil
			case msg := <-messageChan:
				_, err := fmt.Fprintf(w, msg)
				if err != nil {
					log.Println(err.Error())
				}
				w.Flush()
			}
		}
	})

	if err := app.Start(":666"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
