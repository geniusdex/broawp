import * as format from './format.mjs';

class Entry
{
    #el;
    #els = {};

    constructor(template)
    {
        this.#el = template.content.cloneNode(true).firstElementChild;
        this.#els.position = this.#el.querySelector('[name="position"]');
        this.#els.racenumber = this.#el.querySelector('[name="racenumber"]');
        this.#els.carlogo = this.#el.querySelector('[name="carlogo"]');
        this.#els.name = this.#el.querySelector('[name="name"]');
        this.#els.offset = this.#el.querySelector('[name="offset"]');
    }

    appendTo(parent)
    {
        parent.appendChild(this.#el);
    }

    remove()
    {
        this.#el.remove();
    }

    setPosition(position)
    {
        this.#els.position.innerText = position;
    }

    setCar(car)
    {
        this.#els.racenumber.innerText = car.RaceNumber;
        this.#els.carlogo.src = `static/carlogo/26/${car.CarModel.ManufacturerLabel}.png`;
        this.#els.name.innerText = car.Drivers[car.CurrentDriverIndex].LastName;
    }

    setOffset(offset)
    {
        this.#els.offset.innerText = offset;
    }

    setIsCurrentDriver(isCurrentDriver)
    {
        if (isCurrentDriver)
        {
            this.#el.classList.add('current_driver');
        }
        else
        {
            this.#el.classList.remove('current_driver');
        }
    }
}

class ScrollBehavior
{
    // Constants
    #scrollSpeed = 1; // entries / second
    #waitTimeAtTop = 5000; // ms
    #waitTimeAtBottom = 5000; // ms
    #waitTimeAtFocused = 15000; // ms

    // Variables
    #container;
    #el;
    #focusedEntryIndex = 0;
    #isFocused = true;
    #startTimestamp = null;

    // State     | #isFocused | #startTimestamp
    // ----------+------------+-----------------
    // Focused   | true       | null
    // Top       | false      | null
    // Scrolling | false      | not null
    // Bottom    | false      | null

    constructor(el)
    {
        this.#el = el;
        this.#container = el.parentElement;

        this.#el.classList.add('static');
        this.#toFocused();
    }

    setFocusedEntryIndex(index)
    {
        this.#focusedEntryIndex = index;

        if (this.#isFocused)
        {
            this.#focusOn(index);
        }
    }

    #toTop()
    {
        this.#isFocused = false;
        this.#scrollTo(0);

        let self = this;
        setTimeout(() => self.#toScrolling(), this.#waitTimeAtTop);
    }

    #toScrolling()
    {
        this.#el.classList.remove('static');
        this.#el.classList.add('scrolling');

        let scrollTime = 0;
        const nrEntries = this.#el.childElementCount;
        if (nrEntries > 0)
        {
            const entryHeight = this.#el.firstElementChild.getBoundingClientRect().height;
            const containerHeight = this.#container.clientHeight;
            const allEntriesHeight = this.#el.getBoundingClientRect().height;
            const scrollDistance = allEntriesHeight - containerHeight;
            scrollTime = scrollDistance / (entryHeight * this.#scrollSpeed);

            this.#el.style.setProperty('--scroll-time', `${scrollTime}s`);
            this.#scrollTo(scrollDistance);
        }

        let self = this;
        setTimeout(() => self.#toBottom(), scrollTime * 1000);
    }

    #toBottom()
    {
        this.#startTimestamp = null;

        this.#el.classList.remove('scrolling');
        this.#el.classList.add('static');

        let self = this;
        setTimeout(() => self.#toFocused(), this.#waitTimeAtBottom);
    }

    #toFocused()
    {
        this.#isFocused = true;
        this.#focusOn(this.#focusedEntryIndex);

        let self = this;
        setTimeout(() => self.#toTop(), this.#waitTimeAtFocused);
    }

    #focusOn(index)
    {
        const nrEntries = this.#el.childElementCount;
        if (index < nrEntries)
        {
            const entryHeight = this.#el.firstElementChild.getBoundingClientRect().height;
            const containerHeight = this.#container.clientHeight;
            const allEntriesHeight = this.#el.getBoundingClientRect().height;
            const maxScrollTop = allEntriesHeight - containerHeight;
            const nrEntriesVisible = containerHeight / entryHeight;
            const scrollToEntry = Math.max(0, Math.round(index - (nrEntriesVisible / 2)));
            const scrollTop = Math.floor(Math.min(maxScrollTop, scrollToEntry * entryHeight));
            this.#scrollTo(scrollTop);
        }
    }

    #scrollTo(top_px)
    {
        this.#el.style.transform = `translateY(-${top_px}px)`;
    }
}

export class Standings
{
    #state;
    #el;
    #els = {};
    #targetIsLaps = true;
    #entries = [];
    #focusedCarId = null;
    #timingIsInterval = true;
    #scrollBehavior;

    constructor(state)
    {
        this.#state = state;

        this.#el = document.querySelector('#standings');
        this.#els.entries = this.#el.querySelector('.entries');
        this.#els.entryTemplate = this.#el.querySelector('#standings_entry');
        this.#els.sessionType = this.#el.querySelector('[name="session_type"]');
        this.#els.timeRemaining = this.#el.querySelector('[name="remaining_time"]');
        this.#els.timingType = this.#el.querySelector('[name="timing_type"]');

        this.#scrollBehavior = new ScrollBehavior(this.#els.entries);

        let self = this;
        setInterval(() => self.#toggleTimingType(), 15000);
    }

    setFocusedCarId(carId)
    {
        this.#focusedCarId = carId;
    }

    setSessionType(sessionType)
    {
        this.#els.sessionType.innerText = sessionType.substring(0, 1);
    }

    setTimeRemainingMS(timeRemaining_ms)
    {
        this.#els.timeRemaining.innerText = format.time(timeRemaining_ms);
    }

    update()
    {
        this.#els.timingType.innerText = this.#timingIsInterval ? 'Interval' : 'To Leader';

        const sessionTargetIsLaps = this.#state.sessionTargetIsLaps();
        const cars = sessionTargetIsLaps
                   ? this.#state.carsSortedOnMostLaps()
                   : this.#state.carsSortedOnBestLap();

        this.#ensureNrOfEntries(cars.length);

        for (let i = 0; i < cars.length; i++)
        {
            const car = cars[i];
            let entry = this.#entries[i];

            if (this.#focusedCarId == car.CarId)
            {
                entry.setIsCurrentDriver(true);
                this.#scrollBehavior.setFocusedEntryIndex(i);
            }
            else
            {
                entry.setIsCurrentDriver(false);
            }

            entry.setPosition(i + 1);
            entry.setCar(car);

            const aheadIndex = this.#timingIsInterval ? (i-1) : 0;
            if (sessionTargetIsLaps)
            {
                entry.setOffset(format.gapToCar(car, cars[aheadIndex]));
            }
            else
            {
                entry.setOffset(format.bestLapGapToCar(car, cars[aheadIndex]));
            }
        }
    }

    #ensureNrOfEntries(nr)
    {
        while (this.#entries.length > nr)
        {
            let entry = this.#entries.pop();
            entry.remove();
        }

        while (this.#entries.length < nr)
        {
            let entry = new Entry(this.#els.entryTemplate);
            entry.appendTo(this.#els.entries);
            this.#entries.push(entry);
        }
    }

    #toggleTimingType()
    {
        this.#timingIsInterval = !this.#timingIsInterval;
    }
}
