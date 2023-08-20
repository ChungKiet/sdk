package websocket

import "fmt"

type RoomBroadCastMsg struct {
	Room string
	Msg  []byte
}

type ClientMessage struct {
	Client *Client
	Msg    []byte
}

type ClientRoom struct {
	Room   string
	Client *Client
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered Clients.
	Clients map[string]*Client

	// Rooms
	Rooms map[string]map[*Client]bool

	// Broadcast receive channel
	Broadcast chan []byte

	// room broadcast recevive channel
	RoomBroadcast chan *RoomBroadCastMsg

	// client message direct
	ClientMsg chan *ClientMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	JoinRoom chan *ClientRoom

	LeaveRoom chan *ClientRoom
}

func newHub() *Hub {
	return &Hub{
		Rooms:         make(map[string]map[*Client]bool),
		RoomBroadcast: make(chan *RoomBroadCastMsg),
		ClientMsg:     make(chan *ClientMessage),
		Broadcast:     make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		JoinRoom:      make(chan *ClientRoom),
		LeaveRoom:     make(chan *ClientRoom),
		Clients:       make(map[string]*Client),
	}
}

func (h *Hub) unregister(clientId string) {
	if client, ok := h.Clients[clientId]; ok {
		delete(h.Clients, clientId)
		client.Close()
		client.LeaveAllRooms()
	}
}

func (h *Hub) leaveRoom(room string, c *Client) {
	delete(c.Hub.Rooms[room], c)
}

func (h *Hub) start() {
	for {
		h.run()
	}
}

func (h *Hub) run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic error ", r)
		}
	}()
	select {
	case client := <-h.Register:
		h.Clients[client.ID] = client
	case client := <-h.Unregister:
		h.unregister(client.ID)
	case cr := <-h.JoinRoom:
		_, exist := h.Rooms[cr.Room]
		if !exist {
			h.Rooms[cr.Room] = map[*Client]bool{}
		}
		h.Rooms[cr.Room][cr.Client] = true
		cr.Client.Rooms[cr.Room] = true
	case lr := <-h.LeaveRoom:
		h.leaveRoom(lr.Room, lr.Client)
		// delete(cr.Client.Rooms, cr.Room)
	case message := <-h.Broadcast:
		for _, client := range h.Clients {
			client.Send(message)
		}
	case roomMsg := <-h.RoomBroadcast:
		for client := range h.Rooms[roomMsg.Room] {
			client.Send(roomMsg.Msg)
		}
	case msg := <-h.ClientMsg:
		msg.Client.Send(msg.Msg)
		// default:
		// 	log.Info("default hub")
	}

}
