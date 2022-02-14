package frontend

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func (f *frontend) indexHandler(w http.ResponseWriter, r *http.Request) {
	if len(strings.Trim(r.URL.Path, "/")) > 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "index.html", nil)
}

func (f *frontend) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	err := f.websockets.AddConnection(w, r)
	if err != nil {
		log.Panicf("Failed to create websocket: %v", err)
	}

	f.state.Burst()
}

func (f *frontend) sendMessageToWebSockets(msgType string, data interface{}) {
	jsonMsg, err := json.Marshal(map[string]interface{}{"type": msgType, "data": data})
	if err != nil {
		log.Printf("cannot marshal message as JSON: %v", err)
	}

	f.websockets.SendTextMessage(jsonMsg)
}

func (f *frontend) sendWebSocketUpdates() {
	log.Printf("Sending live state updates to all websockets")

	defer func() {
		f.websockets.Close()
	}()

	for {
		select {
		case sessionType := <-f.state.SessionTypeUpdates:
			f.sendMessageToWebSockets("sessionType", sessionType)

		case remaining := <-f.state.TimeRemainingUpdates:
			f.sendMessageToWebSockets("timeRemaining_ms", remaining/time.Millisecond)

		case position := <-f.state.PositionUpdates:
			f.sendMessageToWebSockets("position", position)

		case car := <-f.state.CarUpdates:
			f.sendMessageToWebSockets("car", car)
		}
	}
}
