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

func (f *frontend) broadcastHandler(w http.ResponseWriter, r *http.Request) {
	f.executeTemplate(w, r, "broadcast.html", nil)
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

		case pitEvent := <-f.state.PitEvents:
			f.sendPitEvent("pitEvent", pitEvent)
		}
	}
}

type jsCarGap struct {
	CarId int

	Laps           int
	SplinePosition float32

	GapsAhead_ms  map[int]int
	GapsBehind_ms map[int]int
}

func (f *frontend) sendGaps(msgType string, gapsIn []accrace.CarGap) {
	gapsOut := make([]jsCarGap, 0, len(gapsIn))
	for _, gapIn := range gapsIn {
		gapOut := jsCarGap{
			CarId:          gapIn.CarId,
			Laps:           gapIn.Laps,
			SplinePosition: gapIn.SplinePosition,
			GapsAhead_ms:   make(map[int]int),
			GapsBehind_ms:  make(map[int]int),
		}
		for carId, gap := range gapIn.GapsAhead {
			gapOut.GapsAhead_ms[carId] = int(gap / time.Millisecond)
		}
		for carId, gap := range gapIn.GapsBehind {
			gapOut.GapsBehind_ms[carId] = int(gap / time.Millisecond)
		}
		gapsOut = append(gapsOut, gapOut)
	}
	f.sendMessageToWebSockets(msgType, gapsOut)
}

type jsPitEvent struct {
	Car         *accrace.Car
	OldLocation string
	NewLocation string
}

func (f *frontend) sendPitEvent(msgType string, pitEvent *accrace.PitEvent) {
	locationToString := map[int]string{
		accrace.CarLocationUnknown:  "unknown",
		accrace.CarLocationTrack:    "track",
		accrace.CarLocationPitLane:  "pit_lane",
		accrace.CarLocationPitEntry: "pit_entry",
		accrace.CarLocationPitExit:  "pit_exit",
		accrace.CarLocationPitBox:   "pit_box",
	}

	f.sendMessageToWebSockets(msgType, &jsPitEvent{
		Car:         pitEvent.Car,
		OldLocation: locationToString[pitEvent.OldLocation],
		NewLocation: locationToString[pitEvent.NewLocation],
	})
}
