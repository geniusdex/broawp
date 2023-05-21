import * as format from './format.mjs';

class CurrentDriver
{
    #el;
    #els = {};

    #currentLapTime_ms = 0;
    #currentLapTimeUpdate = performance.now();

    constructor(parentEl)
    {
        this.#el = parentEl.querySelector('#current_driver');
        this.#els.racenumber = this.#el.querySelector('[name="racenumber"]');
        this.#els.carlogo = this.#el.querySelector('[name="carlogo"]');
        this.#els.name = this.#el.querySelector('[name="name"]');
        this.#els.bestLap = this.#el.querySelector('[name="best_lap"]');
        this.#els.lastLap = this.#el.querySelector('[name="last_lap"]');
        this.#els.currentLap = this.#el.querySelector('[name="current_lap"]');
        this.#els.currentLapDelta = this.#el.querySelector('[name="current_lap_delta"]');
        this.#els.gear = this.#el.querySelector('[name="gear"]');
        this.#els.speed = this.#el.querySelector('[name="speed"]');

        let self = this;
        window.requestAnimationFrame(timestamp => self.#redrawCurrentLapTime(timestamp));
    }

    setCar(car)
    {
        const driver = car.Drivers[car.CurrentDriverIndex];

        this.#els.racenumber.innerText = car.RaceNumber;
        this.#els.carlogo.src = `static/carlogo/26/${car.CarModel.ManufacturerLabel}.png`;
        this.#els.name.innerText = `${driver.FirstName.substring(0, 1)}. ${driver.LastName}`;
    }

    setMiniCar(car)
    {
        this.#setCurrentLapTime(car.CurrentLap_ms);
        this.#setCurrentLapDelta(car.Delta_ms);
        this.#els.gear.innerText = this.#getGearString(car.Gear);
        this.#els.speed.innerText = car.SpeedKmh;
    }

    setBestLapTime(time_ms)
    {
        this.#els.bestLap.innerText = format.lapTime(time_ms);
    }

    setLastLapTime(time_ms)
    {
        this.#els.lastLap.innerText = format.lapTime(time_ms);
    }

    #setCurrentLapTime(time_ms)
    {
        this.#currentLapTime_ms = time_ms;
        this.#currentLapTimeUpdate = performance.now();
    }

    #redrawCurrentLapTime(timestamp)
    {
        const lapTime_ms = this.#currentLapTime_ms + Math.round(timestamp - this.#currentLapTimeUpdate);
        this.#els.currentLap.innerText = format.lapTime(lapTime_ms);

        let self = this;
        window.requestAnimationFrame(timestamp => self.#redrawCurrentLapTime(timestamp));
    }

    #setCurrentLapDelta(time_ms)
    {
        this.#els.currentLapDelta.innerText = format.timeDelta(time_ms);

        if (time_ms < 0)
        {
            this.#els.currentLapDelta.classList.remove('slower');
            this.#els.currentLapDelta.classList.add('faster');
        }
        else
        {
            this.#els.currentLapDelta.classList.remove('faster');
            this.#els.currentLapDelta.classList.add('slower');
        }
    }

    #getGearString(gear)
    {
        if (gear > 0)
            return gear;
        else if (gear == 0)
            return 'N';
        else
            return 'R';
    }
}


class DriverRelative
{
    #el;
    #els = {};

    constructor(template, className)
    {
        this.#el = template.content.cloneNode(true).firstElementChild;
        this.#el.classList.add(className);

        this.#els.racenumber = this.#el.querySelector('[name="racenumber"]');
        this.#els.carlogo = this.#el.querySelector('[name="carlogo"]');
        this.#els.name = this.#el.querySelector('[name="name"]');
        this.#els.lapsOffset = this.#el.querySelector('[name="laps_offset"]');
        this.#els.timeOffset = this.#el.querySelector('[name="time_offset"]');
    }

    _insertBefore(parent, referenceNode)
    {
        parent.insertBefore(this.#el, referenceNode);
    }

    update(car, gap_ms, lapsAhead)
    {
        const driver = car.Drivers[car.CurrentDriverIndex];

        this.#els.racenumber.innerText = car.RaceNumber;
        this.#els.carlogo.src = `static/carlogo/26/${car.CarModel.ManufacturerLabel}.png`;
        this.#els.name.innerText = `${driver.FirstName.charAt(0)}. ${driver.LastName}`;
        this.#els.timeOffset.innerText = car.IsInPit ? 'PIT' : format.timeDelta(gap_ms);

        if (lapsAhead > 0)
        {
            this.#el.classList.remove('lap_behind');
            this.#el.classList.add('lap_ahead');
            this.#els.lapsOffset.innerText = `+${lapsAhead} lap${(lapsAhead == 1) ? '' : 's'}`;
        }
        else if (lapsAhead < 0)
        {
            this.#el.classList.remove('lap_ahead');
            this.#el.classList.add('lap_behind');
            this.#els.lapsOffset.innerText = `${lapsAhead} lap${(lapsAhead == -1) ? '' : 's'}`;
        }
        else
        {
            this.#el.classList.remove('lap_ahead');
            this.#el.classList.remove('lap_behind');
            this.#els.lapsOffset.innerText = '';
        }
    }
}

class DriverAhead extends DriverRelative
{
    constructor(template)
    {
        super(template, 'driver_ahead');
    }

    update(car, relativeToCar)
    {
        const gap_ms = relativeToCar.GapsAhead_ms[car.CarId];

        const carLaps = car.Laps + car.SplinePosition;
        const relativeToLaps = relativeToCar.Laps + relativeToCar.SplinePosition;
        let lapsAhead = carLaps - relativeToLaps;
        if (lapsAhead <= 0)
        {
            lapsAhead -= 1;
        }

        super.update(car, gap_ms, Math.trunc(lapsAhead));
    }

    addTo(parent)
    {
        super._insertBefore(parent, parent.firstElementChild);
    }
}

class DriverBehind extends DriverRelative
{
    constructor(template)
    {
        super(template, 'driver_behind');
    }

    update(car, relativeToCar)
    {
        const gap_ms = relativeToCar.GapsBehind_ms[car.CarId];

        const carLaps = car.Laps + car.SplinePosition;
        const relativeToLaps = relativeToCar.Laps + relativeToCar.SplinePosition;
        let lapsAhead = carLaps - relativeToLaps;
        if (lapsAhead >= 0)
        {
            lapsAhead += 1;
        }

        super.update(car, -gap_ms, Math.trunc(lapsAhead));
    }

    addTo(parent)
    {
        super._insertBefore(parent, null);
    }
}


export class DriverBox
{
    #state;
    #focusedCarId = null;
    #el;
    #els = {};
    #currentDriver;

    // Indices counted from current driver
    #ahead = [];
    #behind = [];

    constructor(state)
    {
        const nrRelativeDrivers = 3;

        this.#state = state;

        this.#el = document.querySelector('#driver_box');
        this.#els.driversAhead = this.#el.querySelector('#drivers_ahead');
        this.#els.driversBehind = this.#el.querySelector('#drivers_behind');
        this.#els.driverRelativeTemplate = this.#el.querySelector('#driver_relative_template');

        this.#currentDriver = new CurrentDriver(this.#el);

        for (let i = 0; i < nrRelativeDrivers; i++)
        {
            let ahead = new DriverAhead(this.#els.driverRelativeTemplate);
            ahead.addTo(this.#els.driversAhead);
            this.#ahead.push(ahead);

            let behind = new DriverBehind(this.#els.driverRelativeTemplate);
            behind.addTo(this.#els.driversBehind);
            this.#behind.push(behind);
        }
    }

    setFocusedCar(car)
    {
        this.#focusedCarId = car.CarId;
        this.#currentDriver.setCar(car);
        this.updateNearbyCars();
    }

    setCar(car)
    {
        if (car.CarId == this.#focusedCarId)
        {
            this.#currentDriver.setCar(car);
        }
    }

    setMiniCar(car)
    {
        if (car.CarId == this.#focusedCarId)
        {
            this.#currentDriver.setMiniCar(car);
        }
    }

    setBestLapTime(time_ms)
    {
        this.#currentDriver.setBestLapTime(time_ms);
    }

    setLastLapTime(time_ms)
    {
        this.#currentDriver.setLastLapTime(time_ms);
    }

    updateNearbyCars()
    {
        const atWrapped = (arr, i) => arr.at(i % arr.length);

        const cars = this.#state.carsSortedOnTrackPosition();

        for (let i = 0; i < cars.length; i++)
        {
            if (cars[i].CarId == this.#focusedCarId)
            {
                const current = cars[i];

                for (let j = 0; j < this.#ahead.length; j++)
                {
                    this.#ahead[j].update(atWrapped(cars, i+j+1), current);
                }

                for (let j = 0; j < this.#behind.length; j++)
                {
                    this.#behind[j].update(atWrapped(cars, i-j-1), current);
                }
            }
        }
    }
}
