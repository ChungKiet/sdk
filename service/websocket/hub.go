package websocket

type RoomBroadCastMsg struct {
	Room string
	Msg  []byte
}

type ClientRoom struct {
	Room   string
	Client *Client
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Rooms
	Rooms map[string]map[*Client]bool

	// Broadcast receive channel
	Broadcast chan []byte

	// room broadcast recevive channel

	RoomBroadcast chan RoomBroadCastMsg

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	JoinRoom chan ClientRoom

	LeaveRoom chan ClientRoom
}

func newHub() *Hub {
	return &Hub{
		Rooms:         make(map[string]map[*Client]bool),
		RoomBroadcast: make(chan RoomBroadCastMsg),
		Broadcast:     make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		JoinRoom:      make(chan ClientRoom),
		LeaveRoom:     make(chan ClientRoom),
		Clients:       make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
				client.LeaveAllRooms()
			}
		case cr := <-h.JoinRoom:
			_, exist := h.Rooms[cr.Room]
			if !exist {
				h.Rooms[cr.Room] = map[*Client]bool{}
			}
			h.Rooms[cr.Room][cr.Client] = true
			cr.Client.Rooms = append(cr.Client.Rooms, cr.Room)
		case cr := <-h.LeaveRoom:
			delete(h.Rooms[cr.Room], cr.Client)
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.send <- message:
				default:
					client.LeaveAllRooms()
					close(client.send)
					delete(h.Clients, client)
				}
			}
		case roomMsg := <-h.RoomBroadcast:
			for client := range h.Rooms[roomMsg.Room] {
				client.Send(roomMsg.Msg)
			}
		}
	}
}
