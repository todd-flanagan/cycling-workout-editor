# Cycling Workout Editor — Architecture

## High-Level Overview

```
┌──────────────────┐      HTTPS      ┌──────────────────┐      ┌──────────┐
│   Browser        │ ◄──────────────► │   Go API Server  │ ◄──► │  SQLite  │
│   (React SPA)    │                  │   (container)    │      │  (file)  │
└──────────────────┘                  └──────┬───────────┘      └──────────┘
        ▲                                    │
        │  served by                         │ OAuth 2.0
        │                                    ▼
┌──────────────────┐                  ┌──────────────┐
│  Nginx / static  │                  │   Google     │
│  (container)     │                  │   Identity   │
└──────────────────┘                  └──────────────┘
```

## Containers

| Container       | Contents                                  | Port |
|----------------|-------------------------------------------|------|
| **frontend**   | Nginx serving the React production build   | 80   |
| **backend**    | Go API server binary                       | 8080 |

- Docker Compose orchestrates both containers for self-hosted deployment.
- SQLite database file is volume-mounted into the backend container for persistence.

---

## Frontend

- **Technology**: React (JavaScript).
- Built as a static SPA, served by Nginx in its own container.
- Communicates with the backend via a JSON REST API (proxied through Nginx or direct).
- Renders workout charts client-side using **SVG with React event handlers** — chosen for simplicity, native React event model, and easier debugging over Canvas.
- Undo/redo state management in the editor (client-side stack, likely via useReducer or a lightweight state library).
- Desktop/tablet-first responsive layout; mobile viewing as a future milestone.

## Backend

- **Technology**: Go.
- Exposes a REST API for CRUD operations on workouts, user profile, FTP history, tags, and template management.
- Handles Google OAuth 2.0 flow (authorization code grant).
- Manages sessions via **JWT** (stateless — no server-side session store).
- Generates `.zwo` and text export files on demand.
- Parses `.zwo` and `.fit` import files.
- Seeds template workouts into the database on first startup (migration/seed step).
- FTP calculation endpoint: accepts max 1-minute ramp power, returns 75% as estimated FTP.

## Database

- **SQLite** — single file, volume-mounted.
- Suitable for the self-hosted, moderate-concurrency use case.
- If scale demands it later, the data layer can be swapped to PostgreSQL with minimal changes (use a repository/DAO pattern).

---

## Authentication Flow

1. User clicks "Sign in with Google".
2. Frontend redirects to Google's OAuth consent screen.
3. Google redirects back with an authorization code.
4. Backend exchanges code for tokens, extracts user identity.
5. Backend creates or retrieves user record, issues session token.
6. Frontend stores session token and uses it for subsequent API calls.

---

## Data Model (Draft)

### User
| Field        | Type     | Notes                        |
|-------------|----------|------------------------------|
| id          | UUID     | Primary key                  |
| google_id   | string   | Google subject identifier    |
| email       | string   | From Google profile          |
| display_name| string   | From Google profile          |
| ftp         | int      | Current FTP in watts         |
| created_at  | timestamp|                              |
| updated_at  | timestamp|                              |

### FTP History
| Field        | Type     | Notes                                |
|-------------|----------|--------------------------------------|
| id          | UUID     | Primary key                          |
| user_id     | UUID     | Foreign key → User                   |
| ftp         | int      | FTP value in watts                   |
| recorded_at | date     | Date of the FTP measurement          |
| source      | string?  | Optional note (e.g., "ramp test", "20-min test") |
| created_at  | timestamp|                                      |

### Template
| Field        | Type     | Notes                                        |
|-------------|----------|----------------------------------------------|
| id          | UUID     | Primary key                                  |
| name        | string   | Template title                               |
| description | text     | Free-form notes                              |
| intervals   | JSON     | Structured interval data                     |
| tags        | JSON     | Default tags for the template                |
| created_at  | timestamp| Seeded on first DB migration                 |

Templates are seeded into this table on first startup. They are read-only system records. When a user adds a template to their library, a copy is inserted into the Workout table.

### Workout
| Field        | Type     | Notes                              |
|-------------|----------|------------------------------------|
| id          | UUID     | Primary key                        |
| user_id     | UUID     | Foreign key → User                 |
| name        | string   | Workout title                      |
| description | text     | Free-form notes                    |
| intervals   | JSON     | Structured interval data           |
| template_id | UUID?    | FK → Template if created from one  |
| created_at  | timestamp|                                    |
| updated_at  | timestamp|                                    |

### Workout Tags
| Field        | Type     | Notes                              |
|-------------|----------|------------------------------------|
| workout_id  | UUID     | Foreign key → Workout              |
| tag         | string   | Tag label (e.g., "VO2max")         |

Compound primary key: (workout_id, tag).

### Interval (within Workout JSON)
| Field         | Type   | Notes                                    |
|--------------|--------|------------------------------------------|
| type         | enum   | steady, ramp, free_ride, rest            |
| duration_sec | int    | Duration in seconds                      |
| power_start  | float  | Starting power as fraction of FTP (e.g., 0.75) |
| power_end    | float  | Ending power (same as start for steady)  |
| cadence      | int?   | Optional target cadence                  |

---

## API Endpoints (Draft)

### Auth
| Method | Path                          | Description                  |
|--------|-------------------------------|------------------------------|
| GET    | /auth/google/login            | Initiate OAuth flow          |
| GET    | /auth/google/callback         | OAuth callback               |
| POST   | /auth/logout                  | End session                  |

### User Profile
| Method | Path                          | Description                  |
|--------|-------------------------------|------------------------------|
| GET    | /api/user/profile             | Get current user profile     |
| PUT    | /api/user/profile             | Update profile (FTP, etc.)   |
| GET    | /api/user/ftp-history         | Get FTP history              |
| POST   | /api/user/ftp-history         | Add FTP history entry        |
| POST   | /api/user/ftp-from-ramp       | Calculate FTP from max 1-min ramp power (75%) |

### Workouts
| Method | Path                              | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | /api/workouts                     | List user's workouts         |
| POST   | /api/workouts                     | Create workout               |
| GET    | /api/workouts/{id}                | Get single workout           |
| PUT    | /api/workouts/{id}                | Update workout               |
| DELETE | /api/workouts/{id}                | Delete workout               |
| POST   | /api/workouts/{id}/duplicate      | Duplicate workout            |
| GET    | /api/workouts/{id}/export/zwo     | Export as .zwo file          |
| GET    | /api/workouts/{id}/export/text    | Export as text file          |

### Import
| Method | Path                              | Description                  |
|--------|-----------------------------------|------------------------------|
| POST   | /api/workouts/import              | Import a .zwo or .fit file   |

### Templates
| Method | Path                              | Description                        |
|--------|-----------------------------------|------------------------------------|
| GET    | /api/templates                    | List available templates           |
| POST   | /api/templates/{id}/add           | Copy template to user's library    |

### Tags
| Method | Path                              | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | /api/tags                         | List all tags for current user |
| PUT    | /api/workouts/{id}/tags           | Set tags for a workout         |

---

## Build & Deployment

### Build Process
```
├── frontend/
│   ├── Dockerfile          # Node build stage → Nginx serve stage
│   └── src/                # React app source
├── backend/
│   ├── Dockerfile          # Go build stage → scratch/alpine runtime
│   └── cmd/                # Go source
├── docker-compose.yml      # Orchestrates frontend + backend
└── data/
    └── workouts.db         # SQLite file (volume-mounted)
```

### Frontend Dockerfile (multi-stage)
1. **Build stage**: `node` image, `npm install`, `npm run build`.
2. **Serve stage**: `nginx:alpine`, copy build output to `/usr/share/nginx/html`.

### Backend Dockerfile (multi-stage)
1. **Build stage**: `golang` image, compile binary with CGO enabled (required for SQLite).
2. **Runtime stage**: `alpine` with libc, copy binary, expose port 8080.

### Docker Compose
- `frontend`: builds from `frontend/Dockerfile`, exposes port 80.
- `backend`: builds from `backend/Dockerfile`, exposes port 8080, mounts `./data` volume for SQLite.
- Nginx config in the frontend container proxies `/api/*` and `/auth/*` requests to the backend.

---

## Resolved Architecture Decisions

- **Template storage**: Seeded into the database on first startup via a migration/seed step. Stored in a dedicated `Template` table. Workouts created from templates reference the original via `template_id`.
- **Session management**: JWT (stateless). Backend signs JWTs on login; frontend sends them as Bearer tokens. No server-side session store. Token expiry is configurable.
- **Chart rendering**: SVG with React event handlers. Simpler to build, debug, and integrate with React's event system. Native DOM events for drag-to-resize interactions without a custom hit-testing layer.
- **API format**: REST. Straightforward for the CRUD-heavy API surface; no need for GraphQL complexity at this scale.

## Open Architecture Decisions

1. **FIT file parsing**: Use an existing Go library (e.g., `github.com/tormoder/fit`) or write a minimal parser?
2. **JWT signing**: HS256 with a shared secret (simple) vs. RS256 with key pair (more secure, supports key rotation)?
3. **SVG library**: Custom React SVG components vs. a library like Recharts or Visx for the chart foundation?
