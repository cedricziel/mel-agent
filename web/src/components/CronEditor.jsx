import React, { useState, useEffect } from 'react';

export default function CronEditor({ value, onCronChange }) {
  const [mode, setMode] = useState('interval');
  const [intervalValue, setIntervalValue] = useState(1);
  const [intervalUnit, setIntervalUnit] = useState('minutes');
  const [dateTime, setDateTime] = useState('');
  const [customCron, setCustomCron] = useState('');

  useEffect(() => {
    let cron = '';
    if (mode === 'interval') {
      const v = parseInt(intervalValue, 10) || 1;
      switch (intervalUnit) {
        case 'minutes':
          cron = `*/${v} * * * *`;
          break;
        case 'hours':
          cron = `0 */${v} * * *`;
          break;
        case 'days':
          cron = `0 0 */${v} * *`;
          break;
        default:
          cron = `*/${v} * * * *`;
      }
    } else if (mode === 'once' && dateTime) {
      const dt = new Date(dateTime);
      const minute = dt.getMinutes();
      const hour = dt.getHours();
      const day = dt.getDate();
      const month = dt.getMonth() + 1;
      cron = `${minute} ${hour} ${day} ${month} *`;
    } else if (mode === 'custom') {
      cron = customCron;
    }
    if (cron !== value) {
      onCronChange(cron);
    }
  }, [mode, intervalValue, intervalUnit, dateTime, customCron]);

  return (
    <div className="bg-white border rounded p-2">
      <div className="mb-2">
        <label className="flex items-center">
          <input
            type="radio"
            name="cronMode"
            checked={mode === 'interval'}
            onChange={() => setMode('interval')}
          />
          <span className="ml-2 text-sm">Every</span>
        </label>
        {mode === 'interval' && (
          <div className="flex items-center mt-1 space-x-2">
            <input
              type="number"
              min="1"
              value={intervalValue}
              onChange={(e) => setIntervalValue(e.target.value)}
              className="w-16 border rounded px-1 py-0.5"
            />
            <select
              value={intervalUnit}
              onChange={(e) => setIntervalUnit(e.target.value)}
              className="border rounded px-2 py-0.5"
            >
              <option value="minutes">Minutes</option>
              <option value="hours">Hours</option>
              <option value="days">Days</option>
            </select>
          </div>
        )}
      </div>
      <div className="mb-2">
        <label className="flex items-center">
          <input
            type="radio"
            name="cronMode"
            checked={mode === 'once'}
            onChange={() => setMode('once')}
          />
          <span className="ml-2 text-sm">Once at</span>
        </label>
        {mode === 'once' && (
          <input
            type="datetime-local"
            value={dateTime}
            onChange={(e) => setDateTime(e.target.value)}
            className="mt-1 border rounded px-2 py-1 w-full"
          />
        )}
      </div>
      <div>
        <label className="flex items-center">
          <input
            type="radio"
            name="cronMode"
            checked={mode === 'custom'}
            onChange={() => setMode('custom')}
          />
          <span className="ml-2 text-sm">Custom</span>
        </label>
        {mode === 'custom' && (
          <input
            type="text"
            value={customCron}
            onChange={(e) => setCustomCron(e.target.value)}
            placeholder="* * * * *"
            className="mt-1 border rounded px-2 py-1 w-full"
          />
        )}
      </div>
      <div className="mt-2 text-xs italic text-gray-600">Cron: {value}</div>
    </div>
  );
}
