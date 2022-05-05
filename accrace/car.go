package accrace

import (
	"sort"
	"time"

	"github.com/geniusdex/broawp/accbroadcast"
)

type Driver struct {
	FirstName string
	LastName  string
	ShortName string
}

type Lap struct {
	LapTime     time.Duration   // Negative duration indicates unknown time
	DriverIndex int             // Driver who drove the lap
	SplitTimes  []time.Duration // Always has 3 elements; negative durations indicate unknown times
	IsValid     bool
}

type positionTime struct {
	splinePosition float32
	localTime      time.Time
}

type carGap struct {
	car     *Car
	timeGap time.Duration
}

type Car struct {
	CarId              int
	IsConnected        bool
	TeamName           string
	RaceNumber         int
	CurrentDriverIndex int
	Drivers            []*Driver
	Gear               int // R = -1,  N = 0,  1 = 1,  2 = 2,  ...
	SpeedKmh           int
	Position           int     // Official P/Q/R position (1-based)
	CupPosition        int     // Official P/Q/R position (1-based)
	TrackPosition      int     // Position on track (1-based)
	SplinePosition     float32 // Track position between 0.0 and 1.0
	Laps               int
	Delta              time.Duration // Realtime delta to best session lap
	BestSessionLap     *Lap
	LastLap            *Lap
	CurrentLap         *Lap

	requireEntryListUpdate     bool
	requireTrackPositionUpdate bool

	lastLapPositionTimes    []*positionTime
	currentLapPositionTimes []*positionTime

	nextOnTrack     *carGap
	previousOnTrack *carGap
}

func NewLap(msg *accbroadcast.MsgLap) *Lap {
	return &Lap{
		LapTime:     msg.LapTime,
		DriverIndex: int(msg.DriverIndex),
		SplitTimes:  msg.SplitTimes,
		IsValid:     !msg.IsInvalid,
	}
}

func NewCar(msg *accbroadcast.MsgEntryListCar) *Car {
	car := &Car{
		CarId:                   int(msg.CarId),
		IsConnected:             false,
		lastLapPositionTimes:    make([]*positionTime, 0),
		currentLapPositionTimes: make([]*positionTime, 0),
	}
	car.UpdateFromEntryList(msg)
	return car
}

func (c *Car) UpdateFromEntryList(msg *accbroadcast.MsgEntryListCar) {
	c.TeamName = msg.TeamName
	c.RaceNumber = int(msg.RaceNumber)
	c.Drivers = make([]*Driver, len(msg.Drivers))
	for i := range msg.Drivers {
		c.Drivers[i] = &Driver{
			FirstName: msg.Drivers[i].FirstName,
			LastName:  msg.Drivers[i].LastName,
			ShortName: msg.Drivers[i].ShortName,
		}
	}

	c.requireEntryListUpdate = false
	c.requireTrackPositionUpdate = false
}

func (c *Car) UpdateFromRealtime(msg *accbroadcast.MsgRealtimeCarUpdate) {
	c.IsConnected = true
	if (len(c.Drivers) != int(msg.DriverCount)) || (int(msg.DriverIndex) >= len(c.Drivers)) {
		c.requireEntryListUpdate = true
	} else {
		c.CurrentDriverIndex = int(msg.DriverIndex)
	}
	c.Gear = msg.Gear
	c.SpeedKmh = int(msg.SpeedKmh)
	c.Position = int(msg.Position)
	c.CupPosition = int(msg.CupPosition)
	if c.TrackPosition != int(msg.TrackPosition) {
		c.requireTrackPositionUpdate = true
	}
	c.TrackPosition = int(msg.TrackPosition)
	c.SplinePosition = msg.SplinePosition
	c.Laps = int(msg.Laps)
	c.Delta = msg.Delta
	c.BestSessionLap = NewLap(msg.BestSessionLap)
	c.LastLap = NewLap(msg.LastLap)
	c.CurrentLap = NewLap(msg.CurrentLap)
	if (c.BestSessionLap.DriverIndex >= len(c.Drivers)) ||
		(c.LastLap.DriverIndex >= len(c.Drivers)) ||
		(c.CurrentLap.DriverIndex >= len(c.Drivers)) {
		c.requireEntryListUpdate = true
	}

	if (len(c.currentLapPositionTimes) > 0) &&
		(c.currentLapPositionTimes[len(c.currentLapPositionTimes)-1].splinePosition >= c.SplinePosition) {
		c.lastLapPositionTimes = c.currentLapPositionTimes
		c.currentLapPositionTimes = make([]*positionTime, 0, cap(c.lastLapPositionTimes))
	}
	c.currentLapPositionTimes = append(c.currentLapPositionTimes, &positionTime{c.SplinePosition, time.Now()})
}

func localTimeOfPositionInLap(splinePosition float32, positionTimes []*positionTime) (time.Time, bool) {
	index := sort.Search(len(positionTimes), func(i int) bool {
		return positionTimes[i].splinePosition >= splinePosition
	})
	if index <= 0 || index >= len(positionTimes) {
		return time.Time{}, false
	}
	last := positionTimes[index-1]
	next := positionTimes[index]
	fraction := (splinePosition - last.splinePosition) / (next.splinePosition / last.splinePosition)
	gapFromLast := next.localTime.Sub(last.localTime) * time.Duration(fraction)
	return last.localTime.Add(gapFromLast), true

}

func (c *Car) lastLocalTimeOfPosition(splinePosition float32) (time.Time, bool) {
	localTime, ok := localTimeOfPositionInLap(splinePosition, c.currentLapPositionTimes)
	if !ok {
		localTime, ok = localTimeOfPositionInLap(splinePosition, c.lastLapPositionTimes)
	}
	return localTime, ok
}
