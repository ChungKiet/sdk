package websocket

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

func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ID] = client
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				client.Close()
				client.LeaveAllRooms()
			}
		case cr := <-h.JoinRoom:
			_, exist := h.Rooms[cr.Room]
			if !exist {
				h.Rooms[cr.Room] = map[*Client]bool{}
			}
			h.Rooms[cr.Room][cr.Client] = true
			cr.Client.Rooms[cr.Room] = true
		case cr := <-h.LeaveRoom:
			delete(h.Rooms[cr.Room], cr.Client)
			delete(cr.Client.Rooms, cr.Room)
		case message := <-h.Broadcast:
			for _, client := range h.Clients {
				client.send <- message
			}
		case roomMsg := <-h.RoomBroadcast:
			for client := range h.Rooms[roomMsg.Room] {
				client.Send(roomMsg.Msg)
			}
		case msg := <-h.ClientMsg:
			msg.Client.Send(msg.Msg)
		}
	}

}
