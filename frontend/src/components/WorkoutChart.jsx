import { useState, useRef, useCallback } from 'react';
import { getZoneColor, getRampZoneColor, toPercent, toWatts } from '../utils/zones';

const CHART_PADDING_LEFT = 50;
const CHART_PADDING_RIGHT = 20;
const CHART_PADDING_TOP = 20;
const CHART_PADDING_BOTTOM = 40;
const MAX_POWER_FRACTION = 1.5;
const MIN_POWER_FRACTION = 0;

/**
 * SVG-based workout chart showing power over time, with interactive block editing.
 */
export default function WorkoutChart({
  intervals,
  selectedId,
  onSelectInterval,
  onUpdateInterval,
  ftp = 200,
  width = 800,
  height = 400,
}) {
  const svgRef = useRef(null);
  const [dragState, setDragState] = useState(null);

  const chartWidth = width - CHART_PADDING_LEFT - CHART_PADDING_RIGHT;
  const chartHeight = height - CHART_PADDING_TOP - CHART_PADDING_BOTTOM;

  const totalDuration = intervals.reduce((sum, i) => sum + i.duration_sec, 0);
  const timeScale = totalDuration > 0 ? chartWidth / totalDuration : 1;

  function powerToY(powerFraction) {
    const clamped = Math.max(MIN_POWER_FRACTION, Math.min(MAX_POWER_FRACTION, powerFraction));
    return CHART_PADDING_TOP + chartHeight * (1 - clamped / MAX_POWER_FRACTION);
  }

  function yToPower(y) {
    const fraction = (1 - (y - CHART_PADDING_TOP) / chartHeight) * MAX_POWER_FRACTION;
    return Math.max(0.1, Math.min(2.0, Math.round(fraction * 100) / 100));
  }

  function xToDuration(deltaX) {
    if (timeScale === 0) return 0;
    return Math.round(deltaX / timeScale);
  }

  const getSvgPoint = useCallback(
    (clientX, clientY) => {
      const svg = svgRef.current;
      if (!svg) return { x: 0, y: 0 };
      const rect = svg.getBoundingClientRect();
      return {
        x: clientX - rect.left,
        y: clientY - rect.top,
      };
    },
    []
  );

  function handleMouseDown(e, intervalData, dragType) {
    e.stopPropagation();
    e.preventDefault();
    const point = getSvgPoint(e.clientX, e.clientY);
    setDragState({
      intervalId: intervalData.id,
      dragType,
      startX: point.x,
      startY: point.y,
      originalInterval: { ...intervalData },
    });
  }

  function handleMouseMove(e) {
    if (!dragState) return;
    const point = getSvgPoint(e.clientX, e.clientY);

    if (dragState.dragType === 'power') {
      const newPower = yToPower(point.y);
      const orig = dragState.originalInterval;
      if (orig.type === 'steady' || orig.type === 'rest' || orig.type === 'free_ride') {
        onUpdateInterval(dragState.intervalId, {
          power_start: newPower,
          power_end: newPower,
        });
      } else {
        // For ramps, adjust whichever end is closer
        const diff = newPower - (orig.power_start + orig.power_end) / 2;
        onUpdateInterval(dragState.intervalId, {
          power_start: Math.max(0.1, Math.round((orig.power_start + diff) * 100) / 100),
          power_end: Math.max(0.1, Math.round((orig.power_end + diff) * 100) / 100),
        });
      }
    } else if (dragState.dragType === 'duration-right') {
      const deltaX = point.x - dragState.startX;
      const deltaSec = xToDuration(deltaX);
      const newDuration = Math.max(5, dragState.originalInterval.duration_sec + deltaSec);
      onUpdateInterval(dragState.intervalId, { duration_sec: newDuration });
    } else if (dragState.dragType === 'duration-left') {
      const deltaX = dragState.startX - point.x;
      const deltaSec = xToDuration(deltaX);
      const newDuration = Math.max(5, dragState.originalInterval.duration_sec + deltaSec);
      onUpdateInterval(dragState.intervalId, { duration_sec: newDuration });
    }
  }

  function handleMouseUp() {
    setDragState(null);
  }

  // Build the blocks
  let xOffset = CHART_PADDING_LEFT;
  const blocks = intervals.map((interval) => {
    const blockWidth = interval.duration_sec * timeScale;
    const isSelected = interval.id === selectedId;
    const x = xOffset;
    xOffset += blockWidth;

    const isRamp =
      interval.type === 'ramp_up' ||
      interval.type === 'ramp_down' ||
      (interval.type === 'ramp' && interval.power_start !== interval.power_end);

    const color =
      isRamp
        ? getRampZoneColor(interval.power_start, interval.power_end)
        : getZoneColor(interval.power_start);

    const baseY = CHART_PADDING_TOP + chartHeight;

    if (isRamp || interval.power_start !== interval.power_end) {
      // Render as polygon (sloped shape)
      const y1 = powerToY(interval.power_start);
      const y2 = powerToY(interval.power_end);
      const points = [
        `${x},${baseY}`,
        `${x},${y1}`,
        `${x + blockWidth},${y2}`,
        `${x + blockWidth},${baseY}`,
      ].join(' ');

      return (
        <g key={interval.id} data-testid={`interval-block-${interval.id}`}>
          <polygon
            points={points}
            fill={color}
            opacity={isSelected ? 1 : 0.8}
            stroke={isSelected ? '#1a1a2e' : '#fff'}
            strokeWidth={isSelected ? 2 : 1}
            className="interval-block"
            onClick={(e) => {
              e.stopPropagation();
              onSelectInterval(interval.id);
            }}
          />
          {/* Top edge drag handle for power */}
          <line
            x1={x}
            y1={y1}
            x2={x + blockWidth}
            y2={y2}
            stroke="transparent"
            strokeWidth={10}
            style={{ cursor: 'ns-resize' }}
            onMouseDown={(e) => handleMouseDown(e, interval, 'power')}
          />
          {/* Right edge drag handle for duration */}
          <line
            x1={x + blockWidth}
            y1={y2}
            x2={x + blockWidth}
            y2={baseY}
            stroke="transparent"
            strokeWidth={8}
            style={{ cursor: 'ew-resize' }}
            onMouseDown={(e) => handleMouseDown(e, interval, 'duration-right')}
          />
          {/* Left edge drag handle for duration */}
          <line
            x1={x}
            y1={y1}
            x2={x}
            y2={baseY}
            stroke="transparent"
            strokeWidth={8}
            style={{ cursor: 'ew-resize' }}
            onMouseDown={(e) => handleMouseDown(e, interval, 'duration-left')}
          />
          {/* Selection highlight */}
          {isSelected && (
            <polygon
              points={points}
              fill="none"
              stroke="#1a1a2e"
              strokeWidth={2}
              strokeDasharray="4,2"
              pointerEvents="none"
            />
          )}
        </g>
      );
    }

    // Steady block (rectangle)
    const y = powerToY(interval.power_start);
    const blockHeight = baseY - y;

    return (
      <g key={interval.id} data-testid={`interval-block-${interval.id}`}>
        <rect
          x={x}
          y={y}
          width={Math.max(blockWidth, 1)}
          height={Math.max(blockHeight, 1)}
          fill={color}
          opacity={isSelected ? 1 : 0.8}
          stroke={isSelected ? '#1a1a2e' : '#fff'}
          strokeWidth={isSelected ? 2 : 1}
          className="interval-block"
          onClick={(e) => {
            e.stopPropagation();
            onSelectInterval(interval.id);
          }}
        />
        {/* Top edge drag handle for power */}
        <rect
          x={x}
          y={y - 5}
          width={Math.max(blockWidth, 1)}
          height={10}
          fill="transparent"
          style={{ cursor: 'ns-resize' }}
          onMouseDown={(e) => handleMouseDown(e, interval, 'power')}
        />
        {/* Right edge drag handle for duration */}
        <rect
          x={x + blockWidth - 4}
          y={y}
          width={8}
          height={blockHeight}
          fill="transparent"
          style={{ cursor: 'ew-resize' }}
          onMouseDown={(e) => handleMouseDown(e, interval, 'duration-right')}
        />
        {/* Left edge drag handle for duration */}
        <rect
          x={x - 4}
          y={y}
          width={8}
          height={blockHeight}
          fill="transparent"
          style={{ cursor: 'ew-resize' }}
          onMouseDown={(e) => handleMouseDown(e, interval, 'duration-left')}
        />
        {/* Selection highlight */}
        {isSelected && (
          <rect
            x={x}
            y={y}
            width={Math.max(blockWidth, 1)}
            height={Math.max(blockHeight, 1)}
            fill="none"
            stroke="#1a1a2e"
            strokeWidth={2}
            strokeDasharray="4,2"
            pointerEvents="none"
          />
        )}
      </g>
    );
  });

  // Y-axis labels (% FTP)
  const yAxisLabels = [];
  const yAxisLines = [];
  for (let pct = 0; pct <= 150; pct += 25) {
    const fraction = pct / 100;
    const y = powerToY(fraction);
    yAxisLabels.push(
      <text
        key={`y-label-${pct}`}
        x={CHART_PADDING_LEFT - 5}
        y={y + 4}
        textAnchor="end"
        fontSize={11}
        fill="#666"
      >
        {pct}%
      </text>
    );
    // Secondary watts label
    yAxisLabels.push(
      <text
        key={`y-watts-${pct}`}
        x={CHART_PADDING_LEFT - 5}
        y={y + 15}
        textAnchor="end"
        fontSize={9}
        fill="#999"
      >
        {toWatts(fraction, ftp)}W
      </text>
    );
    yAxisLines.push(
      <line
        key={`y-line-${pct}`}
        x1={CHART_PADDING_LEFT}
        y1={y}
        x2={width - CHART_PADDING_RIGHT}
        y2={y}
        stroke="#e5e7eb"
        strokeWidth={1}
      />
    );
  }

  // X-axis labels (time in minutes)
  const xAxisLabels = [];
  if (totalDuration > 0) {
    const totalMinutes = totalDuration / 60;
    const step = totalMinutes <= 10 ? 1 : totalMinutes <= 30 ? 2 : 5;
    for (let min = 0; min <= totalMinutes; min += step) {
      const x = CHART_PADDING_LEFT + (min * 60 * timeScale);
      xAxisLabels.push(
        <text
          key={`x-label-${min}`}
          x={x}
          y={height - 10}
          textAnchor="middle"
          fontSize={11}
          fill="#666"
        >
          {min}m
        </text>
      );
    }
  }

  return (
    <svg
      ref={svgRef}
      width={width}
      height={height}
      className="workout-chart"
      data-testid="workout-chart"
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onClick={() => onSelectInterval(null)}
    >
      {/* Background */}
      <rect
        x={CHART_PADDING_LEFT}
        y={CHART_PADDING_TOP}
        width={chartWidth}
        height={chartHeight}
        fill="#fafafa"
      />

      {/* Grid lines */}
      {yAxisLines}

      {/* Y-axis labels */}
      {yAxisLabels}

      {/* Interval blocks */}
      {blocks}

      {/* X-axis line */}
      <line
        x1={CHART_PADDING_LEFT}
        y1={CHART_PADDING_TOP + chartHeight}
        x2={width - CHART_PADDING_RIGHT}
        y2={CHART_PADDING_TOP + chartHeight}
        stroke="#333"
        strokeWidth={1}
      />

      {/* Y-axis line */}
      <line
        x1={CHART_PADDING_LEFT}
        y1={CHART_PADDING_TOP}
        x2={CHART_PADDING_LEFT}
        y2={CHART_PADDING_TOP + chartHeight}
        stroke="#333"
        strokeWidth={1}
      />

      {/* X-axis labels */}
      {xAxisLabels}
    </svg>
  );
}
