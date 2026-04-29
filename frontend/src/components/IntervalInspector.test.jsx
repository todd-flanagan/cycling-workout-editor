import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import IntervalInspector from './IntervalInspector';

const steadyInterval = {
  id: 'test-1',
  type: 'steady',
  duration_sec: 300,
  power_start: 0.75,
  power_end: 0.75,
  cadence: null,
};

const rampInterval = {
  id: 'test-2',
  type: 'ramp_up',
  duration_sec: 600,
  power_start: 0.50,
  power_end: 1.0,
  cadence: 90,
};

describe('IntervalInspector', () => {
  it('renders empty state when no interval selected', () => {
    render(
      <IntervalInspector
        interval={null}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    expect(screen.getByTestId('interval-inspector')).toBeInTheDocument();
    expect(screen.getByText(/select an interval/i)).toBeInTheDocument();
  });

  it('renders fields for a selected steady interval', () => {
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    expect(screen.getByLabelText(/type/i)).toHaveValue('steady');
    expect(screen.getByLabelText(/duration/i)).toHaveValue('05:00');
    expect(screen.getByLabelText(/power.*ftp/i)).toHaveValue(75);
  });

  it('shows power end field for ramp intervals', () => {
    render(
      <IntervalInspector
        interval={rampInterval}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    expect(screen.getByLabelText(/power start/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/power end/i)).toBeInTheDocument();
  });

  it('does not show power end field for steady intervals', () => {
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    expect(screen.queryByLabelText(/power end/i)).not.toBeInTheDocument();
  });

  it('calls onUpdate when type is changed', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/type/i), { target: { value: 'ramp_up' } });
    expect(onUpdate).toHaveBeenCalledWith('test-1', expect.objectContaining({ type: 'ramp_up' }));
  });

  it('calls onUpdate when duration is changed', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/duration/i), { target: { value: '10:00' } });
    expect(onUpdate).toHaveBeenCalledWith('test-1', { duration_sec: 600 });
  });

  it('calls onUpdate when power start is changed', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/power.*ftp/i), { target: { value: '80' } });
    expect(onUpdate).toHaveBeenCalledWith('test-1', {
      power_start: 0.80,
      power_end: 0.80,
    });
  });

  it('calls onUpdate when power end is changed for ramp', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={rampInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/power end/i), { target: { value: '110' } });
    expect(onUpdate).toHaveBeenCalledWith('test-2', { power_end: 1.10 });
  });

  it('calls onUpdate when cadence is changed', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/cadence/i), { target: { value: '95' } });
    expect(onUpdate).toHaveBeenCalledWith('test-1', { cadence: 95 });
  });

  it('sets cadence to null when cleared', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={rampInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/cadence/i), { target: { value: '' } });
    expect(onUpdate).toHaveBeenCalledWith('test-2', { cadence: null });
  });

  it('calls onRemove when remove button is clicked', () => {
    const onRemove = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={vi.fn()}
        onRemove={onRemove}
        ftp={200}
      />
    );
    fireEvent.click(screen.getByText(/remove interval/i));
    expect(onRemove).toHaveBeenCalledWith('test-1');
  });

  it('displays watts in summary', () => {
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    expect(screen.getByText(/150W/)).toBeInTheDocument();
  });

  it('displays ramp summary with arrow', () => {
    render(
      <IntervalInspector
        interval={rampInterval}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    // Should show start and end watts
    expect(screen.getByText(/100W/)).toBeInTheDocument();
    expect(screen.getByText(/200W/)).toBeInTheDocument();
  });

  it('syncs power_end with power_start for rest type change', () => {
    const onUpdate = vi.fn();
    render(
      <IntervalInspector
        interval={steadyInterval}
        onUpdate={onUpdate}
        onRemove={vi.fn()}
        ftp={200}
      />
    );
    fireEvent.change(screen.getByLabelText(/type/i), { target: { value: 'rest' } });
    expect(onUpdate).toHaveBeenCalledWith('test-1', expect.objectContaining({
      type: 'rest',
      power_start: 0.40,
      power_end: 0.40,
    }));
  });
});
