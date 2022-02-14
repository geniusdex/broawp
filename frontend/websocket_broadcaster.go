package frontend

import (
	"log"
	"net/http"
)

type webSocketBroadcaster struct {
	clients      map[string]webSocketWriter
	addClient    chan webSocketWriter
	removeClient chan webSocketWriter
	messages     chan []byte
	close        chan bool
}

func newWebSocketBroadcaster() *webSocketBroadcaster {
	b := &webSocketBroadcaster{
		clients:      make(map[string]webSocketWriter),
		addClient:    make(chan webSocketWriter, 8),
		removeClient: make(chan webSocketWriter, 8),
		messages:     make(chan []byte, 8),
		close:        make(chan bool, 1),
	}

	go b.processMessages()

	return b
}

func (b *webSocketBroadcaster) AddConnection(w http.ResponseWriter, r *http.Request) error {
	ws, err := newWriteOnlyWebSocket(w, r)
	if err != nil {
		return err
	}

	b.addClient <- ws
	return nil
}

func (b *webSocketBroadcaster) RemoveConnection(ws webSocketWriter) {
	b.removeClient <- ws
}

func (b *webSocketBroadcaster) SendTextMessage(data []byte) {
	b.messages <- data
}

func (b *webSocketBroadcaster) Close() {
	b.close <- true
}

func (b *webSocketBroadcaster) processMessages() {
	for {
		select {
		case ws := <-b.addClient:
			log.Printf("Adding client %s", ws.Name())
			b.clients[ws.Name()] = ws

		case ws := <-b.removeClient:
			log.Printf("Removing client %s", ws.Name())
			delete(b.clients, ws.Name())
			ws.Close()

		case msg := <-b.messages:
			for _, ws := range b.clients {
				if err := ws.WriteTextMessage(msg); err != nil {
					log.Printf("Deleting client %s", ws.Name())
					delete(b.clients, ws.Name())
					ws.Close()
				}
			}

		case <-b.close:
			for _, ws := range b.clients {
				ws.Close()
			}
			return
		}
	}
}
