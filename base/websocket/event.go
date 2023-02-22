package websocket

type WsEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
