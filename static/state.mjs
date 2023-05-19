export class State
{
    #cars = {};
    #sessionType = 'Practice';

    setCar(car)
    {
        this.#cars[car.CarId] = car;
    }

    car(carId)
    {
        return this.#cars[carId];
    }

    carsSortedOnMostLaps()
    {
        return Object.values(this.#cars).sort((a, b) => b.ExactLaps - a.ExactLaps);
    }

    carsSortedOnBestLap()
    {
        return Object.values(this.#cars).sort(function(a, b)
        {
            if (a.BestSessionLap.LapTime > 0 && b.BestSessionLap.LapTime > 0)
                return a.BestSessionLap.LapTime - b.BestSessionLap.LapTime;
            else if (a.BestSessionLap.LapTime > 0)
                return -1;
            else if (b.BestSessionLap.LapTime > 0)
                return 1;
            else
                return a.Position - b.Position;
        });
    }

    carsSortedOnTrackPosition()
    {
        return Object.values(this.#cars).sort((a, b) => a.SplinePosition - b.SplinePosition);
    }

    setSessionType(sessionType)
    {
        this.#sessionType = sessionType;
    }

    sessionTargetIsLaps()
    {
        return (this.#sessionType == 'Race');
    }

    setTrackGaps(trackGaps)
    {
        for (const gap of trackGaps)
        {
            if (gap.CarId in this.#cars)
            {
                let car = this.#cars[gap.CarId];
                car.Laps = gap.Laps;
                car.SplinePosition = gap.SplinePosition;
                car.ExactLaps = car.Laps + car.SplinePosition;
                car.GapsAhead_ms = gap.GapsAhead_ms;
                car.GapsBehind_ms = gap.GapsBehind_ms;
            }
        }
    }
}
