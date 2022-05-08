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

const (
	entryListRequestTimeout time.Duration = 1 * time.Second
	gapsCalculationPeriod                 = 1 * time.Second
	maintenancePeriod                     = 10 * time.Second
	carUpdateTimeout                      = 10 * time.Second
)

type State struct {
	client               *accbroadcast.Client
	connectionId         uint32
	lastEntryListRequest time.Time

	SessionType        SessionType
	TimeRemaining      time.Duration
	FocusedCarPosition int
	FocusedCarId       int
	FocusedBestLap     time.Duration
	FocusedLastLap     time.Duration
	FocusedCurrentLap  time.Duration
	FocusedLapDelta    time.Duration
	TrackGaps          []CarGap

	Cars map[int]*Car

	SessionTypeUpdates   chan string
	TimeRemainingUpdates chan time.Duration
	FocusedCarUpdates    chan *Car
	PositionUpdates      chan int
	CarUpdates           chan *Car
	BestLapUpdates       chan time.Duration
	LastLapUpdates       chan time.Duration
	CurrentLapUpdates    chan time.Duration
	LapDeltaUpdates      chan time.Duration
	TrackGapUpdates      chan []CarGap
}

func NewState(client *accbroadcast.Client) *State {
	state := &State{
		client:               client,
		FocusedBestLap:       -1 * time.Second,
		FocusedLastLap:       -1 * time.Second,
		FocusedCurrentLap:    -1 * time.Second,
		Cars:                 make(map[int]*Car),
		SessionTypeUpdates:   make(chan string, 1024),
		TimeRemainingUpdates: make(chan time.Duration, 1024),
		FocusedCarUpdates:    make(chan *Car, 1024),
		PositionUpdates:      make(chan int, 1024),
		CarUpdates:           make(chan *Car, 1024),
		BestLapUpdates:       make(chan time.Duration, 1024),
		LastLapUpdates:       make(chan time.Duration, 1024),
		CurrentLapUpdates:    make(chan time.Duration, 1024),
		LapDeltaUpdates:      make(chan time.Duration, 1024),
		TrackGapUpdates:      make(chan []CarGap, 1024),
	}

	go state.handleIncomingMessages()
	go state.updateGapsEvery(gapsCalculationPeriod)
	go state.performMaintenanceEvery(maintenancePeriod)

	return state
}

func (s *State) Close() {
	s.client.Unregister()
}

func (s *State) Burst() {
	s.SessionTypeUpdates <- s.SessionType.ToString()
	s.TimeRemainingUpdates <- s.TimeRemaining
	if car, ok := s.Cars[s.FocusedCarId]; ok {
		s.FocusedCarUpdates <- car
	}
	s.PositionUpdates <- s.FocusedCarPosition
	for _, car := range s.Cars {
		s.CarUpdates <- car
	}
	s.BestLapUpdates <- s.FocusedBestLap
	s.LastLapUpdates <- s.FocusedLastLap
	s.CurrentLapUpdates <- s.FocusedCurrentLap
	s.LapDeltaUpdates <- s.FocusedLapDelta
	s.TrackGapUpdates <- s.TrackGaps
}

func (s *State) setSessionType(sessionType SessionType) {
	if s.SessionType != sessionType {
		s.SessionType = sessionType
		s.SessionTypeUpdates <- sessionType.ToString()
	}
}

func (s *State) setTimeRemaining(timeRemaining time.Duration) {
	if s.TimeRemaining != timeRemaining {
		s.TimeRemaining = timeRemaining
		s.TimeRemainingUpdates <- timeRemaining
	}
}

func (s *State) setFocusedCarPosition(position int) {
	if s.FocusedCarPosition != position {
		s.FocusedCarPosition = position
		s.PositionUpdates <- position
	}
}

func (s *State) setFocusedCarId(carId int) {
	if s.FocusedCarId != carId {
		s.FocusedCarId = carId
		if car, ok := s.Cars[carId]; ok {
			s.FocusedCarUpdates <- car
		}
	}
}

func (s *State) setBestLap(lapTime time.Duration) {
	if s.FocusedBestLap != lapTime {
		s.FocusedBestLap = lapTime
		s.BestLapUpdates <- lapTime
	}
}

func (s *State) setLastLap(lapTime time.Duration) {
	if s.FocusedLastLap != lapTime {
		s.FocusedLastLap = lapTime
		s.LastLapUpdates <- lapTime
	}
}

func (s *State) setCurrentLap(lapTime time.Duration) {
	if s.FocusedCurrentLap != lapTime {
		s.FocusedCurrentLap = lapTime
		s.CurrentLapUpdates <- lapTime
	}
}

func (s *State) setLapDelta(delta time.Duration) {
	if s.FocusedLapDelta != delta {
		s.FocusedLapDelta = delta
		s.LapDeltaUpdates <- delta
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
	var carIdIsPresent = make(map[int]bool)
	for _, carId := range msg.CarIds {
		carIdIsPresent[int(carId)] = true
	}

	for carId, car := range s.Cars {
		car.IsConnected = carIdIsPresent[carId]
	}
}

func (s *State) handleEntryListCar(msg *accbroadcast.MsgEntryListCar) {
	// log.Printf("Car ID %v has race number %v", msg.CarId, msg.RaceNumber)
	carId := int(msg.CarId)
	if car, ok := s.Cars[carId]; ok {
		car.UpdateFromEntryList(msg)
		if car.requireEntryListUpdate {
			s.requestEntryList()
		}
	} else {
		s.Cars[carId] = NewCar(msg)
	}
	s.CarUpdates <- s.Cars[carId]
}

func (s *State) handleRealtimeUpdate(msg *accbroadcast.MsgRealtimeUpdate) {
	log.Printf("Time Remaining: %v", msg.SessionEndTime)
	s.setSessionType(SessionType(msg.SessionType))
	s.setTimeRemaining(msg.SessionEndTime)
	s.setFocusedCarId(int(msg.FocusedCarIndex))
}

func (s *State) handleRealtimeCarUpdate(msg *accbroadcast.MsgRealtimeCarUpdate) {
	// log.Printf("Update for car ID %v", msg.CarIndex)
	if int(msg.CarIndex) == s.FocusedCarId {
		log.Printf("Position: %v/%v", msg.CupPosition, len(s.Cars))
		s.setFocusedCarPosition(int(msg.CupPosition))
		s.setBestLap(msg.BestSessionLap.LapTime)
		s.setLastLap(msg.LastLap.LapTime)
		s.setCurrentLap(msg.CurrentLap.LapTime)
		s.setLapDelta(msg.Delta)
	}
	carId := int(msg.CarIndex)
	if car, ok := s.Cars[carId]; ok {
		car.UpdateFromRealtime(msg)
	} else {
		s.requestEntryList()
	}
}

func (s *State) requestEntryList() {
	now := time.Now()
	if (now.Sub(s.lastEntryListRequest)) > entryListRequestTimeout {
		s.lastEntryListRequest = now
		if err := s.client.RequestEntryList(s.connectionId); err != nil {
			log.Printf("Failed to request entry list: %v", err)
		}
	}
}

func (s *State) performMaintenanceEvery(interval time.Duration) {
	for range time.Tick(interval) {
		s.performMaintenance()
	}
}

func (s *State) performMaintenance() {
	now := time.Now()
	requireEntryListUpdate := false
	for _, car := range s.Cars {
		if now.Sub(car.lastUpdate) > carUpdateTimeout {
			requireEntryListUpdate = true
		}
	}
	if requireEntryListUpdate {
		s.requestEntryList()
	}
}
