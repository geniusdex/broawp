class PitBarEntry {
    constructor(parent_el, car) {
        this.car = car;
        this.timerRunning = false;
        parent_el.appendChild(this._createOuterElement());
    }

    _createDivWithClasses(classNames) {
        let el = document.createElement('div');
        classNames.forEach(className => el.classList.add(className));
        return el;
    }

    _createOuterElement() {
        let el = this._createDivWithClasses(['car_in_pit_outer']);
        el.appendChild(this._createInnerElement());
        this.el = el;
        return el;
    }

    _createInnerElement() {
        let el = this._createDivWithClasses(['car_in_pit', 'hidden']);
        window.requestAnimationFrame(() => window.requestAnimationFrame(() => el.classList.remove('hidden')));
        el.appendChild(this._createCarNumberElement());
        el.appendChild(this._createDriverNameElement());
        el.appendChild(this._createPositionElement());
        el.appendChild(this._createPitTimeElement());
        this.elInner = el;
        return el;
    }

    _createCarNumberElement() {
        let el = this._createDivWithClasses(['car_number']);
        el.appendChild(document.createTextNode(this.car.RaceNumber));
        return el;
    }

    _createDriverNameElement() {
        let el = this._createDivWithClasses(['driver_name']);
        let driver = this.car.Drivers[this.car.CurrentDriverIndex];
        this.driverNameNode = document.createTextNode(`${driver.FirstName[0]}. ${driver.LastName}`);
        el.appendChild(this.driverNameNode);
        return el;
    }

    _createPositionElement() {
        let el = this._createDivWithClasses(['position']);
        
        el.appendChild(document.createTextNode('P'));
        let span = document.createElement('span');
        span.appendChild(document.createTextNode(this.car.Position));
        el.appendChild(span);

        // el.appendChild(document.createTextNode(' > '));
        // let span2 = document.createElement('span');
        this.currentPositionNode = document.createTextNode(this.car.Position);
        // span2.appendChild(this.currentPositionNode);
        // el.appendChild(span2);

        return el;
    }

    _createCurrentPositionElement() {
        let el = this._createDivWithClasses(['current_position']);
        return el;
    }

    _createPitTimeElement() {
        let el = this._createDivWithClasses(['pit_time']);
        this.pitTimeNode = document.createTextNode('0.0');
        el.appendChild(this.pitTimeNode);
        return el;
    }

    startTimer() {
        if (!('startTime' in this)) {
            this.startTime = performance.now();
        }
        this.timerRunning = true;
        let self = this;
        window.requestAnimationFrame((timestamp) => self._updateTime(timestamp));
    }

    stopTimer() {
        this.timerRunning = false;
    }

    remove() {
        let self = this;
        self.elInner.classList.add('hidden');
        setTimeout(() => self.el.classList.add('hidden'), 500);
        setTimeout(() => this.el.remove(), 1000);
    }

    _updateTime(timestamp) {
        if (this.timerRunning) {
            let timePassed_ms = Math.round(timestamp - this.startTime);
            if (timePassed_ms > 0) {
                this.pitTimeNode.nodeValue = this._formatTime_ms(timePassed_ms);
            }

            // Remove cars after 3 minutes
            if (timePassed_ms > 180000) {
                this.remove();
            } else {
                let self = this;
                window.requestAnimationFrame((timestamp) => self._updateTime(timestamp));
            }
        }
    }

    _formatTime_ms(time_ms) {
        var ms = time_ms % 1000;
        var ds = Math.floor(ms / 100);
        var time_s = (time_ms - ms) / 1000;
        var s = time_s % 60;
        var m = (time_s - s) / 60;

        if (m > 0) {
            return m.toString() + ':' + s.toString().padStart(2, '0') + '.' + ds.toString();
        } else {
            return s.toString() + '.' + ds.toString();
        }
    }

    updateCar(car) {
        console.log(car);
        this.car = car;

        this.currentPositionNode.nodeValue = car.Position;

        let driver = this.car.Drivers[this.car.CurrentDriverIndex];
        this.driverNameNode.nodeValue = `${driver.FirstName[0]}. ${driver.LastName}`;
    }
}

export class PitBar {
    constructor(el) {
        this._el = el;
        this._entries = new Map();
    }

    handlePitEvent(pitEvent) {
        // console.log(pitEvent);
        let car = pitEvent.Car;
        let oldLocation = pitEvent.OldLocation;
        let newLocation = pitEvent.NewLocation;

        if (oldLocation == 'track' && newLocation == 'pit_entry') {
            this._entries.get(car.CarId)?.remove();

            let entry = new PitBarEntry(this._el, pitEvent.Car);
            this._entries.set(car.CarId, entry);
        } else if (oldLocation == 'pit_lane' && newLocation == 'pit_box') {
            this._entries.get(car.CarId)?.startTimer();
        } else if (oldLocation == 'pit_box' && newLocation == 'pit_lane') {
            this._entries.get(car.CarId)?.stopTimer();
        } else if (newLocation == 'track') {
            this._entries.get(car.CarId)?.remove();
            this._entries.delete(car.CarId);
        }
    }

    handleCar(car) {
        // console.log(car);
        this._entries.get(car.CarId)?.updateCar(car);
    }
}
