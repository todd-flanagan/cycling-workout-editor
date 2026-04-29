# Cycling Workout Editor — Functional / UI Design

## Target Platforms

- **Primary**: Desktop and tablet browsers (landscape orientation).
- **Future**: Mobile browsers for read-only workout viewing.
- React frontend to support responsive evolution over time.

---

## Screens & Flows

### S1 — Login

- Single "Sign in with Google" button.
- On success, redirect to the Workout Library (S2).
- On first login, auto-provision user account and prompt for initial FTP setting.

### S2 — Workout Library

- Grid or list view of the user's saved workouts.
- Each card/row shows:
  - Workout name.
  - Thumbnail of the visual power/time chart.
  - Brief description excerpt.
  - Tags displayed as chips/badges.
  - Duration and estimated TSS (if calculable).
- Actions per workout: **Edit**, **Duplicate**, **Delete**, **Export to Zwift**, **Export to Text**.
- Global actions: **New Workout**, **Import Workout**, **Browse Templates**.
- Search bar and sort/filter controls (by name, duration, date created, tag).
- Tag filter: click a tag to filter the library to matching workouts.

### S3 — Workout Editor

#### Layout

- **Left panel / toolbar**: interval building blocks (steady state, ramp up, ramp down, free ride, rest). Drag a block from the toolbar onto the timeline to add it, or click to append.
- **Center canvas**: visual power-over-time chart that updates in real time as intervals are added/modified.
  - X-axis: time (minutes).
  - Y-axis: power (% FTP on primary axis, absolute watts as secondary label based on user's FTP).
- **Upper-right inspector panel** (fixed position): when an interval block is selected, shows editable fields for that interval. Does not follow the cursor or block — always pinned to the upper-right corner of the editor area.
  - Fields: interval type, duration (mm:ss), target power start (% FTP), target power end (% FTP, for ramps), cadence (optional).
  - Changes in the inspector immediately update the chart.
- Workout metadata fields above the chart: name, description, tags.
- **Save** and **Cancel** actions.

#### Interaction Model (Hybrid)

Two complementary ways to build and edit workouts:

1. **Visual / direct manipulation on the chart**:
   - Drag the **top edge** of a block up/down to adjust target power.
   - Drag the **left/right edges** of a block to adjust duration.
   - Drag blocks to **reorder** them on the timeline.
   - Drag blocks from the left toolbar onto the timeline to insert.
   - Click a block to select it (highlights the block, populates the inspector).

2. **Form-based editing via the inspector panel**:
   - Click a block to select it; edit precise values in the upper-right inspector.
   - Type exact power %, duration, and cadence values.
   - Use for precise numeric entry that's hard to achieve by dragging.

Both methods stay in sync — dragging updates the inspector fields, and typing in the inspector updates the chart.

#### Undo / Redo

- The editor maintains an undo/redo stack for all interval modifications (add, delete, reorder, power change, duration change).
- Standard keyboard shortcuts: **Ctrl+Z** (undo), **Ctrl+Shift+Z** (redo).
- The stack is cleared when the user saves or cancels.

#### Keyboard Shortcuts

| Key            | Action                                      |
|---------------|---------------------------------------------|
| Delete / Backspace | Remove the selected interval            |
| Arrow Up      | Nudge selected interval power up (+1% FTP)  |
| Arrow Down    | Nudge selected interval power down (-1% FTP)|
| Arrow Right   | Nudge selected interval duration longer (+5s)|
| Arrow Left    | Nudge selected interval duration shorter (-5s)|
| Ctrl+Z        | Undo                                        |
| Ctrl+Shift+Z  | Redo                                        |

### S4 — Template Browser (Modal Drawer)

- Opens as a **modal drawer** sliding in from the right side of the Workout Library screen.
- Does not navigate away from the library — the library remains visible behind the drawer overlay.
- Each template shows: name, description, visual chart preview, tags, duration.
- "Add to My Library" action creates an editable copy in the user's workout library.
- Templates include at minimum: FTP Ramp Test, Sweet Spot, VO2max Intervals, Endurance, Threshold Intervals, Over-Unders, Recovery.

### S5 — Import

- Upload dialog accepting `.zwo` and `.fit` files.
- After parsing, show a preview of the imported workout (name, description, visual chart).
- User confirms import to add to their library.
- Clear error messages for invalid or unsupported files.

### S6 — User Profile / Settings

- **Set FTP directly**: enter a known FTP value (with date and optional note/source).
- **Calculate FTP from ramp test**: enter max 1-minute power from a ramp test; the app calculates FTP as 75% of that value. User confirms before saving.
- FTP history: chart or table showing FTP values over time, with source labels.
- View account info (Google display name, email).
- Log out.

---

## Visual Workout Representation

- Bar/block chart style (similar to Zwift/TrainerRoad workout view).
- Color coding by zone:
  - Z1 Recovery (< 55% FTP) — gray/light blue.
  - Z2 Endurance (55–75%) — blue.
  - Z3 Tempo (76–90%) — green.
  - Z4 Threshold (91–105%) — yellow.
  - Z5 VO2max (106–120%) — orange.
  - Z6 Anaerobic (> 120%) — red.
- Ramps rendered as sloped blocks.
- Hover/click on a block shows interval details tooltip.

---

## Text Export Format

Human-readable workout summary, suitable for sharing via text/email:

```
Workout: Sweet Spot 2x20
Duration: 60 min
Description: Two 20-minute sweet spot intervals with recovery.

Intervals:
  1. Warm-up        — 10:00 @ 50% FTP (150W)
  2. Sweet Spot      — 20:00 @ 90% FTP (270W)
  3. Recovery        —  5:00 @ 40% FTP (120W)
  4. Sweet Spot      — 20:00 @ 90% FTP (270W)
  5. Cool-down       —  5:00 @ 40% FTP (120W)

Total TSS: ~75
Based on FTP: 300W
```

---

## Design Decisions

- **Interaction model**: Hybrid — drag-and-drop on the visual chart plus form-based editing in the inspector panel. Both stay in sync.
- **Direct chart editing**: Yes — drag top edge to adjust power, drag side edges to adjust duration.
- **Inspector placement**: Fixed upper-right corner of the editor area. Does not track the selected block.
- **Theme**: Light theme only for initial release.
- **FTP entry**: Two methods — direct FTP entry, or enter max 1-minute ramp power and calculate (75%).
- **Undo/redo**: Full undo/redo stack in the workout editor. Ctrl+Z / Ctrl+Shift+Z.
- **Keyboard shortcuts**: Arrow keys to nudge power/duration, Delete to remove interval.
- **Template browser**: Modal drawer overlaying the library view (not a separate page).
- **Chart rendering**: SVG with React event handlers — simpler to build and debug, native React event model for drag interactions.
- **Session management**: JWT (stateless) — simple, no server-side session store needed.

## Open UI Questions

1. Any accessibility requirements beyond standard WCAG 2.1 AA?
