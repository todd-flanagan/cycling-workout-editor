# Cycling Workout Editor — Requirements

## Overview

A web-based application for creating, managing, and exporting structured cycling workouts. Each user maintains a personal library of workouts anchored to their Functional Threshold Power (FTP).

---

## R1 — Multi-User Support

- The system must support multiple independent users.
- Each user has their own workout library, profile, and settings.
- User data is isolated; one user cannot view or modify another user's workouts.

## R2 — Authentication via Google Login

- Users authenticate using Google OAuth 2.0 / OpenID Connect.
- No local username/password registration; Google is the sole identity provider at launch.
- On first login, a user account is automatically provisioned.
- Session management must support persistent login (remember me) with configurable expiry.

## R3 — Workout Library

- Each user maintains a personal library of workouts.
- A workout consists of:
  - **Name** — short descriptive title.
  - **Textual description** — free-form notes (purpose, tips, context).
  - **Visual representation** — a graphical power/time chart rendered in the UI.
  - **Structured interval data** — the machine-readable workout definition (intervals, ramps, rest periods, etc.).
- Users can create, edit, duplicate, and delete workouts.
- Workouts should be searchable and sortable within the library.

## R4 — FTP-Anchored Workouts

- All workout intensities are defined relative to the user's FTP (e.g., 75% FTP, 120% FTP).
- Each user sets their FTP in their profile.
- When FTP changes, all workouts automatically reflect the new absolute wattage values without requiring edits.

## R5 — Export

- **Zwift `.zwo`**: Users can export any workout to Zwift's `.zwo` XML format. Exported files must be valid and importable by the Zwift application. Export preserves workout name, description, and all interval structure.
- **Text file**: Users can export a workout as a human-readable text file for sharing (copy/paste, email, etc.). This is the initial sharing mechanism — no in-app social features at launch.

## R6 — Import

- Users can import workouts from common formats:
  - Zwift `.zwo` files.
  - Garmin `.fit` files.
- Imported workouts are parsed, converted to the internal interval model, and added to the user's library.
- Import should handle graceful errors (invalid files, unsupported features) with clear user feedback.

## R7 — Workout Templates

- The app ships with a built-in library of common workout templates that users can add to their library.
- Templates include (at minimum):
  - **FTP Ramp Test** — progressive ramp to exhaustion for FTP estimation.
  - **Sweet Spot** — sustained efforts at 88–94% FTP.
  - **VO2max Intervals** — high-intensity repeats at 106–120% FTP.
  - **Endurance Ride** — steady Z2 effort (55–75% FTP).
  - **Threshold Intervals** — sustained efforts at 95–105% FTP.
  - **Over-Unders** — alternating above/below threshold.
  - **Recovery Ride** — easy Z1 spin (< 55% FTP).
- Templates are read-only; adding one to a library creates an editable copy.

## R8 — Workout Tags

- Users can assign one or more tags to each workout (e.g., "VO2max", "recovery", "race prep", "indoor", "outdoor").
- Tags are user-defined (free-form text).
- The workout library can be filtered by tag.
- Template workouts come pre-tagged.

## R9 — FTP Tracking

- The app stores the user's current FTP and maintains a history of FTP changes over time.
- FTP history is displayed as a simple chart or table in the user profile.
- Each FTP entry records the value, date, and optionally a source/note (e.g., "ramp test", "20-min test", "estimated").
- **Two methods to set FTP**:
  1. **Direct entry** — user enters a known FTP value.
  2. **Ramp test calculation** — user enters their max 1-minute power from a ramp test; the app calculates FTP as 75% of that value. User confirms before saving.
- When FTP is updated, the new value is recorded in the history and becomes the current FTP.

## R11 — Editor Undo/Redo and Keyboard Shortcuts

- The workout editor supports a full undo/redo stack for all interval modifications.
- Keyboard shortcuts for common operations:
  - **Delete/Backspace**: remove selected interval.
  - **Arrow Up/Down**: nudge selected interval power +/- 1% FTP.
  - **Arrow Left/Right**: nudge selected interval duration +/- 5 seconds.
  - **Ctrl+Z**: undo.
  - **Ctrl+Shift+Z**: redo.

## R10 — Training Plans (Future)

- The system will eventually support organizing workouts into multi-day/multi-week training plans with a calendar view.
- This is out of scope for the initial release but should be considered in data model and architecture decisions.

## R12 — "Train My Commute" (Future Vision)

Over time, evolve this application into a **"Train My Commute"** platform. The core idea: many cyclists get a significant portion of their weekly volume from a daily bike commute. Traditional training plans assume all ride time is discretionary, but commute riders have fixed ride slots (to/from work) with real-world constraints — fixed route, traffic, time pressure, weather.

**Concept:**

- Users import or define a **structured training plan** (e.g., a base-build-peak plan, polarized block, sweet spot progression).
- Users describe their **commute profile**: distance, typical duration, elevation, frequency (days/week), and any constraints (e.g., "must arrive by 8:30 AM", "no showering at work so cap intensity").
- The system **adapts the training plan to overlay onto commute rides** — turning commute segments into interval sessions where appropriate, scheduling recovery commutes on easy days, and identifying which plan workouts genuinely need a separate dedicated ride.
- The plan accounts for **cumulative fatigue** from commuting volume that traditional plans wouldn't include.
- Weekly summaries show how commute rides satisfy plan objectives vs. what remains for dedicated training sessions.

**Key design considerations for R12:**

- The commute ride is the atomic building block, not just another calendar slot.
- Plan adaptation should respect real-world commute constraints (arrival time, intensity caps, gear/logistics).
- Support for asymmetric commutes (e.g., hillier one direction, different times of day).
- Integration with R5/R6 exports so adapted commute workouts can be pushed to head units.

## R13 — Platform Integrations (Future — supports R12)

To close the loop between planned training and actual riding, the system needs bidirectional integration with fitness platforms.

**Inbound — Ride Data Collection:**

- **Strava API**: Pull completed ride activities (power, HR, duration, route, TSS/IF) to track what the rider actually did — especially commute rides.
- **Garmin Connect API**: Pull ride data from Garmin devices as an alternative or complement to Strava.
- Automatic classification of imported rides as "commute", "dedicated training", or "other" based on route matching, time-of-day, and user-defined rules.
- Use actual ride data to assess plan compliance — did the commute ride hit the prescribed intervals? How does cumulative load compare to the plan's intent?

**Outbound — Publish Training Activities:**

- **Garmin Connect API**: Push scheduled workouts directly to the user's Garmin device via Garmin Connect. The rider sees the day's prescribed workout (with target power zones and intervals) on their head unit before rolling out.
- Support pushing both commute-adapted workouts and standalone training sessions.
- Sync the training calendar so upcoming workouts appear on the Garmin device automatically.

**Key design considerations for R13:**

- OAuth 2.0 flows for both Strava and Garmin Connect, with token refresh and revocation handling.
- Polling/webhook strategy for ingesting new activities (Strava supports webhooks; Garmin Connect uses push via its API).
- Rate limiting and data caching to stay within API quotas.
- Privacy: users control which activities are imported and what data is retained.
- Graceful degradation when a platform connection is unavailable or revoked.
- **Future expansion**: extend platform support beyond Garmin to other popular cycling head units and ecosystems, including **Wahoo (ELEMNT series)**, **Hammerhead (Karoo)**, **Bryton**, and **Stages**. Each platform has its own API, workout format, and sync mechanism — the integration layer should be designed with a provider abstraction so adding new platforms is incremental rather than architectural. Wahoo's cloud API and Hammerhead's open Karoo platform are natural next targets after Garmin.

---

## Non-Functional Requirements

### NF1 — Technology Stack

| Layer    | Technology  |
|----------|------------|
| Frontend | React (JavaScript) |
| Backend  | Go         |
| Database | SQLite     |

### NF2 — Browser Support

- Desktop and tablet browsers are the primary targets: Chrome, Firefox, Edge, Safari.
- Mobile browser support is a future goal (viewing workouts, not editing). React is chosen in part to facilitate this path.

### NF3 — Deployment

- Self-hosted as the initial deployment target.
- Each layer (frontend, backend) is containerized as part of the build process.
- Docker Compose or similar for local/self-hosted orchestration.

### NF4 — Data Persistence

- SQLite as the database, stored as a file on the host.
- Volume-mounted in the backend container for persistence across restarts.

---

## Open Questions

1. **Offline support** — Any need for offline/PWA capability?
2. **Collaboration** — Should multiple users be able to co-edit a workout (e.g., coach and athlete)?
3. **Additional export formats** — Beyond `.zwo` and text, should we add TrainerRoad or MRC/ERG formats?
4. **Workout versioning** — Should edits create a new version, or just overwrite?
