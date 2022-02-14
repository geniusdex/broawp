package accbroadcast

const (
	protocolVersion = 4

	// Outbound message types
	outRegisterCommandApplication   = 1
	outUnregisterCommandApplication = 9
	outRequestEntryList             = 10
	outRequestTrackData             = 11
	outChangeHudPage                = 49
	outChangeFocus                  = 50
	outInstantReplayRequest         = 51
	outPlayManualReplayHighlight    = 52
	outSaveManualReplayHighlight    = 60

	// Inbound message types
	inRegistrationResult = 1
	inRealtimeUpdate     = 2
	inRealtimeCarUpdate  = 3
	inEntryList          = 4
	inEntryListCar       = 6
	inTrackData          = 5
	inBroadcastingEvent  = 7
)
