import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import WorkoutChart from './WorkoutChart';

const steadyInterval = {
  id: 'interval-1',
  type: 'steady',
  duration_sec: 300,
  power_start: 0.75,
  power_end: 0.75,
  cadence: null,
};

const rampInterval = {
  id: 'interval-2',
  type: 'ramp_up',
  duration_sec: 300,
  power_start: 0.50,
  power_end: 1.0,
  cadence: null,
};

const restInterval = {
  id: 'interval-3',
  type: 'rest',
  duration_sec: 120,
  power_start: 0.40,
  power_end: 0.40,
  cadence: null,
};

describe('WorkoutChart', () => {
  const defaultProps = {
    intervals: [],
    selectedId: null,
    onSelectInterval: vi.fn(),
    onUpdateInterval: vi.fn(),
    ftp: 200,
    width: 800,
    height: 400,
  };

  it('renders an SVG element', () => {
    render(<WorkoutChart {...defaultProps} />);
    expect(screen.getByTestId('workout-chart')).toBeInTheDocument();
  });

  it('renders empty chart with no intervals', () => {
    render(<WorkoutChart {...defaultProps} />);
    const svg = screen.getByTestId('workout-chart');
    // Should have background rect and axes
    expect(svg.querySelectorAll('rect').length).toBeGreaterThan(0);
    expect(svg.querySelectorAll('line').length).toBeGreaterThan(0);
  });

  it('renders steady interval as a rectangle', () => {
    render(
      <WorkoutChart {...defaultProps} intervals={[steadyInterval]} />
    );
    const block = screen.getByTestId('interval-block-interval-1');
    expect(block).toBeInTheDocument();
    // Should contain a rect (the main block)
    expect(block.querySelector('rect')).not.toBeNull();
  });

  it('renders ramp interval as a polygon', () => {
    render(
      <WorkoutChart {...defaultProps} intervals={[rampInterval]} />
    );
    const block = screen.getByTestId('interval-block-interval-2');
    expect(block).toBeInTheDocument();
    expect(block.querySelector('polygon')).not.toBeNull();
  });

  it('renders multiple intervals', () => {
    render(
      <WorkoutChart
        {...defaultProps}
        intervals={[steadyInterval, rampInterval, restInterval]}
      />
    );
    expect(screen.getByTestId('interval-block-interval-1')).toBeInTheDocument();
    expect(screen.getByTestId('interval-block-interval-2')).toBeInTheDocument();
    expect(screen.getByTestId('interval-block-interval-3')).toBeInTheDocument();
  });

  it('calls onSelectInterval when clicking an interval block', () => {
    const onSelect = vi.fn();
    render(
      <WorkoutChart
        {...defaultProps}
        intervals={[steadyInterval]}
        onSelectInterval={onSelect}
      />
    );
    const block = screen.getByTestId('interval-block-interval-1');
    const rect = block.querySelector('rect.interval-block');
    fireEvent.click(rect);
    expect(onSelect).toHaveBeenCalledWith('interval-1');
  });

  it('highlights selected interval', () => {
    render(
      <WorkoutChart
        {...defaultProps}
        intervals={[steadyInterval]}
        selectedId="interval-1"
      />
    );
    const block = screen.getByTestId('interval-block-interval-1');
    const rect = block.querySelector('rect.interval-block');
    // Selected interval should have full opacity
    expect(rect.getAttribute('opacity')).toBe('1');
  });

  it('renders Y-axis labels with FTP percentages', () => {
    render(<WorkoutChart {...defaultProps} intervals={[steadyInterval]} />);
    const svg = screen.getByTestId('workout-chart');
    const texts = svg.querySelectorAll('text');
    const textContents = Array.from(texts).map((t) => t.textContent);
    expect(textContents).toContain('100%');
    expect(textContents).toContain('50%');
  });

  it('renders X-axis time labels when there are intervals', () => {
    render(
      <WorkoutChart {...defaultProps} intervals={[steadyInterval]} />
    );
    const svg = screen.getByTestId('workout-chart');
    const texts = svg.querySelectorAll('text');
    const textContents = Array.from(texts).map((t) => t.textContent);
    expect(textContents).toContain('0m');
  });

  it('renders watts labels on Y-axis', () => {
    render(
      <WorkoutChart {...defaultProps} intervals={[steadyInterval]} ftp={300} />
    );
    const svg = screen.getByTestId('workout-chart');
    const texts = svg.querySelectorAll('text');
    const textContents = Array.from(texts).map((t) => t.textContent);
    expect(textContents).toContain('300W');
  });

  it('deselects when clicking empty area', () => {
    const onSelect = vi.fn();
    render(
      <WorkoutChart
        {...defaultProps}
        intervals={[steadyInterval]}
        onSelectInterval={onSelect}
      />
    );
    const svg = screen.getByTestId('workout-chart');
    fireEvent.click(svg);
    expect(onSelect).toHaveBeenCalledWith(null);
  });

  it('applies zone colors based on power', () => {
    // 40% FTP = Z1 = light blue (#A0C4E8)
    render(
      <WorkoutChart {...defaultProps} intervals={[restInterval]} />
    );
    const block = screen.getByTestId('interval-block-interval-3');
    const rect = block.querySelector('rect.interval-block');
    expect(rect.getAttribute('fill')).toBe('#A0C4E8');
  });
});
