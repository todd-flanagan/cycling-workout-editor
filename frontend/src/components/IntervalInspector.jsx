import { toPercent, toWatts } from '../utils/zones';

function formatDuration(seconds) {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

function parseDuration(str) {
  const parts = str.split(':');
  if (parts.length === 2) {
    const m = parseInt(parts[0], 10) || 0;
    const s = parseInt(parts[1], 10) || 0;
    return m * 60 + s;
  }
  return parseInt(str, 10) || 0;
}

export default function IntervalInspector({
  interval,
  onUpdate,
  onRemove,
  ftp = 200,
}) {
  if (!interval) {
    return (
      <div className="interval-inspector" data-testid="interval-inspector">
        <h3>Interval Inspector</h3>
        <p className="inspector-empty">Select an interval to edit its properties.</p>
      </div>
    );
  }

  function handleTypeChange(e) {
    const newType = e.target.value;
    const changes = { type: newType };
    // Adjust power_end for type changes
    if (newType === 'steady' || newType === 'rest' || newType === 'free_ride') {
      changes.power_end = interval.power_start;
    }
    if (newType === 'rest') {
      changes.power_start = 0.40;
      changes.power_end = 0.40;
    }
    onUpdate(interval.id, changes);
  }

  function handleDurationChange(e) {
    const seconds = parseDuration(e.target.value);
    if (seconds >= 5) {
      onUpdate(interval.id, { duration_sec: seconds });
    }
  }

  function handlePowerStartChange(e) {
    const pct = parseInt(e.target.value, 10);
    if (!isNaN(pct) && pct >= 10 && pct <= 200) {
      const fraction = pct / 100;
      const changes = { power_start: fraction };
      // For steady/rest/free_ride, keep power_end in sync
      if (
        interval.type === 'steady' ||
        interval.type === 'rest' ||
        interval.type === 'free_ride'
      ) {
        changes.power_end = fraction;
      }
      onUpdate(interval.id, changes);
    }
  }

  function handlePowerEndChange(e) {
    const pct = parseInt(e.target.value, 10);
    if (!isNaN(pct) && pct >= 10 && pct <= 200) {
      onUpdate(interval.id, { power_end: pct / 100 });
    }
  }

  function handleCadenceChange(e) {
    const val = e.target.value;
    if (val === '') {
      onUpdate(interval.id, { cadence: null });
    } else {
      const cadence = parseInt(val, 10);
      if (!isNaN(cadence) && cadence >= 0 && cadence <= 200) {
        onUpdate(interval.id, { cadence });
      }
    }
  }

  const isRamp =
    interval.type === 'ramp_up' || interval.type === 'ramp_down';

  return (
    <div className="interval-inspector" data-testid="interval-inspector">
      <h3>Interval Inspector</h3>

      <div className="inspector-field">
        <label htmlFor="interval-type">Type</label>
        <select
          id="interval-type"
          value={interval.type}
          onChange={handleTypeChange}
        >
          <option value="steady">Steady State</option>
          <option value="ramp_up">Ramp Up</option>
          <option value="ramp_down">Ramp Down</option>
          <option value="free_ride">Free Ride</option>
          <option value="rest">Rest</option>
        </select>
      </div>

      <div className="inspector-field">
        <label htmlFor="interval-duration">Duration (mm:ss)</label>
        <input
          id="interval-duration"
          type="text"
          value={formatDuration(interval.duration_sec)}
          onChange={handleDurationChange}
        />
      </div>

      <div className="inspector-field">
        <label htmlFor="interval-power-start">
          {isRamp ? 'Power Start (% FTP)' : 'Power (% FTP)'}
        </label>
        <input
          id="interval-power-start"
          type="number"
          min={10}
          max={200}
          value={Math.round(interval.power_start * 100)}
          onChange={handlePowerStartChange}
        />
        <span className="inspector-watts">
          {toWatts(interval.power_start, ftp)}W
        </span>
      </div>

      {isRamp && (
        <div className="inspector-field">
          <label htmlFor="interval-power-end">Power End (% FTP)</label>
          <input
            id="interval-power-end"
            type="number"
            min={10}
            max={200}
            value={Math.round(interval.power_end * 100)}
            onChange={handlePowerEndChange}
          />
          <span className="inspector-watts">
            {toWatts(interval.power_end, ftp)}W
          </span>
        </div>
      )}

      <div className="inspector-field">
        <label htmlFor="interval-cadence">Cadence (rpm, optional)</label>
        <input
          id="interval-cadence"
          type="number"
          min={0}
          max={200}
          value={interval.cadence ?? ''}
          onChange={handleCadenceChange}
          placeholder="--"
        />
      </div>

      <div className="inspector-summary">
        <p>
          {toPercent(interval.power_start)} FTP = {toWatts(interval.power_start, ftp)}W
          {isRamp && (
            <>
              {' → '}
              {toPercent(interval.power_end)} FTP = {toWatts(interval.power_end, ftp)}W
            </>
          )}
        </p>
        <p>Duration: {formatDuration(interval.duration_sec)}</p>
      </div>

      <button
        className="btn btn-danger inspector-remove-btn"
        onClick={() => onRemove(interval.id)}
      >
        Remove Interval
      </button>
    </div>
  );
}
