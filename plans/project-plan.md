# Cycling Workout Editor — Phased Implementation Plan

## Phase Overview

| Phase | Name                        | Summary                                                        | Requirements Covered        |
|-------|-----------------------------|----------------------------------------------------------------|-----------------------------|
| 1     | Foundation + Editor         | Project scaffolding, auth, basic workout editor, user profile  | R1, R2, R4, R9, R11, NF1–4 |
| 2     | Export, Import, and Catalog | Workout library with search/tags, export, import               | R3, R5, R6, R8             |
| 3     | Template Library            | Built-in workout templates, template browser drawer            | R7                         |
| 4     | Training Plans              | Multi-week plan builder, calendar view, plan management        | R10                        |
| 5     | Train My Commute            | Commute profiles, plan-to-commute adaptation engine            | R12                        |
| 6     | Platform Integrations       | Strava/Garmin ingest, Garmin workout push, Wahoo/others        | R13                        |

Each phase produces a deployable, self-hosted system. Later phases build on earlier ones without reworking prior deliverables.

---

## Phase 1 — Foundation + Editor

**Goal**: A running app where a user can log in, set their FTP, and create/edit a single workout using the visual editor. End-to-end infrastructure is in place.

### 1.1 — Project Scaffolding and Build Pipeline

- [ ] Initialize Git repo structure:
  ```
  ├── frontend/          # React app (create-react-app or Vite)
  ├── backend/           # Go module
  ├── docker-compose.yml
  └── data/              # SQLite volume mount point
  ```
- [ ] **Frontend**: Initialize React project (Vite recommended for speed). Configure dev server proxy to backend.
- [ ] **Backend**: Initialize Go module. Set up HTTP router (e.g., `chi` or `gorilla/mux`). Health-check endpoint (`GET /healthz`).
- [ ] **Frontend Dockerfile** (multi-stage): Node build → Nginx serve.
- [ ] **Backend Dockerfile** (multi-stage): Go build (CGO enabled for SQLite) → Alpine runtime.
- [ ] **docker-compose.yml**: Wire up both containers, volume mount for `./data`, Nginx proxy config for `/api/*` and `/auth/*` → backend:8080.
- [ ] Verify: `docker compose up` serves the React app at `:80` and proxies API calls to the Go backend.

### 1.2 — Database Setup

- [ ] Choose a Go SQLite driver (e.g., `modernc.org/sqlite` for pure-Go, or `mattn/go-sqlite3` with CGO).
- [ ] Implement a migration runner (embed SQL files or use a library like `goose`/`golang-migrate`).
- [ ] Create initial migration — `users`, `ftp_history` tables:
  ```sql
  CREATE TABLE users (
    id TEXT PRIMARY KEY,
    google_id TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    display_name TEXT NOT NULL,
    ftp INTEGER NOT NULL DEFAULT 200,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE ftp_history (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    ftp INTEGER NOT NULL,
    recorded_at DATE NOT NULL,
    source TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Implement repository/DAO layer with interfaces (to support a future PostgreSQL swap).

### 1.3 — Authentication (Google OAuth + JWT)

- [ ] Register a Google OAuth 2.0 client (instructions in README for self-hosters).
- [ ] Backend endpoints:
  - `GET /auth/google/login` — redirect to Google consent screen.
  - `GET /auth/google/callback` — exchange code for tokens, extract identity, upsert user, issue JWT.
  - `POST /auth/logout` — client-side token discard (JWT is stateless).
- [ ] JWT implementation:
  - Sign with HS256 + configurable secret (from environment variable).
  - Include `sub` (user ID), `exp` (configurable expiry), `iat`.
  - Middleware to validate JWT on all `/api/*` routes.
- [ ] Frontend:
  - Login screen (S1) with "Sign in with Google" button.
  - Store JWT in memory (or `localStorage` with XSS considerations).
  - Attach `Authorization: Bearer <token>` to all API calls.
  - Redirect to login if token is missing or expired.
- [ ] First-login flow: after account creation, prompt for initial FTP.

### 1.4 — User Profile Page (S6)

- [ ] Backend endpoints:
  - `GET /api/user/profile` — return user record.
  - `PUT /api/user/profile` — update FTP (and other profile fields).
  - `GET /api/user/ftp-history` — return FTP history entries.
  - `POST /api/user/ftp-history` — add FTP entry (direct value, with date and source).
  - `POST /api/user/ftp-from-ramp` — accept max 1-minute power, return 75% as FTP estimate; if confirmed, save to history and update current FTP.
- [ ] Frontend — User Profile screen:
  - Display Google account info (name, email).
  - **Set FTP directly**: input field for FTP value, date picker, optional source note.
  - **Calculate from ramp test**: input field for max 1-minute power. Display calculated FTP (75%). Confirm button to save.
  - FTP history table/chart (SVG line chart or simple table).
  - Log out button.

### 1.5 — Workout Editor (S3) — Core

- [ ] Backend:
  - Create `workouts` table migration:
    ```sql
    CREATE TABLE workouts (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL REFERENCES users(id),
      name TEXT NOT NULL DEFAULT 'Untitled Workout',
      description TEXT DEFAULT '',
      intervals TEXT NOT NULL DEFAULT '[]',  -- JSON
      template_id TEXT REFERENCES templates(id),
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    ```
  - `POST /api/workouts` — create workout.
  - `GET /api/workouts/{id}` — get workout.
  - `PUT /api/workouts/{id}` — update workout.
- [ ] Frontend — Editor layout:
  - **Left toolbar**: draggable interval blocks (steady, ramp up, ramp down, free ride, rest).
  - **Center SVG canvas**: power-over-time bar chart.
    - X-axis: time (minutes). Y-axis: % FTP (with absolute watts as secondary labels).
    - Blocks color-coded by power zone (Z1–Z6).
    - Ramps rendered as sloped polygons.
  - **Upper-right inspector panel** (fixed position): fields for selected interval (type, duration, power start/end, cadence).
- [ ] Frontend — Editor interactions:
  - Click to select a block. Selected block highlighted; inspector populated.
  - Drag top edge of a block to adjust power.
  - Drag left/right edges to adjust duration.
  - Drag to reorder blocks on the timeline.
  - Drag blocks from toolbar onto the timeline to insert.
  - Inspector fields update chart in real time; chart drag updates inspector fields.
- [ ] Frontend — Undo/redo:
  - `useReducer`-based state management for intervals.
  - Maintain undo/redo stack (array of past/future states).
  - Ctrl+Z / Ctrl+Shift+Z.
- [ ] Frontend — Keyboard shortcuts:
  - Delete/Backspace: remove selected interval.
  - Arrow Up/Down: nudge power +/- 1% FTP.
  - Arrow Left/Right: nudge duration +/- 5 seconds.
- [ ] Workout metadata: name and description fields above the chart.
- [ ] Save button → `PUT /api/workouts/{id}`. Cancel button → discard changes, navigate back.

### Phase 1 — Exit Criteria

- [ ] `docker compose up` starts the full stack from scratch.
- [ ] User can sign in with Google, set their FTP, and view FTP history.
- [ ] User can create a new workout, add/remove/reorder intervals visually, edit via inspector, undo/redo, and save.
- [ ] Workout chart renders with zone colors, ramp slopes, and dual-axis labels.
- [ ] Keyboard shortcuts work (delete, arrow nudge, undo/redo).

---

## Phase 2 — Export, Import, and Workout Catalog

**Goal**: Users can manage a library of workouts with search and tags, export workouts to Zwift and text formats, and import `.zwo` and `.fit` files.

### 2.1 — Workout Library (S2)

- [ ] Backend:
  - `GET /api/workouts` — list workouts for current user. Support query params: `?search=`, `?tag=`, `?sort=name|duration|created_at`, `?order=asc|desc`.
  - `DELETE /api/workouts/{id}` — delete workout.
  - `POST /api/workouts/{id}/duplicate` — duplicate workout (copy with "(Copy)" appended to name).
- [ ] Frontend — Library screen:
  - Grid view of workout cards. Each card shows: name, chart thumbnail (small SVG), description excerpt, tag chips, duration, estimated TSS.
  - Search bar (filters by name and description).
  - Tag filter (click a tag chip to filter; show all user tags as a filter bar).
  - Sort controls (name, duration, date created).
  - Per-workout actions: Edit, Duplicate, Delete (with confirmation), Export to Zwift, Export to Text.
  - "New Workout" button → navigate to blank editor.
  - "Import Workout" button → open import dialog (2.3).
  - "Browse Templates" button → open template drawer (Phase 3).

### 2.2 — Tags

- [ ] Backend:
  - Create `workout_tags` table migration:
    ```sql
    CREATE TABLE workout_tags (
      workout_id TEXT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
      tag TEXT NOT NULL,
      PRIMARY KEY (workout_id, tag)
    );
    ```
  - `PUT /api/workouts/{id}/tags` — set tags for a workout (replace all).
  - `GET /api/tags` — list distinct tags for the current user.
- [ ] Frontend:
  - Tag input in the workout editor (S3 metadata area) — free-form text, add on Enter/comma.
  - Tag chips displayed on library cards.
  - Tag filter bar in the library.

### 2.3 — Export

- [ ] Backend — Zwift `.zwo` export:
  - `GET /api/workouts/{id}/export/zwo` — generate `.zwo` XML.
  - Map interval types to Zwift XML elements: `<SteadyState>`, `<Ramp>`, `<FreeRide>`, `<Cooldown>`, `<Warmup>`.
  - Include `<workout_file>` root, `<name>`, `<description>`, `<sportType>`.
  - Response: `Content-Type: application/xml`, `Content-Disposition: attachment; filename="workout-name.zwo"`.
- [ ] Backend — Text export:
  - `GET /api/workouts/{id}/export/text` — generate human-readable text file.
  - Format per the spec in `function/ui.md` (workout name, duration, intervals with %FTP and absolute watts, TSS).
  - Response: `Content-Type: text/plain`, `Content-Disposition: attachment`.
- [ ] Frontend:
  - Export buttons on library cards and in the editor trigger file downloads.

### 2.4 — Import

- [ ] Backend — `.zwo` import:
  - `POST /api/workouts/import` — accept multipart file upload.
  - Parse Zwift XML, map elements back to internal interval model.
  - Return parsed workout for preview (name, description, intervals).
- [ ] Backend — `.fit` import:
  - Use existing Go library (`github.com/tormoder/fit` or `github.com/muktihari/fit`) to parse FIT files.
  - Extract workout/interval data from FIT workout messages.
  - Map to internal interval model. Handle graceful fallback for unsupported workout types.
- [ ] Frontend — Import dialog (S5):
  - File upload (drag-and-drop or file picker). Accept `.zwo` and `.fit`.
  - After upload, show preview: parsed name, description, visual chart.
  - "Import" button to confirm → creates workout in user's library.
  - Error state for invalid/unsupported files with clear messaging.

### Phase 2 — Exit Criteria

- [ ] User can browse their workout library with search, tag filtering, and sorting.
- [ ] User can tag workouts and filter by tag.
- [ ] User can export any workout to `.zwo` (importable by Zwift) and to a text file.
- [ ] User can import `.zwo` and `.fit` files with preview and confirmation.
- [ ] Duplicate and delete work correctly.

---

## Phase 3 — Template Library

**Goal**: The app ships with a set of built-in workout templates. Users can browse them in a modal drawer and add copies to their library.

### 3.1 — Template Database Seed

- [ ] Create `templates` table migration:
  ```sql
  CREATE TABLE templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    intervals TEXT NOT NULL,  -- JSON
    tags TEXT NOT NULL,       -- JSON array
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Create seed data (embedded in Go binary or as a SQL seed file) for the following templates:

#### FTP Ramp Test
- 5:00 warm-up @ 46% FTP
- Ramp: 1:00 steps increasing by ~6% FTP per step, starting at 46% and continuing until exhaustion (model as ramps from 46% → 150%+ FTP over ~20 minutes)
- 5:00 cool-down @ 40% FTP
- Tags: `testing`, `FTP`

#### Sweet Spot (2x20)
- 10:00 warm-up @ 50% FTP
- 20:00 @ 90% FTP
- 5:00 recovery @ 45% FTP
- 20:00 @ 90% FTP
- 5:00 cool-down @ 45% FTP
- Tags: `sweet spot`, `endurance`, `base`

#### VO2max Intervals (5x3)
- 10:00 warm-up @ 55% FTP
- 5 × (3:00 @ 115% FTP + 3:00 recovery @ 45% FTP)
- 5:00 cool-down @ 40% FTP
- Tags: `VO2max`, `intervals`, `high intensity`

#### Endurance Ride
- 5:00 warm-up @ 50% FTP
- 60:00 @ 65% FTP
- 5:00 cool-down @ 45% FTP
- Tags: `endurance`, `Z2`, `base`

#### Threshold Intervals (3x10)
- 10:00 warm-up @ 55% FTP
- 3 × (10:00 @ 100% FTP + 5:00 recovery @ 50% FTP)
- 5:00 cool-down @ 45% FTP
- Tags: `threshold`, `FTP`, `intervals`

#### Over-Unders (3x12)
- 10:00 warm-up @ 55% FTP
- 3 × (12:00 alternating 2:00 @ 95% / 1:00 @ 105% FTP + 5:00 recovery @ 50% FTP)
- 5:00 cool-down @ 45% FTP
- Tags: `threshold`, `over-under`, `lactate`

#### Recovery Ride
- 5:00 warm-up @ 40% FTP
- 30:00 @ 45% FTP
- 5:00 cool-down @ 35% FTP
- Tags: `recovery`, `Z1`, `easy`

- [ ] Seed runs automatically on first startup (idempotent — skip if templates already exist).

### 3.2 — Template API

- [ ] `GET /api/templates` — list all templates.
- [ ] `POST /api/templates/{id}/add` — copy a template into the current user's workout library:
  - Create a new workout with the template's name, description, intervals, and tags.
  - Set `template_id` on the new workout for provenance tracking.
  - Return the newly created workout.

### 3.3 — Template Browser (S4 — Modal Drawer)

- [ ] Frontend — "Browse Templates" button in the library triggers the drawer.
- [ ] Drawer slides in from the right, overlaying the library. Scrim/overlay behind it.
- [ ] Template list inside the drawer:
  - Each template card: name, description, mini SVG chart preview, tags, total duration.
  - "Add to My Library" button per template.
- [ ] On add: call `POST /api/templates/{id}/add`, close drawer, show new workout in library (or navigate to editor for it).
- [ ] Close drawer: click scrim, click X, or press Escape.

### Phase 3 — Exit Criteria

- [ ] 7 templates seeded on first startup (including FTP Ramp Test).
- [ ] Template browser drawer opens from the library, shows all templates with chart previews.
- [ ] User can add any template to their library; it appears as an editable workout.
- [ ] Added workouts from templates have pre-populated tags and a `template_id` reference.

---

## Cross-Cutting Concerns (All Phases)

### Error Handling
- Backend returns structured JSON errors: `{ "error": "message", "code": "ERROR_CODE" }`.
- Frontend shows toast/snackbar notifications for errors and success confirmations.

### Testing Strategy
- **Backend**: Unit tests for repository layer, handler tests for API endpoints, integration tests against an in-memory SQLite database.
- **Frontend**: Component tests with React Testing Library. E2E tests with Playwright or Cypress for critical flows (login → create workout → save → export).

### Security
- All API endpoints behind JWT auth middleware (except `/auth/*` and `/healthz`).
- Input validation on all endpoints (workout name length, interval value ranges, file upload size limits).
- CORS configuration for the frontend origin.
- Rate limiting on auth endpoints.

### Configuration
- Environment variables for:
  - `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
  - `JWT_SECRET`, `JWT_EXPIRY`
  - `DATABASE_PATH`
  - `FRONTEND_URL` (for CORS and OAuth redirect)
- `.env.example` file in the repo.

---

## Phase 4 — Training Plans

**Goal**: Users can organize workouts into multi-day/multi-week training plans and view them on a calendar. This phase lays the groundwork for the "Train My Commute" feature.

### 4.1 — Training Plan Data Model

- [ ] Create `training_plans` table migration:
  ```sql
  CREATE TABLE training_plans (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    start_date DATE NOT NULL,
    weeks INTEGER NOT NULL DEFAULT 4,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE plan_entries (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL REFERENCES training_plans(id) ON DELETE CASCADE,
    workout_id TEXT REFERENCES workouts(id),
    day_offset INTEGER NOT NULL,  -- day within the plan (0-based)
    notes TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Repository/DAO layer for plans and entries.

### 4.2 — Training Plan API

- [ ] `POST /api/plans` — create a new plan.
- [ ] `GET /api/plans` — list user's plans.
- [ ] `GET /api/plans/{id}` — get plan with all entries.
- [ ] `PUT /api/plans/{id}` — update plan metadata.
- [ ] `DELETE /api/plans/{id}` — delete plan.
- [ ] `POST /api/plans/{id}/entries` — add a workout to a day.
- [ ] `PUT /api/plans/{id}/entries/{entryId}` — move/update an entry.
- [ ] `DELETE /api/plans/{id}/entries/{entryId}` — remove an entry.

### 4.3 — Calendar View (Frontend)

- [ ] Calendar UI showing the plan's date range with workouts placed on days.
- [ ] Drag-and-drop workouts from the library onto calendar days.
- [ ] Drag to move workouts between days within the plan.
- [ ] Per-day view: click a day to see the scheduled workout(s) with chart preview.
- [ ] Week summary row: total planned duration, TSS, and intensity distribution.

### Phase 4 — Exit Criteria

- [ ] User can create, edit, and delete training plans.
- [ ] User can assign workouts to specific days on a calendar view.
- [ ] Calendar displays workout chart thumbnails and weekly summaries.
- [ ] Drag-and-drop scheduling works for adding and rearranging workouts.

---

## Phase 5 — Train My Commute

**Goal**: Transform training plans into commute-aware schedules. The system adapts structured plans to fit around a rider's daily commute, making the commute ride the primary training vehicle.

### 5.1 — Commute Profile

- [ ] Create `commute_profiles` table migration:
  ```sql
  CREATE TABLE commute_profiles (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL DEFAULT 'My Commute',
    direction TEXT NOT NULL,  -- 'to_work', 'from_work'
    distance_km REAL,
    typical_duration_min INTEGER NOT NULL,
    elevation_gain_m REAL DEFAULT 0,
    intensity_cap_pct INTEGER,          -- max %FTP (e.g., no showering constraint)
    arrival_deadline TIME,              -- e.g., '08:30'
    days_of_week TEXT NOT NULL,         -- JSON array, e.g., ["mon","tue","wed","thu","fri"]
    notes TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Support asymmetric commutes: separate profiles for to-work and from-work directions.
- [ ] API endpoints for CRUD on commute profiles.
- [ ] Frontend: commute profile editor in the user profile area.

### 5.2 — Plan Adaptation Engine

- [ ] Adaptation algorithm that takes a training plan + commute profile(s) and produces a commute-aware schedule:
  - Map plan workouts onto commute slots where the workout fits within commute duration and intensity constraints.
  - Assign easy/recovery days to commute rides that stay under the intensity cap.
  - Flag workouts that cannot fit a commute slot (too long, too intense, requires specific terrain) as requiring a dedicated ride.
  - Distribute weekly volume across commute rides, accounting for the cumulative load that commuting adds on top of the plan.
- [ ] Output: an adapted calendar where each day shows either a commute-adapted workout, a standalone workout, or a rest day.
- [ ] Allow manual overrides — user can move a workout from commute to dedicated slot or vice versa.

### 5.3 — Compliance Tracking

- [ ] Weekly summary view comparing planned vs. actual:
  - Which commute rides hit the prescribed targets.
  - Remaining plan objectives that need dedicated rides.
  - Cumulative TSS/volume vs. plan intent.
- [ ] Visual indicators on the calendar: completed (green), missed (red), upcoming (blue), adapted-for-commute (commute icon).

### Phase 5 — Exit Criteria

- [ ] User can define asymmetric commute profiles with constraints.
- [ ] The adaptation engine produces a commute-aware schedule from a training plan.
- [ ] Calendar view distinguishes commute workouts from dedicated sessions.
- [ ] Weekly compliance summaries show planned vs. actual training load.

---

## Phase 6 — Platform Integrations

**Goal**: Close the feedback loop between planned and actual training by pulling ride data from Strava/Garmin and pushing scheduled workouts to head units.

### 6.1 — OAuth & Connection Management

- [ ] Implement OAuth 2.0 flows for:
  - **Strava API** — read activity data.
  - **Garmin Connect API** — read activity data + write workouts.
- [ ] Token storage (encrypted), refresh, and revocation handling.
- [ ] Frontend: "Connected Accounts" section in user profile with connect/disconnect buttons and status indicators.
- [ ] Environment variables for client IDs/secrets: `STRAVA_CLIENT_ID`, `STRAVA_CLIENT_SECRET`, `GARMIN_CLIENT_ID`, `GARMIN_CLIENT_SECRET`.

### 6.2 — Inbound: Ride Data Collection

- [ ] **Strava integration**:
  - Register a Strava webhook to receive new activity notifications.
  - On notification, fetch activity details (power, HR, duration, route, TSS/IF).
  - Store activity summaries in a local `activities` table linked to the user.
- [ ] **Garmin Connect integration**:
  - Use Garmin's push API (Health API / activity file push) to receive completed activities.
  - Parse activity data and store alongside Strava activities.
- [ ] **Auto-classification**:
  - Classify imported rides as "commute", "dedicated training", or "other" based on:
    - Route matching against commute profile (start/end location proximity).
    - Time-of-day correlation with commute schedule.
    - User-defined rules and manual override.
- [ ] **Plan compliance assessment**:
  - Compare actual ride metrics against the day's prescribed workout.
  - Update the compliance tracking view (Phase 5.3) with real data.

### 6.3 — Outbound: Publish Workouts to Head Units

- [ ] **Garmin Connect workout push**:
  - Push scheduled workouts to Garmin Connect so they appear on the rider's device.
  - Support both commute-adapted workouts and standalone training sessions.
  - Map internal interval model to Garmin workout format (FIT workout file or Connect API workout structure).
- [ ] **Calendar sync**:
  - Automatically push upcoming workouts (e.g., next 7 days) to Garmin Connect's training calendar.
  - Handle updates — if the user reschedules a workout, update or replace the Garmin entry.
- [ ] Frontend:
  - Per-workout "Send to Garmin" button.
  - Auto-sync toggle in settings: push upcoming workouts automatically.
  - Sync status indicators on calendar entries.

### 6.4 — Future: Additional Head Unit Platforms

- [ ] Design the integration layer with a **provider abstraction** so new platforms are additive:
  - Common interface: `PullActivities()`, `PushWorkout()`, `SyncCalendar()`.
  - Per-provider implementation behind the interface.
- [ ] Planned future providers (not implemented in Phase 6, but architecture supports them):
  - **Wahoo (ELEMNT)** — via Wahoo Cloud API.
  - **Hammerhead (Karoo)** — via Karoo's open platform.
  - **Bryton** and **Stages** — as APIs become available.

### Phase 6 — Exit Criteria

- [ ] User can connect Strava and/or Garmin accounts via OAuth.
- [ ] Completed rides are automatically pulled from Strava (webhook) and Garmin (push API).
- [ ] Imported rides are auto-classified as commute/training/other.
- [ ] Plan compliance view reflects actual ride data.
- [ ] User can push workouts to Garmin Connect; workouts appear on the Garmin device.
- [ ] Auto-sync pushes upcoming workouts to Garmin's training calendar.
- [ ] Integration layer is abstracted for future Wahoo/Hammerhead/Bryton/Stages support.
