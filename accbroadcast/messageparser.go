package accbroadcast

import (
	"fmt"
	"math"
	"time"
)

type messageParser struct {
}

func newMessageParser() *messageParser {
	return &messageParser{}
}

func (mp *messageParser) Parse(buf []byte) (interface{}, error) {
	br := newBufferReader(buf)
	switch br.ReadByte() {
	case inRegistrationResult:
		return mp.parseRegistrationResult(br)
	case inEntryList:
		return mp.parseEntryList(br)
	case inEntryListCar:
		return mp.parseEntryListCar(br)
	case inRealtimeUpdate:
		return mp.parseRealtimeUpdate(br)
	case inRealtimeCarUpdate:
		return mp.parseRealtimeCarUpdate(br)
	case inTrackData:
		return mp.parseTrackData(br)
	case inBroadcastingEvent:
		return mp.parseBroadcastingEvent(br)
	default:
		return nil, fmt.Errorf("unknown message type %v", int(buf[0]))
	}
}

func (mp *messageParser) parseRegistrationResult(br *bufferReader) (*MsgRegistrationResult, error) {
	msg := &MsgRegistrationResult{
		ConnectionId:  br.ReadUint32(),
		IsSuccessful:  br.ReadBool(),
		AllowCommands: br.ReadBool(),
		ErrorMessage:  br.ReadString(),
	}
	return msg, nil
}

func (mp *messageParser) parseEntryList(br *bufferReader) (*MsgEntryList, error) {
	msg := &MsgEntryList{
		ConnectionId: br.ReadUint32(),
		CarIds:       make([]uint16, br.ReadUint16()),
	}
	for i := range msg.CarIds {
		msg.CarIds[i] = br.ReadUint16()
	}
	return msg, nil
}

func (mp *messageParser) parseEntryListCar(br *bufferReader) (*MsgEntryListCar, error) {
	msg := &MsgEntryListCar{
		CarId:              br.ReadUint16(),
		CarModelType:       br.ReadByte(),
		TeamName:           br.ReadString(),
		RaceNumber:         br.ReadUint32(),
		CupCategory:        br.ReadByte(),
		CurrentDriverIndex: br.ReadByte(),
		Nationality:        br.ReadUint16(),
		Drivers:            make([]*MsgEntryListCarDriver, br.ReadByte()),
	}
	for i := range msg.Drivers {
		msg.Drivers[i] = &MsgEntryListCarDriver{
			FirstName:   br.ReadString(),
			LastName:    br.ReadString(),
			ShortName:   br.ReadString(),
			Category:    br.ReadByte(),
			Nationality: br.ReadUint16(),
		}
	}
	return msg, nil
}

func (mp *messageParser) parseRealtimeUpdate(br *bufferReader) (*MsgRealtimeUpdate, error) {
	msg := &MsgRealtimeUpdate{
		EventIndex:      br.ReadUint16(),
		SessionIndex:    br.ReadUint16(),
		SessionType:     br.ReadByte(),
		SessionPhase:    br.ReadByte(),
		SessionTime:     time.Duration(br.ReadFloat32()) * time.Millisecond,
		SessionEndTime:  time.Duration(br.ReadFloat32()) * time.Millisecond,
		FocusedCarIndex: br.ReadUint32(),
		ActiveCameraSet: br.ReadString(),
		ActiveCamera:    br.ReadString(),
		CurrentHudPage:  br.ReadString(),
		IsReplayPlaying: br.ReadBool(),
	}
	if msg.IsReplayPlaying {
		msg.ReplaySessionTime = time.Duration(br.ReadFloat32()) * time.Millisecond
		msg.ReplayRemainingTime = time.Duration(br.ReadFloat32()) * time.Millisecond
	}
	msg.TimeOfDay = time.Duration(br.ReadFloat32()) * time.Millisecond
	msg.AmbientTemp = br.ReadByte()
	msg.TrackTemp = br.ReadByte()
	msg.Clouds = br.ReadByte()
	msg.RainLevel = br.ReadByte()
	msg.Wetness = br.ReadByte()

	var err error
	msg.BestSessionLap, err = mp.parseLap(br)

	return msg, err
}

func (mp *messageParser) parseRealtimeCarUpdate(br *bufferReader) (*MsgRealtimeCarUpdate, error) {
	msg := &MsgRealtimeCarUpdate{
		CarIndex:       br.ReadUint16(),
		DriverIndex:    br.ReadUint16(),
		DriverCount:    br.ReadByte(),
		Gear:           int(br.ReadByte()) - 2,
		WorldPosX:      br.ReadFloat32(),
		WorldPosY:      br.ReadFloat32(),
		Yaw:            br.ReadFloat32(),
		CarLocation:    br.ReadByte(),
		SpeedKmh:       br.ReadUint16(),
		Position:       br.ReadUint16(),
		CupPosition:    br.ReadUint16(),
		TrackPosition:  br.ReadUint16(),
		SplinePosition: br.ReadFloat32(),
		Laps:           br.ReadUint16(),
		Delta:          time.Duration(br.ReadInt32()) * time.Millisecond,
	}
	var err error
	if msg.BestSessionLap, err = mp.parseLap(br); err != nil {
		return msg, err
	}
	if msg.LastLap, err = mp.parseLap(br); err != nil {
		return msg, err
	}
	if msg.CurrentLap, err = mp.parseLap(br); err != nil {
		return msg, err
	}
	return msg, nil
}

func (mp *messageParser) parseTrackData(br *bufferReader) (*MsgTrackData, error) {
	msg := &MsgTrackData{}
	return msg, nil
}

func (mp *messageParser) parseBroadcastingEvent(br *bufferReader) (*MsgBroadcastingEvent, error) {
	msg := &MsgBroadcastingEvent{}
	return msg, nil
}

func (mp *messageParser) readOptionalDuration(br *bufferReader) time.Duration {
	duration := br.ReadInt32()
	if duration == math.MaxInt32 {
		return -1 * time.Millisecond
	} else {
		return time.Duration(duration) * time.Millisecond
	}
}

func (mp *messageParser) parseLap(br *bufferReader) (*MsgLap, error) {
	msg := &MsgLap{
		LapTime:     mp.readOptionalDuration(br),
		CarIndex:    br.ReadUint16(),
		DriverIndex: br.ReadUint16(),
		SplitTimes:  make([]time.Duration, 3),
	}
	for i := range msg.SplitTimes {
		msg.SplitTimes[i] = -1 * time.Millisecond
	}
	nrSplits := int(br.ReadByte())
	for i := 0; i < nrSplits; i++ {
		msg.SplitTimes[i] = mp.readOptionalDuration(br)
	}
	msg.IsInvalid = br.ReadBool()
	msg.IsValidForBest = br.ReadBool()
	msg.IsOutLap = br.ReadBool()
	msg.IsInLap = br.ReadBool()

	return msg, nil
}
