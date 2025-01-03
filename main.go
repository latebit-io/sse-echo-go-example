package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func main() {
	connectionService := NewDefaultConnectionService()
	app := echo.New()

	app.GET("/", func(c echo.Context) error {
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
		if closed, exists := connectionService.Get(user); exists {
			closed()
		}

		closeChan := make(chan struct{})
		messageChan := make(chan string, 1)
		connectionService.Add(user, messageChan, closeChan, func() {
			close(closeChan)
			close(messageChan)
		})

		connection := NewSSEConnection(c, user, closeChan, messageChan, connectionService)
		return connection.Run()

	})

	if err := app.Start(":666"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}