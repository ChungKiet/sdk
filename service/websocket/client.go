package websocket

import (
	"time"

	"github.com/google/uuid"

	"github.com/goonma/sdk/log"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID     string
	UserId string

	Hub *Hub

	Rooms map[string]bool

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	isClosed bool
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Hub:      hub,
		Rooms:    make(map[string]bool),
		Conn:     conn,
		send:     make(chan []byte, 256),
		isClosed: false,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump(handleAction func(msg []byte), handleError func(cli *Client)) {
	defer func() {
		c.Hub.unregister(c.ID)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Error(err.Error(), "Websocket Read Pump")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error(err.Error(), "Websocket Read Pump")
			}
			handleError(c)
			break
		}

		handleAction(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Hub.unregister(c.ID)
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Error("Websocket send close message", err.Error())
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Error(err.Error(), "Websocket Write Pump")
				return
			}
			_, err = w.Write(message)

			// Add queued chat messages to the current websocket message.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write(newline)
			// 	w.Write(<-c.send)
			// }

			if err != nil {
				log.Error("Websocket send message error ", err)
				return
			}
			err = w.Close()
			if err != nil {
				log.Error("Websocket flush message error ", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error(err.Error(), "Websocket ping error")
				return
			}
		}
	}
}

func (c *Client) LeaveAllRooms() {
	for room := range c.Rooms {
		// log.Info("room ", room)
		c.Hub.leaveRoom(room, c)
		// c.Hub.LeaveRoom <- &ClientRoom{
		// 	Client: c,
		// 	Room:   room,
		// }
	}
	c.Rooms = make(map[string]bool)
}

func (c *Client) Send(msg []byte) {
	if !c.isClosed {
		c.send <- msg
	} else {
		close(c.send)
	}
}

func (c *Client) Close() {
	c.isClosed = true
}

//// serveWs handles websocket requests from the peer.
// func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
// 	client.hub.register <- client

// 	// Allow collection of memory referenced by the caller by doing all work in
// 	// new goroutines.
// 	go client.writePump()
// 	go client.readPump()
// }
