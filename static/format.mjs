export function bestLapGapToCar(current, ahead)
{
    let currentLap_ms = current.BestSessionLap.LapTime * 1e-6;

    if (currentLap_ms < 0)
    {
        return 'NO TIME';
    }

    if ((ahead === undefined) || (current.CarId == ahead.CarId))
    {
        return lapTime(currentLap_ms);
    }

    let aheadLap_ms = ahead.BestSessionLap.LapTime * 1e-6;
    let timeDifference_ms = currentLap_ms - aheadLap_ms;

    return '+' + timeDelta(timeDifference_ms, 3, false);
}

export function laps(nrLaps)
{
    if (nrLaps == 1)
    {
        return `1 lap`;
    }
    else
    {
        return `${nrLaps} laps`;
    }
}

export function lapTime(time_ms)
{
    if (time_ms < 0)
    {
        return '-:--.---';
    }

    var ms = time_ms % 1000;
    var time_s = (time_ms - ms) / 1000;
    var s = time_s % 60;
    var m = (time_s - s) / 60;

    return m.toString() + ':' + s.toString().padStart(2, '0') + '.' + ms.toString().padStart(3, '0');
}

export function gapToCar(current, ahead)
{
    if ((ahead === undefined) || (ahead.CarId == current.CarId))
    {
        return laps(current.Laps);
    }

    let lapDifference = ahead.ExactLaps - current.ExactLaps;
    if (lapDifference > 1)
    {
        return '+' + laps(Math.floor(lapDifference));
    }

    return '+' + timeDelta(current.GapsAhead_ms[ahead.CarId], 1, false);
}

export function time(time_ms)
{
    var ms = time_ms % 1000;
    var time_s = (time_ms - ms) / 1000;
    var s = time_s % 60;
    var time_m = (time_s - s) / 60;
    var m = time_m % 60;
    var h = (time_m - m) / 60;

    return h.toString() + ':' + m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0');
}

export function timeDelta(time_ms, nrDigits = 2, addSign = true)
{
    if (typeof time_ms === 'undefined' || time_ms != time_ms)
    {
        return '';
    }

    var sign = addSign ? ((time_ms < 0) ? '-' : '+') : '';
    var absTime_ms = Math.abs(time_ms);
    var ms = absTime_ms % 1000;
    var s = (absTime_ms - ms) / 1000;

    return sign + s.toString() + '.' + ms.toString().padStart(3, '0').substring(0, nrDigits);
}
