package ws

import "github.com/google/uuid"

type BroadcastMessage struct {
	ChatExternalID uuid.UUID
	Data           []byte
}

type Hub struct {
	Clients    map[string]map[*Client]bool
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast: make(chan BroadcastMessage),

		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.Clients[client.ChatExternalID.String()]; !ok {
				h.Clients[client.ChatExternalID.String()] = make(map[*Client]bool)
			}
			h.Clients[client.ChatExternalID.String()][client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ChatExternalID.String()]; ok {
				delete(h.Clients[client.ChatExternalID.String()], client)
				close(client.Send)

				if len(h.Clients[client.ChatExternalID.String()]) == 0 {
					delete(h.Clients, client.ChatExternalID.String())
				}
			}
		case message := <-h.Broadcast:
			clientsInChat := h.Clients[message.ChatExternalID.String()]
			for client := range clientsInChat {
				select {
				case client.Send <- message.Data:

				default:
					close(client.Send)
					delete(h.Clients[message.ChatExternalID.String()], client)
				}
			}
		}
	}
}