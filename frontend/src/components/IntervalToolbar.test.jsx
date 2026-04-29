import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import IntervalToolbar, { INTERVAL_TYPES } from './IntervalToolbar';

describe('IntervalToolbar', () => {
  it('renders the toolbar', () => {
    render(<IntervalToolbar onAddInterval={vi.fn()} />);
    expect(screen.getByTestId('interval-toolbar')).toBeInTheDocument();
  });

  it('renders all interval types', () => {
    render(<IntervalToolbar onAddInterval={vi.fn()} />);
    expect(screen.getByTestId('toolbar-item-steady')).toBeInTheDocument();
    expect(screen.getByTestId('toolbar-item-ramp_up')).toBeInTheDocument();
    expect(screen.getByTestId('toolbar-item-ramp_down')).toBeInTheDocument();
    expect(screen.getByTestId('toolbar-item-free_ride')).toBeInTheDocument();
    expect(screen.getByTestId('toolbar-item-rest')).toBeInTheDocument();
  });

  it('displays labels for each interval type', () => {
    render(<IntervalToolbar onAddInterval={vi.fn()} />);
    expect(screen.getByText('Steady State')).toBeInTheDocument();
    expect(screen.getByText('Ramp Up')).toBeInTheDocument();
    expect(screen.getByText('Ramp Down')).toBeInTheDocument();
    expect(screen.getByText('Free Ride')).toBeInTheDocument();
    expect(screen.getByText('Rest')).toBeInTheDocument();
  });

  it('calls onAddInterval when clicking a toolbar item', () => {
    const onAddInterval = vi.fn();
    render(<IntervalToolbar onAddInterval={onAddInterval} />);

    fireEvent.click(screen.getByTestId('toolbar-item-steady'));
    expect(onAddInterval).toHaveBeenCalledWith({
      type: 'steady',
      power_start: 0.75,
      power_end: 0.75,
      duration_sec: 300,
    });
  });

  it('calls onAddInterval with ramp_up defaults', () => {
    const onAddInterval = vi.fn();
    render(<IntervalToolbar onAddInterval={onAddInterval} />);

    fireEvent.click(screen.getByTestId('toolbar-item-ramp_up'));
    expect(onAddInterval).toHaveBeenCalledWith({
      type: 'ramp_up',
      power_start: 0.50,
      power_end: 1.0,
      duration_sec: 300,
    });
  });

  it('calls onAddInterval with rest defaults', () => {
    const onAddInterval = vi.fn();
    render(<IntervalToolbar onAddInterval={onAddInterval} />);

    fireEvent.click(screen.getByTestId('toolbar-item-rest'));
    expect(onAddInterval).toHaveBeenCalledWith({
      type: 'rest',
      power_start: 0.40,
      power_end: 0.40,
      duration_sec: 300,
    });
  });

  it('toolbar items are draggable', () => {
    render(<IntervalToolbar onAddInterval={vi.fn()} />);
    const item = screen.getByTestId('toolbar-item-steady');
    expect(item.getAttribute('draggable')).toBe('true');
  });

  it('sets data transfer on drag start', () => {
    render(<IntervalToolbar onAddInterval={vi.fn()} />);
    const item = screen.getByTestId('toolbar-item-steady');

    const mockDataTransfer = {
      setData: vi.fn(),
      effectAllowed: '',
    };

    fireEvent.dragStart(item, { dataTransfer: mockDataTransfer });
    expect(mockDataTransfer.setData).toHaveBeenCalledWith(
      'application/json',
      expect.any(String)
    );
    expect(mockDataTransfer.effectAllowed).toBe('copy');
  });

  it('handles keyboard activation (Enter)', () => {
    const onAddInterval = vi.fn();
    render(<IntervalToolbar onAddInterval={onAddInterval} />);
    const item = screen.getByTestId('toolbar-item-steady');

    fireEvent.keyDown(item, { key: 'Enter' });
    expect(onAddInterval).toHaveBeenCalled();
  });

  it('handles keyboard activation (Space)', () => {
    const onAddInterval = vi.fn();
    render(<IntervalToolbar onAddInterval={onAddInterval} />);
    const item = screen.getByTestId('toolbar-item-steady');

    fireEvent.keyDown(item, { key: ' ' });
    expect(onAddInterval).toHaveBeenCalled();
  });

  it('exports INTERVAL_TYPES array with expected length', () => {
    expect(INTERVAL_TYPES).toHaveLength(5);
  });
});
