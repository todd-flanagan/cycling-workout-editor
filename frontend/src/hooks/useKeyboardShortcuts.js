import { useEffect } from 'react';

/**
 * Hook that registers keyboard shortcuts for the workout editor.
 *
 * @param {object} params
 * @param {string|null} params.selectedId - ID of the currently selected interval
 * @param {function} params.removeInterval - Remove an interval by ID
 * @param {function} params.updateInterval - Update an interval: (id, changes)
 * @param {function} params.undo - Undo last action
 * @param {function} params.redo - Redo last undone action
 * @param {Array} params.intervals - Current intervals array
 */
export function useKeyboardShortcuts({
  selectedId,
  removeInterval,
  updateInterval,
  undo,
  redo,
  intervals,
}) {
  useEffect(() => {
    function handleKeyDown(e) {
      // Ignore events from input/textarea/select elements
      const tag = (e.target.tagName || '').toLowerCase();
      if (tag === 'input' || tag === 'textarea' || tag === 'select') {
        return;
      }

      // Ctrl+Z / Cmd+Z = undo, Ctrl+Shift+Z / Cmd+Shift+Z = redo
      if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
        e.preventDefault();
        if (e.shiftKey) {
          redo();
        } else {
          undo();
        }
        return;
      }

      // Remaining shortcuts require a selected interval
      if (!selectedId) return;

      const interval = intervals.find((i) => i.id === selectedId);
      if (!interval) return;

      switch (e.key) {
        case 'Delete':
        case 'Backspace': {
          e.preventDefault();
          removeInterval(selectedId);
          break;
        }

        case 'ArrowUp': {
          e.preventDefault();
          const newStart = Math.min(interval.power_start + 0.01, 2.0);
          const newEnd = Math.min(interval.power_end + 0.01, 2.0);
          updateInterval(selectedId, {
            power_start: Math.round(newStart * 100) / 100,
            power_end: Math.round(newEnd * 100) / 100,
          });
          break;
        }

        case 'ArrowDown': {
          e.preventDefault();
          const newStart = Math.max(interval.power_start - 0.01, 0.1);
          const newEnd = Math.max(interval.power_end - 0.01, 0.1);
          updateInterval(selectedId, {
            power_start: Math.round(newStart * 100) / 100,
            power_end: Math.round(newEnd * 100) / 100,
          });
          break;
        }

        case 'ArrowRight': {
          e.preventDefault();
          const newDuration = interval.duration_sec + 5;
          updateInterval(selectedId, { duration_sec: newDuration });
          break;
        }

        case 'ArrowLeft': {
          e.preventDefault();
          const newDuration = Math.max(interval.duration_sec - 5, 5);
          updateInterval(selectedId, { duration_sec: newDuration });
          break;
        }

        default:
          break;
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedId, removeInterval, updateInterval, undo, redo, intervals]);
}
