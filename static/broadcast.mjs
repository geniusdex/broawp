import { newHandler } from "./websocket.mjs";
import { PitBar } from "./pitbar.mjs";

let pitbar = new PitBar(document.getElementById('cars_in_pit'));

newHandler('pitEvent', (pitEvent) => pitbar.handlePitEvent(pitEvent));
newHandler('car', (car) => pitbar.handleCar(car));

function triggerPitEvent(carId, from, to) {
    pitbar.handlePitEvent({
        "Car": {
            "CarId": carId,
            "IsConnected": true,
            "TeamName": "GDX3",
            "RaceNumber": carId,
            "CurrentDriverIndex": 0,
            "Drivers": [
                {
                    "FirstName": "Pieter",
                    "LastName": "Bootsma",
                    "ShortName": "BOO"
                },
                {
                    "FirstName": "First",
                    "LastName": "Teammate",
                    "ShortName": "FTE"
                },
                {
                    "FirstName": "Second",
                    "LastName": "Teammate",
                    "ShortName": "STE"
                }
            ],
            "IsInPit": (to != "unknown") && (to != "track"),
            "Location": 2,
            "Gear": -1,
            "SpeedKmh": 0,
            "Position": 13,
            "CupPosition": 13,
            "TrackPosition": 0,
            "SplinePosition": 0.9827272,
            "Laps": 0,
            "Delta": 0,
            "BestSessionLap": {
                "LapTime": -1000000,
                "DriverIndex": 0,
                "SplitTimes": [
                    -1000000,
                    -1000000,
                    -1000000
                ],
                "IsValid": true
            },
            "LastLap": {
                "LapTime": -1000000,
                "DriverIndex": 0,
                "SplitTimes": [
                    -1000000,
                    -1000000,
                    -1000000
                ],
                "IsValid": true
            },
            "CurrentLap": {
                "LapTime": 92921000000,
                "DriverIndex": 0,
                "SplitTimes": [
                    -1000000,
                    -1000000,
                    -1000000
                ],
                "IsValid": true
            }
        },
        "OldLocation": from,
        "NewLocation": to
    });
}

setTimeout(() => triggerPitEvent(1, 'track', 'pit_entry'), 500);
setTimeout(() => triggerPitEvent(1, 'pit_lane', 'pit_box'), 4000);

setTimeout(() => triggerPitEvent(418, 'track', 'pit_entry'), 1000);
setTimeout(() => triggerPitEvent(418, 'pit_lane', 'pit_box'), 6000);
setTimeout(() => triggerPitEvent(418, 'pit_box', 'pit_lane'), 11000);
setTimeout(() => triggerPitEvent(418, 'pit_exit', 'track'), 16000);


setTimeout(() => triggerPitEvent(42, 'track', 'pit_entry'), 2500);
setTimeout(() => triggerPitEvent(42, 'pit_lane', 'pit_box'), 7500);
