package accrace

import (
	"log"
	"time"
)

type CarGap struct {
	CarId int

	Laps           int
	SplinePosition float32 // Track position between 0.0 and 1.0

	GapsAhead  map[int]time.Duration
	GapsBehind map[int]time.Duration
}

func (s *State) updateGaps() {
	for _, a := range s.Cars {
		for _, b := range s.Cars {
			if a.CarId != b.CarId {
				// How far is B ahead of A?
				if (a.IsConnected && b.IsConnected &&
					len(a.currentLapPositionTimes) > 0) && (len(b.currentLapPositionTimes) > 0) {
					// Find when B was last at position of A
					currentPosition := a.currentLapPositionTimes[len(a.currentLapPositionTimes)-1]
					bWasLastAtCurrentPosition, ok := b.lastLocalTimeOfPosition(currentPosition.splinePosition)
					if ok {
						gap := currentPosition.localTime.Sub(bWasLastAtCurrentPosition)
						a.gapsAhead[b.CarId] = gap
						b.gapsBehind[a.CarId] = gap
					}
				} else {
					delete(a.gapsAhead, b.CarId)
					delete(b.gapsBehind, a.CarId)
				}
			}
		}
	}

	s.sendGapUpdates()
}

func (s *State) sendGapUpdates() {
	trackGaps := make([]CarGap, 0, len(s.Cars))
	for _, car := range s.Cars {
		if car.IsConnected {
			trackGaps = append(trackGaps, CarGap{
				CarId:          car.CarId,
				Laps:           car.Laps,
				SplinePosition: car.SplinePosition,
				GapsAhead:      car.gapsAhead,
				GapsBehind:     car.gapsBehind,
			})
		}
	}

	log.Printf("Calculated gaps for %v cars", len(trackGaps))

	if len(trackGaps) > 0 {
		s.TrackGaps = trackGaps
		s.TrackGapUpdates <- trackGaps
	}
}
