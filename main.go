package main

import (
	"bufio"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/html/v2"
	"strings"
	"sync"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	connections := sync.Map{}
	viewsEngine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: viewsEngine,
	})

	online := NewOnline()

	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		user := c.Cookies("user")
		if user == "" {
			userCookie := new(fiber.Cookie)
			userCookie.Name = "user"
			userCookie.Value = utils.UUID()
			user = userCookie.Value
			c.Cookie(userCookie)
		}
		if conn, exists := connections.Load(user); exists {
			if closeFunc, ok := conn.(func()); ok {
				closeFunc()
			}
		}
		online.AddUser(user)
		closeChan := make(chan struct{})
		connections.Store(user, func() {
			close(closeChan)
		})

		c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			fmt.Printf("connected: %s\n", user)
			for {
				select {
				case <-closeChan:
					fmt.Fprintf(w, "data: Message: %s\n\n", "connection closed")
					return
				default:
					fmt.Fprintf(w, "data: Message: %s\n\n", strings.Join(online.GetUsers(), ","))
					err := w.Flush()
					if err != nil {
						connections.Delete(user)
						online.RemoveUser(user)
						fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
						return
					}
					time.Sleep(2 * time.Second)
				}
			}
		})

		return nil
	})

	if err := app.Listen(":666"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
