# Cycling Workout Editor — Phased Implementation Plan

## Phase Overview

| Phase | Name                        | Summary                                                        | Requirements Covered        |
|-------|-----------------------------|----------------------------------------------------------------|-----------------------------|
| 1     | Foundation + Editor         | Project scaffolding, auth, basic workout editor, user profile  | R1, R2, R4, R9, R11, NF1–4 |
| 2     | Export, Import, and Catalog | Workout library with search/tags, export, import               | R3, R5, R6, R8             |
| 3     | Template Library            | Built-in workout templates, template browser drawer            | R7                         |

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
