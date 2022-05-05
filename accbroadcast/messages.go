package accbroadcast

import "time"

const (
	CupCategoryPro      = 0
	CupCategoryProAm    = 1
	CupCategoryAm       = 2
	CupCategorySilver   = 3
	CupCategoryNational = 4

	DriverCategoryPlatinum = 3
	DriverCategoryGold     = 2
	DriverCategorySilver   = 1
	DriverCategoryBronze   = 0

	CarLocationUnknown  = 0
	CarLocationTrack    = 1
	CarLocationPitLane  = 2
	CarLocationPitEntry = 3
	CarLocationPitExit  = 4
)

type MsgRegistrationResult struct {
	ConnectionId  uint32
	IsSuccessful  bool
	AllowCommands bool
	ErrorMessage  string
}

type MsgEntryList struct {
	ConnectionId uint32
	CarIds       []uint16
}

type MsgEntryListCar struct {
	CarId              uint16
	CarModelType       byte
	TeamName           string
	RaceNumber         uint32
	CupCategory        byte
	CurrentDriverIndex byte
	Nationality        uint16
	Drivers            []*MsgEntryListCarDriver
}

type MsgEntryListCarDriver struct {
	FirstName   string
	LastName    string
	ShortName   string
	Category    byte
	Nationality uint16
}

type MsgRealtimeUpdate struct {
	EventIndex          uint16
	SessionIndex        uint16
	SessionType         byte
	SessionPhase        byte
	SessionTime         time.Duration
	SessionEndTime      time.Duration
	FocusedCarIndex     uint32
	ActiveCameraSet     string
	ActiveCamera        string
	CurrentHudPage      string
	IsReplayPlaying     bool
	ReplaySessionTime   time.Duration
	ReplayRemainingTime time.Duration
	TimeOfDay           time.Duration
	AmbientTemp         byte
	TrackTemp           byte
	Clouds              byte
	RainLevel           byte
	Wetness             byte
	BestSessionLap      *MsgLap
}

type MsgRealtimeCarUpdate struct {
	CarIndex       uint16
	DriverIndex    uint16
	DriverCount    byte
	Gear           int // R = -1,  N = 0,  1 = 1,  2 = 2,  ...
	WorldPosX      float32
	WorldPosY      float32
	Yaw            float32
	CarLocation    byte
	SpeedKmh       uint16
	Position       uint16  // Official P/Q/R position (1 based)
	CupPosition    uint16  // Official P/Q/R position (1 based)
	TrackPosition  uint16  // Position on track (1-based)
	SplinePosition float32 // Track position between 0.0 and 1.0
	Laps           uint16
	Delta          time.Duration // Realtime delta to best session lap
	BestSessionLap *MsgLap
	LastLap        *MsgLap
	CurrentLap     *MsgLap
}

type MsgTrackData struct {
}

type MsgBroadcastingEvent struct {
}

type MsgLap struct {
	LapTime        time.Duration // Negative duration indicates unknown time
	CarIndex       uint16
	DriverIndex    uint16
	SplitTimes     []time.Duration // Extended to include 3 times; negative durations indicate unknown times
	IsInvalid      bool
	IsValidForBest bool
	IsOutLap       bool
	IsInLap        bool
}
