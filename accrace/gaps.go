package accrace

import (
	"log"
	"sort"
	"time"
)

type CarGap struct {
	CarId int

	CarIdAhead int
	GapAhead   time.Duration

	CarIdBehind int
	GapBehind   time.Duration
}

func (s *State) updateGapsEvery(interval time.Duration) {
	for range time.Tick(interval) {
		s.updateGaps()
	}
}

func (s *State) updateGaps() {
	cars := make([]*Car, 0, len(s.Cars))
	for _, car := range s.Cars {
		cars = append(cars, car)
	}

	sort.Slice(cars, func(i, j int) bool {
		if cars[i].IsConnected != cars[j].IsConnected {
			return cars[i].IsConnected
		}
		return cars[i].SplinePosition < cars[j].SplinePosition
	})

	if len(cars) >= 2 {
		for currentIndex, current := range cars {
			nextIndex := (currentIndex + 1) % len(cars)
			next := cars[nextIndex]

			if !next.IsConnected {
				nextIndex = 0
				next = cars[nextIndex]
			}

			if (len(current.currentLapPositionTimes) > 0) && (len(next.currentLapPositionTimes) > 0) {
				// Find when the next car was last at my position
				currentPosition := current.currentLapPositionTimes[len(current.currentLapPositionTimes)-1]
				nextWasLastAtCurrentPosition, ok := next.lastLocalTimeOfPosition(currentPosition.splinePosition)
				if ok {
					gap := currentPosition.localTime.Sub(nextWasLastAtCurrentPosition)
					current.nextOnTrack = &carGap{next, gap}
					next.previousOnTrack = &carGap{current, gap}
				}
			}

			if nextIndex == 0 {
				break
			}
		}
	}

	s.sendGapUpdates()
}

func (s *State) sendGapUpdates() {
	trackGaps := make([]CarGap, 0, len(s.Cars))
	for _, car := range s.Cars {
		if (car.nextOnTrack != nil) && (car.previousOnTrack != nil) {
			trackGaps = append(trackGaps, CarGap{
				CarId:       car.CarId,
				CarIdAhead:  car.nextOnTrack.car.CarId,
				GapAhead:    car.nextOnTrack.timeGap,
				CarIdBehind: car.previousOnTrack.car.CarId,
				GapBehind:   car.previousOnTrack.timeGap,
			})
		}
	}

	log.Printf("Calculated gaps for %v cars", len(trackGaps))

	if len(trackGaps) > 0 {
		s.TrackGaps = trackGaps
		s.TrackGapUpdates <- trackGaps
	}
}
