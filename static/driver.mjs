import { DriverBox } from './driverbox.mjs';
import { Standings } from './standings.mjs';
import { State } from './state.mjs';
import { newHandler } from './websocket.mjs';

let state = new State();
let driverbox = new DriverBox(state);
let standings = new Standings(state);

function updateCar(car)
{
    state.setCar(car);
    driverbox.setCar(car);
}

function updateFocusedCar(car)
{
    driverbox.setFocusedCar(car);
    standings.setFocusedCarId(car.CarId);
}

function updateMiniCar(car)
{
    driverbox.setMiniCar(car);
}

function updateSessionType(sessionType)
{
    state.setSessionType(sessionType);
    standings.setSessionType(sessionType);
}

function updateTimeRemainingMS(timeRemaining_ms)
{
    standings.setTimeRemainingMS(timeRemaining_ms);
}

function updateTrackGaps(trackGaps)
{
    state.setTrackGaps(trackGaps);
    driverbox.updateNearbyCars();
    standings.update();
}

function updateBestLapTime(time_ms)
{
    driverbox.setBestLapTime(time_ms);
}

function updateLastLapTime(time_ms)
{
    driverbox.setLastLapTime(time_ms);
}

newHandler('car', updateCar);
newHandler('focusedCar', updateFocusedCar);
newHandler('minicar', updateMiniCar);
newHandler('sessionType', updateSessionType);
newHandler('timeRemaining_ms', updateTimeRemainingMS);
newHandler('trackGaps', updateTrackGaps);

newHandler('bestlaptime_ms', updateBestLapTime);
newHandler('lastlaptime_ms', updateLastLapTime);


const wigglerTemplates = {
    'none': document.querySelector('#wiggler_none'),
    'both': document.querySelector('#wiggler_both'),
    'both_reverse_left': document.querySelector('#wiggler_both_reverse_left'),
    'left': document.querySelector('#wiggler_left'),
    'right': document.querySelector('#wiggler_right'),
}

document.querySelectorAll('[data-wiggler]').forEach(el => {
    const templateId = (el.dataset.wiggler in wigglerTemplates) ? el.dataset.wiggler : 'both';
    const template = wigglerTemplates[templateId];
    el.insertBefore(template.content.cloneNode(true), el.firstChild);
});
