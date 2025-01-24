package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"online/views"
)

// main trying to come up with a decent pattern for using sse
func main() {
	//manages channels and messages for client connections
	connectionService := NewDefaultConnectionService()
	grid := NewRedisGrid()
	init := true
	if init {
		grid.Init(context.Background())
	}
	app := echo.New()
	app.Use(middleware.Logger())
	app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status}\n",
	}))

	app.Renderer = views.NewHTMLRenderer("views/*.html")
	app.Static("/static/", "./static")
	// entry point at root
	app.GET("/grid-stream", func(c echo.Context) error {
		// simulate users logging in with cookies
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
		connection := NewSSEConnection(c, user, closeChan, messageChan, connectionService, *grid)
		return connection.Run()
	})

	app.GET("/", func(c echo.Context) error {
		data := map[string]interface{}{}
		return c.Render(http.StatusOK, "grid.html", data)
	})

	// start echo web service
	if err := app.Start(":666"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
