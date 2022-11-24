package main

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/goonma/sdk/jwt"
	"github.com/goonma/sdk/service/websocket"
	"github.com/labstack/echo/v4"
)

type Websocket struct {
	websocket.Websocket
}

func (w *Websocket) Consume(msg *message.Message) error {
	return nil
}

func (w *Websocket) WsHandle(c echo.Context) error {
	user, ok := c.Get("user").(*jwt.CustomClaims)

	conn, err := websocket.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := &websocket.Client{
		Hub:  w.Hub,
		Conn: conn,
		Send: make(chan []byte),
	}

	client.Hub.Register <- client
	if ok {
		client.Hub.JoinRoom <- websocket.ClientRoom{
			Room:   fmt.Sprintf("user-%s", user.UserID),
			Client: client,
		}
	}

	defer func() {
		client.Hub.Unregister <- client
		if ok {
			client.Hub.LeaveRoom <- websocket.ClientRoom{
				Room:   fmt.Sprintf("user-%s", user.UserID),
				Client: client,
			}
		}
	}()

	go client.ReadPump()
	go client.WritePump()

	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-ticker.C:
		client.Send <- []byte("Hello")
	}
	return nil
}

func main() {
	var w Websocket
	w.Initial("price_data", w.WsHandle, w.Consume)
	w.Start()
}
