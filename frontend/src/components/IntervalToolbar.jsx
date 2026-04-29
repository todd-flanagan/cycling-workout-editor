const INTERVAL_TYPES = [
  {
    type: 'steady',
    label: 'Steady State',
    icon: '▆',
    defaults: { power_start: 0.75, power_end: 0.75, duration_sec: 300 },
    color: '#3B82F6',
  },
  {
    type: 'ramp_up',
    label: 'Ramp Up',
    icon: '╱',
    defaults: { power_start: 0.50, power_end: 1.0, duration_sec: 300 },
    color: '#22C55E',
  },
  {
    type: 'ramp_down',
    label: 'Ramp Down',
    icon: '╲',
    defaults: { power_start: 1.0, power_end: 0.50, duration_sec: 300 },
    color: '#EAB308',
  },
  {
    type: 'free_ride',
    label: 'Free Ride',
    icon: '∿',
    defaults: { power_start: 0.60, power_end: 0.60, duration_sec: 600 },
    color: '#A0C4E8',
  },
  {
    type: 'rest',
    label: 'Rest',
    icon: '⏸',
    defaults: { power_start: 0.40, power_end: 0.40, duration_sec: 300 },
    color: '#9CA3AF',
  },
];

export { INTERVAL_TYPES };

export default function IntervalToolbar({ onAddInterval }) {
  function handleDragStart(e, intervalType) {
    e.dataTransfer.setData('application/json', JSON.stringify(intervalType));
    e.dataTransfer.effectAllowed = 'copy';
  }

  function handleClick(intervalType) {
    onAddInterval({
      type: intervalType.type,
      ...intervalType.defaults,
    });
  }

  return (
    <div className="interval-toolbar" data-testid="interval-toolbar">
      <h3 className="toolbar-title">Intervals</h3>
      <div className="toolbar-items">
        {INTERVAL_TYPES.map((item) => (
          <div
            key={item.type}
            className="toolbar-item"
            data-testid={`toolbar-item-${item.type}`}
            draggable
            onDragStart={(e) => handleDragStart(e, item)}
            onClick={() => handleClick(item)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                handleClick(item);
              }
            }}
          >
            <span
              className="toolbar-item-icon"
              style={{ color: item.color }}
            >
              {item.icon}
            </span>
            <span className="toolbar-item-label">{item.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
