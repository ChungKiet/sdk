package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	base_websocket "github.com/goonma/sdk/base/websocket"
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
	websocket.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := websocket.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := websocket.NewClient(w.Hub, conn)

	client.Hub.Register <- client
	if ok {
		client.Hub.JoinRoom <- websocket.ClientRoom{
			Room:   user.UserID,
			Client: client,
		}
	}

	defer func() {
		client.Hub.Unregister <- client
		if ok {
			client.LeaveAllRooms()
		}
	}()

	go client.ReadPump()
	go client.WritePump()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	bcastTicket := time.NewTicker(10 * time.Second)
	defer bcastTicket.Stop()
	for {
		select {
		case <-ticker.C:
			println("num of connections: ", len(client.Hub.Rooms))
			sendData := make([]interface{}, 0)
			sendData = append(sendData, "aaaaa")
			data := base_websocket.WsEvent{
				Event: conn.LocalAddr().Network(),
				Data:  sendData,
			}
			byteData, _ := json.Marshal(data)
			client.Hub.RoomBroadcast <- websocket.RoomBroadCastMsg{
				Room: "test",
				Msg:  byteData,
			}
		case <-bcastTicket.C:
			sendData := make([]interface{}, 0)
			sendData = append(sendData, "aaaaa")
			data := base_websocket.WsEvent{
				Event: "Broadcast",
				Data:  sendData,
			}
			byteData, _ := json.Marshal(data)
			client.Hub.Broadcast <- byteData
		}
	}
	return nil
}

func main() {
	var w Websocket
	w.Initial("price_data", w.WsHandle, w.Consume)
	w.Start()
}
