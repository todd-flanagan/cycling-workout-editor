import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  workoutReducer,
  createInitialState,
  createInterval,
  ActionTypes,
  useWorkoutReducer,
} from './useWorkoutReducer';

// Mock crypto.randomUUID
beforeEach(() => {
  let counter = 0;
  vi.stubGlobal('crypto', {
    randomUUID: () => `test-uuid-${++counter}`,
  });
});

describe('createInterval', () => {
  it('creates an interval with defaults', () => {
    const interval = createInterval();
    expect(interval.type).toBe('steady');
    expect(interval.duration_sec).toBe(300);
    expect(interval.power_start).toBe(0.75);
    expect(interval.power_end).toBe(0.75);
    expect(interval.cadence).toBeNull();
    expect(interval.id).toBeTruthy();
  });

  it('accepts overrides', () => {
    const interval = createInterval({ type: 'ramp_up', power_start: 0.50, power_end: 1.0 });
    expect(interval.type).toBe('ramp_up');
    expect(interval.power_start).toBe(0.50);
    expect(interval.power_end).toBe(1.0);
  });
});

describe('workoutReducer', () => {
  let state;

  beforeEach(() => {
    state = createInitialState([]);
  });

  describe('ADD_INTERVAL', () => {
    it('adds an interval to the end', () => {
      const result = workoutReducer(state, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'steady', power_start: 0.75, power_end: 0.75, duration_sec: 300 },
      });
      expect(result.intervals).toHaveLength(1);
      expect(result.intervals[0].type).toBe('steady');
    });

    it('adds interval at specific index', () => {
      const stateWith2 = {
        ...state,
        intervals: [
          createInterval({ type: 'steady' }),
          createInterval({ type: 'rest' }),
        ],
      };
      const result = workoutReducer(stateWith2, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'ramp_up' },
        index: 1,
      });
      expect(result.intervals).toHaveLength(3);
      expect(result.intervals[1].type).toBe('ramp_up');
    });

    it('pushes current state to past', () => {
      const result = workoutReducer(state, {
        type: ActionTypes.ADD_INTERVAL,
        payload: {},
      });
      expect(result.past).toHaveLength(1);
      expect(result.future).toHaveLength(0);
    });
  });

  describe('REMOVE_INTERVAL', () => {
    it('removes an interval by id', () => {
      const interval = createInterval({ type: 'steady' });
      const stateWith1 = { ...state, intervals: [interval] };
      const result = workoutReducer(stateWith1, {
        type: ActionTypes.REMOVE_INTERVAL,
        payload: interval.id,
      });
      expect(result.intervals).toHaveLength(0);
    });

    it('does not remove if id does not match', () => {
      const interval = createInterval({ type: 'steady' });
      const stateWith1 = { ...state, intervals: [interval] };
      const result = workoutReducer(stateWith1, {
        type: ActionTypes.REMOVE_INTERVAL,
        payload: 'nonexistent-id',
      });
      expect(result.intervals).toHaveLength(1);
    });
  });

  describe('UPDATE_INTERVAL', () => {
    it('updates an interval by id', () => {
      const interval = createInterval({ type: 'steady', duration_sec: 300 });
      const stateWith1 = { ...state, intervals: [interval] };
      const result = workoutReducer(stateWith1, {
        type: ActionTypes.UPDATE_INTERVAL,
        payload: { id: interval.id, changes: { duration_sec: 600 } },
      });
      expect(result.intervals[0].duration_sec).toBe(600);
      expect(result.intervals[0].type).toBe('steady');
    });
  });

  describe('REORDER_INTERVALS', () => {
    it('reorders intervals', () => {
      const i1 = createInterval({ type: 'steady' });
      const i2 = createInterval({ type: 'rest' });
      const i3 = createInterval({ type: 'ramp_up' });
      const stateWith3 = { ...state, intervals: [i1, i2, i3] };

      const result = workoutReducer(stateWith3, {
        type: ActionTypes.REORDER_INTERVALS,
        payload: { fromIndex: 2, toIndex: 0 },
      });
      expect(result.intervals[0].type).toBe('ramp_up');
      expect(result.intervals[1].type).toBe('steady');
      expect(result.intervals[2].type).toBe('rest');
    });
  });

  describe('SET_INTERVALS', () => {
    it('replaces all intervals', () => {
      const newIntervals = [
        createInterval({ type: 'ramp_up' }),
        createInterval({ type: 'steady' }),
      ];
      const result = workoutReducer(state, {
        type: ActionTypes.SET_INTERVALS,
        payload: newIntervals,
      });
      expect(result.intervals).toHaveLength(2);
      expect(result.intervals[0].type).toBe('ramp_up');
    });
  });

  describe('UNDO', () => {
    it('restores previous state', () => {
      // Add an interval
      const after = workoutReducer(state, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'steady' },
      });
      expect(after.intervals).toHaveLength(1);

      // Undo
      const undone = workoutReducer(after, { type: ActionTypes.UNDO });
      expect(undone.intervals).toHaveLength(0);
      expect(undone.future).toHaveLength(1);
    });

    it('does nothing if no past states', () => {
      const result = workoutReducer(state, { type: ActionTypes.UNDO });
      expect(result).toBe(state);
    });
  });

  describe('REDO', () => {
    it('restores next future state', () => {
      // Add an interval then undo
      const after = workoutReducer(state, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'steady' },
      });
      const undone = workoutReducer(after, { type: ActionTypes.UNDO });
      expect(undone.intervals).toHaveLength(0);

      // Redo
      const redone = workoutReducer(undone, { type: ActionTypes.REDO });
      expect(redone.intervals).toHaveLength(1);
    });

    it('does nothing if no future states', () => {
      const result = workoutReducer(state, { type: ActionTypes.REDO });
      expect(result).toBe(state);
    });
  });

  describe('undo/redo stack management', () => {
    it('clears future on new action after undo', () => {
      const s1 = workoutReducer(state, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'steady' },
      });
      const s2 = workoutReducer(s1, { type: ActionTypes.UNDO });
      expect(s2.future).toHaveLength(1);

      // New action clears future
      const s3 = workoutReducer(s2, {
        type: ActionTypes.ADD_INTERVAL,
        payload: { type: 'rest' },
      });
      expect(s3.future).toHaveLength(0);
    });
  });

  it('returns state for unknown action', () => {
    const result = workoutReducer(state, { type: 'UNKNOWN' });
    expect(result).toBe(state);
  });
});

describe('useWorkoutReducer hook', () => {
  it('initializes with empty intervals', () => {
    const { result } = renderHook(() => useWorkoutReducer([]));
    expect(result.current.intervals).toEqual([]);
    expect(result.current.canUndo).toBe(false);
    expect(result.current.canRedo).toBe(false);
  });

  it('addInterval adds an interval', () => {
    const { result } = renderHook(() => useWorkoutReducer([]));

    act(() => {
      result.current.addInterval({ type: 'steady', power_start: 0.75, power_end: 0.75, duration_sec: 300 });
    });

    expect(result.current.intervals).toHaveLength(1);
    expect(result.current.canUndo).toBe(true);
  });

  it('removeInterval removes an interval', () => {
    const { result } = renderHook(() => useWorkoutReducer([]));

    act(() => {
      result.current.addInterval({ type: 'steady' });
    });

    const id = result.current.intervals[0].id;

    act(() => {
      result.current.removeInterval(id);
    });

    expect(result.current.intervals).toHaveLength(0);
  });

  it('updateInterval updates an interval', () => {
    const { result } = renderHook(() => useWorkoutReducer([]));

    act(() => {
      result.current.addInterval({ type: 'steady', duration_sec: 300 });
    });

    const id = result.current.intervals[0].id;

    act(() => {
      result.current.updateInterval(id, { duration_sec: 600 });
    });

    expect(result.current.intervals[0].duration_sec).toBe(600);
  });

  it('undo/redo work correctly', () => {
    const { result } = renderHook(() => useWorkoutReducer([]));

    act(() => {
      result.current.addInterval({ type: 'steady' });
    });

    expect(result.current.intervals).toHaveLength(1);

    act(() => {
      result.current.undo();
    });

    expect(result.current.intervals).toHaveLength(0);
    expect(result.current.canRedo).toBe(true);

    act(() => {
      result.current.redo();
    });

    expect(result.current.intervals).toHaveLength(1);
  });
});
