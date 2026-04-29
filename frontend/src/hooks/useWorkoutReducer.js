import { useReducer, useCallback } from 'react';

/**
 * Creates a default interval object.
 */
export function createInterval(overrides = {}) {
  return {
    id: crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}-${Math.random()}`,
    type: 'steady',
    duration_sec: 300,
    power_start: 0.75,
    power_end: 0.75,
    cadence: null,
    ...overrides,
  };
}

export const ActionTypes = {
  ADD_INTERVAL: 'ADD_INTERVAL',
  REMOVE_INTERVAL: 'REMOVE_INTERVAL',
  UPDATE_INTERVAL: 'UPDATE_INTERVAL',
  REORDER_INTERVALS: 'REORDER_INTERVALS',
  SET_INTERVALS: 'SET_INTERVALS',
  UNDO: 'UNDO',
  REDO: 'REDO',
};

function pushHistory(state) {
  return {
    past: [...state.past, state.intervals],
    future: [],
  };
}

export function workoutReducer(state, action) {
  switch (action.type) {
    case ActionTypes.ADD_INTERVAL: {
      const history = pushHistory(state);
      const interval = createInterval(action.payload);
      const index = action.index !== undefined ? action.index : state.intervals.length;
      const newIntervals = [...state.intervals];
      newIntervals.splice(index, 0, interval);
      return {
        ...state,
        intervals: newIntervals,
        ...history,
      };
    }

    case ActionTypes.REMOVE_INTERVAL: {
      const history = pushHistory(state);
      return {
        ...state,
        intervals: state.intervals.filter((i) => i.id !== action.payload),
        ...history,
      };
    }

    case ActionTypes.UPDATE_INTERVAL: {
      const history = pushHistory(state);
      return {
        ...state,
        intervals: state.intervals.map((i) =>
          i.id === action.payload.id ? { ...i, ...action.payload.changes } : i
        ),
        ...history,
      };
    }

    case ActionTypes.REORDER_INTERVALS: {
      const history = pushHistory(state);
      const { fromIndex, toIndex } = action.payload;
      const newIntervals = [...state.intervals];
      const [moved] = newIntervals.splice(fromIndex, 1);
      newIntervals.splice(toIndex, 0, moved);
      return {
        ...state,
        intervals: newIntervals,
        ...history,
      };
    }

    case ActionTypes.SET_INTERVALS: {
      const history = pushHistory(state);
      return {
        ...state,
        intervals: action.payload,
        ...history,
      };
    }

    case ActionTypes.UNDO: {
      if (state.past.length === 0) return state;
      const previous = state.past[state.past.length - 1];
      return {
        ...state,
        intervals: previous,
        past: state.past.slice(0, -1),
        future: [state.intervals, ...state.future],
      };
    }

    case ActionTypes.REDO: {
      if (state.future.length === 0) return state;
      const next = state.future[0];
      return {
        ...state,
        intervals: next,
        past: [...state.past, state.intervals],
        future: state.future.slice(1),
      };
    }

    default:
      return state;
  }
}

export function createInitialState(intervals = []) {
  return {
    intervals,
    past: [],
    future: [],
  };
}

export function useWorkoutReducer(initialIntervals = []) {
  const [state, dispatch] = useReducer(
    workoutReducer,
    initialIntervals,
    createInitialState
  );

  const addInterval = useCallback(
    (intervalData, index) => {
      dispatch({ type: ActionTypes.ADD_INTERVAL, payload: intervalData, index });
    },
    []
  );

  const removeInterval = useCallback((id) => {
    dispatch({ type: ActionTypes.REMOVE_INTERVAL, payload: id });
  }, []);

  const updateInterval = useCallback((id, changes) => {
    dispatch({ type: ActionTypes.UPDATE_INTERVAL, payload: { id, changes } });
  }, []);

  const reorderIntervals = useCallback((fromIndex, toIndex) => {
    dispatch({
      type: ActionTypes.REORDER_INTERVALS,
      payload: { fromIndex, toIndex },
    });
  }, []);

  const setIntervals = useCallback((intervals) => {
    dispatch({ type: ActionTypes.SET_INTERVALS, payload: intervals });
  }, []);

  const undo = useCallback(() => {
    dispatch({ type: ActionTypes.UNDO });
  }, []);

  const redo = useCallback(() => {
    dispatch({ type: ActionTypes.REDO });
  }, []);

  return {
    intervals: state.intervals,
    canUndo: state.past.length > 0,
    canRedo: state.future.length > 0,
    addInterval,
    removeInterval,
    updateInterval,
    reorderIntervals,
    setIntervals,
    undo,
    redo,
    dispatch,
  };
}
