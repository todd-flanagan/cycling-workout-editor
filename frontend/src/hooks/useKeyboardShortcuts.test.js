import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';

describe('useKeyboardShortcuts', () => {
  let removeInterval;
  let updateInterval;
  let undo;
  let redo;
  let intervals;

  beforeEach(() => {
    removeInterval = vi.fn();
    updateInterval = vi.fn();
    undo = vi.fn();
    redo = vi.fn();
    intervals = [
      {
        id: 'test-1',
        type: 'steady',
        duration_sec: 300,
        power_start: 0.75,
        power_end: 0.75,
        cadence: null,
      },
    ];
  });

  function fireKeyDown(key, options = {}) {
    const event = new KeyboardEvent('keydown', {
      key,
      bubbles: true,
      ...options,
    });
    window.dispatchEvent(event);
  }

  it('calls undo on Ctrl+Z', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: null,
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('z', { ctrlKey: true });
    expect(undo).toHaveBeenCalledOnce();
  });

  it('calls redo on Ctrl+Shift+Z', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: null,
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('z', { ctrlKey: true, shiftKey: true });
    expect(redo).toHaveBeenCalledOnce();
  });

  it('calls removeInterval on Delete key when interval is selected', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('Delete');
    expect(removeInterval).toHaveBeenCalledWith('test-1');
  });

  it('calls removeInterval on Backspace key when interval is selected', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('Backspace');
    expect(removeInterval).toHaveBeenCalledWith('test-1');
  });

  it('does not call removeInterval if no interval selected', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: null,
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('Delete');
    expect(removeInterval).not.toHaveBeenCalled();
  });

  it('nudges power up on ArrowUp', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowUp');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      power_start: 0.76,
      power_end: 0.76,
    });
  });

  it('nudges power down on ArrowDown', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowDown');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      power_start: 0.74,
      power_end: 0.74,
    });
  });

  it('nudges duration longer on ArrowRight', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowRight');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      duration_sec: 305,
    });
  });

  it('nudges duration shorter on ArrowLeft', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowLeft');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      duration_sec: 295,
    });
  });

  it('does not nudge duration below 5 seconds', () => {
    intervals[0].duration_sec = 5;

    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowLeft');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      duration_sec: 5,
    });
  });

  it('does not nudge power above 2.0', () => {
    intervals[0].power_start = 2.0;
    intervals[0].power_end = 2.0;

    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    fireKeyDown('ArrowUp');
    expect(updateInterval).toHaveBeenCalledWith('test-1', {
      power_start: 2.0,
      power_end: 2.0,
    });
  });

  it('ignores shortcuts when target is an input element', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        selectedId: 'test-1',
        removeInterval,
        updateInterval,
        undo,
        redo,
        intervals,
      })
    );

    const input = document.createElement('input');
    document.body.appendChild(input);
    input.focus();

    const event = new KeyboardEvent('keydown', {
      key: 'Delete',
      bubbles: true,
    });
    Object.defineProperty(event, 'target', { value: input });
    window.dispatchEvent(event);

    expect(removeInterval).not.toHaveBeenCalled();
    document.body.removeChild(input);
  });
});
