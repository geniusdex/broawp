package frontend

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/geniusdex/broawp/accrace"
)

func (f *frontend) indexHandler(w http.ResponseWriter, r *http.Request) {
	if len(strings.Trim(r.URL.Path, "/")) > 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "index.html", nil)
}

func (f *frontend) disconnectHandler(w http.ResponseWriter, r *http.Request) {
	f.state.Close()
	f.executeTemplate(w, r, "index.html", nil)
}

func (f *frontend) overlayHandler(w http.ResponseWriter, r *http.Request) {
	f.executeTemplate(w, r, "overlay.html", nil)
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

		case car := <-f.state.FocusedCarUpdates:
			f.sendMessageToWebSockets("focusedCar", car)

		case position := <-f.state.PositionUpdates:
			f.sendMessageToWebSockets("position", position)

		case car := <-f.state.CarUpdates:
			f.sendMessageToWebSockets("car", car)

		case laptime := <-f.state.BestLapUpdates:
			f.sendMessageToWebSockets("bestlaptime_ms", laptime/time.Millisecond)

		case laptime := <-f.state.LastLapUpdates:
			f.sendMessageToWebSockets("lastlaptime_ms", laptime/time.Millisecond)

		case laptime := <-f.state.CurrentLapUpdates:
			f.sendMessageToWebSockets("currentlaptime_ms", laptime/time.Millisecond)

		case delta := <-f.state.LapDeltaUpdates:
			f.sendMessageToWebSockets("lapdelta_ms", delta/time.Millisecond)

		case gaps := <-f.state.TrackGapUpdates:
			f.sendGaps("trackGaps", gaps)
		}
	}
}

type jsCarGap struct {
	CarId int

	CarIdAhead  int
	GapAhead_ms int

	CarIdBehind  int
	GapBehind_ms int
}

func (f *frontend) sendGaps(msgType string, gapsIn []accrace.CarGap) {
	gapsOut := make([]jsCarGap, 0, len(gapsIn))
	for _, gap := range gapsIn {
		gapsOut = append(gapsOut, jsCarGap{
			CarId:        gap.CarId,
			CarIdAhead:   gap.CarIdAhead,
			GapAhead_ms:  int(gap.GapAhead) / int(time.Millisecond),
			CarIdBehind:  gap.CarIdBehind,
			GapBehind_ms: int(gap.GapBehind) / int(time.Millisecond),
		})
	}
	f.sendMessageToWebSockets(msgType, gapsOut)
}
