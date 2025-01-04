package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// main trying to come up with a decent pattern for using sse
func main() {
	//manages channels and messages for client connections
	connectionService := NewDefaultConnectionService()
	app := echo.New()
	// entry point at root
	app.GET("/", func(c echo.Context) error {
		// simulate users logging in
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
		//if a user is already logged in close the connection
		if closed, exists := connectionService.Get(user); exists {
			closed()
		}
		//create new connections
		closeChan := make(chan struct{})
		messageChan := make(chan string, 1)
		connectionService.Add(user, messageChan, closeChan, func() {
			close(closeChan)
			close(messageChan)
		})
		// start sse event loop works with ConnectionService
		connection := NewSSEConnection(c, user, closeChan, messageChan, connectionService)
		return connection.Run()

	})

	// start echo web service
	if err := app.Start(":666"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}