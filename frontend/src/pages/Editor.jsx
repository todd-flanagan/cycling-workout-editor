import { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { apiGet, apiPost, apiPut } from '../api';
import { useWorkoutReducer } from '../hooks/useWorkoutReducer';
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts';
import WorkoutChart from '../components/WorkoutChart';
import IntervalToolbar from '../components/IntervalToolbar';
import IntervalInspector from '../components/IntervalInspector';

export default function Editor() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const chartContainerRef = useRef(null);

  const [selectedId, setSelectedId] = useState(null);
  const [workoutName, setWorkoutName] = useState('Untitled Workout');
  const [workoutDescription, setWorkoutDescription] = useState('');
  const [workoutId, setWorkoutId] = useState(id || null);
  const [saving, setSaving] = useState(false);
  const [chartWidth, setChartWidth] = useState(800);

  const {
    intervals,
    canUndo,
    canRedo,
    addInterval,
    removeInterval,
    updateInterval,
    reorderIntervals,
    setIntervals,
    undo,
    redo,
  } = useWorkoutReducer([]);

  const ftp = user?.ftp || 200;

  // Load existing workout if editing
  useEffect(() => {
    if (id) {
      apiGet(`/api/workouts/${id}`)
        .then((data) => {
          setWorkoutName(data.name || 'Untitled Workout');
          setWorkoutDescription(data.description || '');
          if (data.intervals) {
            const parsed =
              typeof data.intervals === 'string'
                ? JSON.parse(data.intervals)
                : data.intervals;
            setIntervals(parsed);
          }
          setWorkoutId(data.id);
        })
        .catch(() => {
          // Failed to load workout
        });
    }
  }, [id, setIntervals]);

  // Responsive chart width
  useEffect(() => {
    function updateWidth() {
      if (chartContainerRef.current) {
        setChartWidth(chartContainerRef.current.clientWidth);
      }
    }
    updateWidth();
    window.addEventListener('resize', updateWidth);
    return () => window.removeEventListener('resize', updateWidth);
  }, []);

  const handleSelectInterval = useCallback((intervalId) => {
    setSelectedId(intervalId);
  }, []);

  const handleAddInterval = useCallback(
    (intervalData) => {
      addInterval(intervalData);
    },
    [addInterval]
  );

  const handleRemoveInterval = useCallback(
    (intervalId) => {
      removeInterval(intervalId);
      if (selectedId === intervalId) {
        setSelectedId(null);
      }
    },
    [removeInterval, selectedId]
  );

  const handleUpdateInterval = useCallback(
    (intervalId, changes) => {
      updateInterval(intervalId, changes);
    },
    [updateInterval]
  );

  // Keyboard shortcuts
  useKeyboardShortcuts({
    selectedId,
    removeInterval: handleRemoveInterval,
    updateInterval: handleUpdateInterval,
    undo,
    redo,
    intervals,
  });

  // Handle drop from toolbar
  function handleDrop(e) {
    e.preventDefault();
    try {
      const data = JSON.parse(e.dataTransfer.getData('application/json'));
      addInterval({
        type: data.type,
        ...data.defaults,
      });
    } catch {
      // Invalid drop data
    }
  }

  function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
  }

  async function handleSave() {
    setSaving(true);
    try {
      const payload = {
        name: workoutName,
        description: workoutDescription,
        intervals: JSON.stringify(intervals),
      };

      if (workoutId) {
        await apiPut(`/api/workouts/${workoutId}`, payload);
      } else {
        const result = await apiPost('/api/workouts', payload);
        setWorkoutId(result.id);
        navigate(`/editor/${result.id}`, { replace: true });
      }
    } catch {
      // Save failed
    } finally {
      setSaving(false);
    }
  }

  function handleCancel() {
    navigate('/');
  }

  const selectedInterval = intervals.find((i) => i.id === selectedId) || null;

  const totalDuration = intervals.reduce((sum, i) => sum + i.duration_sec, 0);
  const totalMinutes = Math.floor(totalDuration / 60);
  const totalSeconds = totalDuration % 60;

  return (
    <div className="editor-page" data-testid="editor-page">
      <header className="editor-header">
        <div className="editor-metadata">
          <input
            className="workout-name-input"
            type="text"
            value={workoutName}
            onChange={(e) => setWorkoutName(e.target.value)}
            placeholder="Workout Name"
            aria-label="Workout name"
          />
          <input
            className="workout-desc-input"
            type="text"
            value={workoutDescription}
            onChange={(e) => setWorkoutDescription(e.target.value)}
            placeholder="Description (optional)"
            aria-label="Workout description"
          />
          <span className="workout-duration">
            Total: {totalMinutes}:{String(totalSeconds).padStart(2, '0')}
          </span>
        </div>
        <div className="editor-actions">
          <button
            className="btn btn-sm"
            onClick={undo}
            disabled={!canUndo}
            title="Undo (Ctrl+Z)"
          >
            Undo
          </button>
          <button
            className="btn btn-sm"
            onClick={redo}
            disabled={!canRedo}
            title="Redo (Ctrl+Shift+Z)"
          >
            Redo
          </button>
          <button className="btn btn-secondary" onClick={() => navigate('/profile')}>
            Profile
          </button>
          <button className="btn btn-secondary" onClick={handleCancel}>
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </header>

      <div className="editor-layout">
        <aside className="editor-sidebar">
          <IntervalToolbar onAddInterval={handleAddInterval} />
        </aside>

        <main
          className="editor-canvas"
          ref={chartContainerRef}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
        >
          <WorkoutChart
            intervals={intervals}
            selectedId={selectedId}
            onSelectInterval={handleSelectInterval}
            onUpdateInterval={handleUpdateInterval}
            ftp={ftp}
            width={chartWidth}
            height={400}
          />
        </main>

        <aside className="editor-inspector">
          <IntervalInspector
            interval={selectedInterval}
            onUpdate={handleUpdateInterval}
            onRemove={handleRemoveInterval}
            ftp={ftp}
          />
        </aside>
      </div>
    </div>
  );
}
