package accrace

import (
	"log"
	"time"

	"github.com/geniusdex/broawp/accbroadcast"
)

type SessionType byte

const (
	PracticeSession        SessionType = 0
	QualifyingSession                  = 4
	SuperpoleSession                   = 9
	RaceSession                        = 10
	HotlapSession                      = 11
	HotstintSession                    = 12
	HotlapSuperpoleSession             = 13
	ReplaySession                      = 14
)

func (st SessionType) ToString() string {
	switch st {
	case PracticeSession:
		return "Practice"
	case QualifyingSession:
		return "Qualifying"
	case SuperpoleSession:
		return "Superpole"
	case RaceSession:
		return "Race"
	case HotlapSession:
		return "Hotlap"
	case HotstintSession:
		return "Hotstint"
	case HotlapSuperpoleSession:
		return "Hotlap Superpole"
	case ReplaySession:
		return "Replay"
	}
	return ""
}

type Car struct {
	CarId int
}

type State struct {
	client       *accbroadcast.Client
	connectionId uint32

	SessionType  SessionType
	FocusedCarId uint16

	Cars map[int]*Car

	SessionTypeUpdates   chan string
	TimeRemainingUpdates chan time.Duration
	PositionUpdates      chan int
	CarUpdates           chan *Car
}

func NewState(client *accbroadcast.Client) *State {
	state := &State{
		client:               client,
		Cars:                 make(map[int]*Car),
		SessionTypeUpdates:   make(chan string, 1024),
		TimeRemainingUpdates: make(chan time.Duration, 1024),
		PositionUpdates:      make(chan int, 1024),
		CarUpdates:           make(chan *Car, 1024),
	}

	go state.handleIncomingMessages()

	return state
}

func (s *State) Close() {

}

func (s *State) Burst() {
	s.SessionTypeUpdates <- s.SessionType.ToString()
	for _, car := range s.Cars {
		s.CarUpdates <- car
	}
}

func (s *State) handleIncomingMessages() {
	for {
		raw, ok := <-s.client.IncomingMessages
		if !ok {
			break
		}

		if msg, ok := raw.(*accbroadcast.MsgRegistrationResult); ok {
			s.handleRegistrationResult(msg)
		} else if msg, ok := raw.(*accbroadcast.MsgEntryList); ok {
			s.handleEntryList(msg)
		} else if msg, ok := raw.(*accbroadcast.MsgEntryListCar); ok {
			s.handleEntryListCar(msg)
		} else if msg, ok := raw.(*accbroadcast.MsgRealtimeUpdate); ok {
			s.handleRealtimeUpdate(msg)
		} else if msg, ok := raw.(*accbroadcast.MsgRealtimeCarUpdate); ok {
			s.handleRealtimeCarUpdate(msg)
		} else {
			log.Printf("Received unhandled message: %#v", raw)
		}
	}
}

func (s *State) handleRegistrationResult(msg *accbroadcast.MsgRegistrationResult) {
	if !msg.IsSuccessful {
		log.Fatalf("Registration failed! %v", msg.ErrorMessage)
	}
	log.Printf("Registered with connectionID %v", msg.ConnectionId)
	s.connectionId = msg.ConnectionId
	s.requestEntryList()
}

func (s *State) handleEntryList(msg *accbroadcast.MsgEntryList) {
	// log.Printf("Received entry list: %#v", msg)
}

func (s *State) handleEntryListCar(msg *accbroadcast.MsgEntryListCar) {
	log.Printf("Car ID %v has race number %v", msg.CarId, msg.RaceNumber)
	car := &Car{
		CarId: int(msg.CarId),
	}
	s.Cars[int(msg.CarId)] = car
	s.CarUpdates <- car
}

func (s *State) handleRealtimeUpdate(msg *accbroadcast.MsgRealtimeUpdate) {
	log.Printf("Time Remaining: %v", msg.SessionEndTime)
	if s.SessionType != SessionType(msg.SessionType) {
		s.SessionType = SessionType(msg.SessionType)
		s.SessionTypeUpdates <- s.SessionType.ToString()
	}
	s.TimeRemainingUpdates <- msg.SessionEndTime
	s.FocusedCarId = uint16(msg.FocusedCarIndex)
}

func (s *State) handleRealtimeCarUpdate(msg *accbroadcast.MsgRealtimeCarUpdate) {
	// log.Printf("Update for car ID %v", msg.CarIndex)
	if msg.CarIndex == s.FocusedCarId {
		log.Printf("Position: %v/%v", msg.CupPosition, len(s.Cars))
		s.PositionUpdates <- int(msg.CupPosition)
	}
}

func (s *State) requestEntryList() {
	if err := s.client.RequestEntryList(s.connectionId); err != nil {
		log.Printf("Failed to request entry list: %v", err)
	}
}
