# Logistics

This project has a Go backend, a Nuxt frontend, and a PostgreSQL database.

## Prerequisites

- Go 1.25.11 or newer
- Node.js and npm
- Docker with the Docker Compose plugin
- GNU Make

## Start the full project

From the project root, install the frontend dependencies and start the database,
migrations, backend, and frontend:

```sh
cd frontend && npm ci
cd ..
make s_up
```

When startup completes, open:

- Frontend: http://localhost:3000
- Backend health check: http://localhost:8080/health
- Example API endpoint: http://localhost:8080/api/hello
- Swagger API docs: http://localhost:8080/swagger/index.html

The Go server connects to the local database using
`postgres://app:app@localhost:5432/appdb?sslmode=disable` by default. The
database runs in Docker Compose and keeps its data in a named volume.

Stop the services with:

```sh
make s_down
```

Backend and frontend logs are written to `backend/.run/be.log` and
`backend/.run/fe.log`.

## Run the servers separately

Install frontend dependencies as above. Start PostgreSQL and apply database
migrations:

```sh
make db_up m_up
```

Then start each server in its own terminal from the project root:

```sh
make run-be
```

```sh
make run-fe
```

Stop PostgreSQL when finished with `make db_down`. The database
volume is retained.

## Configuration

The backend accepts `BE_PORT` (default `8080`) and `DATABASE_URL`. For example:

```sh
make run-be BE_PORT=8081 DATABASE_URL='postgres://app:app@localhost:5432/appdb?sslmode=disable'
```

The frontend API base defaults to `http://localhost:8080`; override it with
`NUXT_PUBLIC_API_BASE` when running the frontend against a backend at another
URL.

There is also `STAFF_ID_PREFIX` (default `GAL`), the prefix used for
auto-generated staff IDs, e.g. `GAL-001`, `GAL-002`.

## Progressive Web App (PWA)

The frontend is installable: it ships a web app manifest, app icons and a
service worker that precaches the app shell, so an installed copy opens even
without a network connection (API calls still need the network).

Everything is wired through the [`@vite-pwa/nuxt`](https://vite-pwa-org.netlify.app/frameworks/nuxt)
module:

- `frontend/nuxt.config.ts` — the `pwa` block: manifest (name, colours, icons),
  Workbox precache settings and `registerType: 'autoUpdate'`, so a new build is
  picked up automatically.
- `frontend/public/` — `pwa-192x192.png`, `pwa-512x512.png`,
  `maskable-512x512.png`, `apple-touch-icon.png` and `favicon.svg`. Replace
  these files to change the installed icon; keep the same names and sizes
  (192×192, 512×512, 512×512 maskable, 180×180) or update the `pwa.manifest.icons`
  entries to match.

### Enable / test it

The service worker and manifest are generated at build time only — `nuxt dev`
serves no service worker by design. To test locally:

```sh
make build          # nuxt generate + go:embed + go build
backend/bin/server  # serve the built app
```

Then open http://localhost:8080 and check Chrome DevTools → **Application** →
**Manifest** and **Service workers**. The install icon appears in the address
bar when the manifest is valid; install from there (or the browser's
“Install app” menu) to run it in a standalone window.

To iterate on the service worker without a full Go build, generate the static
output and serve it:

```sh
cd frontend && npx nuxt generate
npx serve .output/public
```

To exercise the PWA in the dev server instead, set
`pwa.devOptions.enabled: true` in `frontend/nuxt.config.ts`. Turn it back off
for normal development: a dev service worker caches assets and can mask HMR
updates.

### Notes

- Service workers only register over HTTPS, or on `http://localhost` for local
testing. Deploy the built app behind TLS.
- The service worker must be reachable at `/sw.js` and the manifest at
  `/manifest.webmanifest`. The embedded Go server sets the manifest's
  `application/manifest+json` content type explicitly; browsers ignore a
  manifest served as a generic binary type.
- Pages and static assets are precached; API responses are never cached, so
  screens show their normal error states when offline.
- Update the app by deploying a new build: `autoUpdate` activates the new
  service worker on the next load without asking the user.

## API Documentation (Swagger)

Interactive Swagger UI documentation is available at:

- `http://localhost:8080/swagger/index.html` (or `http://localhost:8080/swagger`)
- OpenAPI JSON specification: `http://localhost:8080/swagger/doc.json`

To regenerate Swagger documentation after adding or updating endpoints and annotations:

```sh
make swagger
```

## Authentication & Staff Management

The backend uses JWT tokens for authentication, sqlc for type-safe database queries, and Redis for token revocation (with in-memory fallback).

### Users, Roles & Access Control

There are three kinds of users, plus a system administrator seeded from env:

- **System Administrator** (`ADMIN_USERNAME` and `ADMIN_PASSWORD` in `.env`):
  - Logs in via `POST /api/auth/admin/login` or `POST /api/auth/login`.
  - Creates `owner` accounts via `POST /api/owners`.
- **Owner**:
  - Created by the System Admin (`POST /api/owners`).
  - Can CRUD managers (`POST/GET /api/managers`,
    `PUT/DELETE /api/managers/:id`) and branches
    (`POST/PUT/DELETE /api/branches`), and can manage any user or branch.
  - Can also create staff members in any branch (`POST /api/staff`).
- **Manager**:
  - Created by the Owner (`POST /api/managers`).
  - Can CRUD staff details in **their own branch only**
    (`PUT/DELETE /api/staff/:id`).
  - Creates staff with `POST /api/staff` (job_role required); operations are
    limited to their assigned branch.
  - Can act in any staff job role (operations, dispatch, receiving).
- **Staff**:
  - Created by a Manager; must be given a `job_role`: `operations`, `dispatch`
    or `receiving`.
  - Cannot manage staff, managers or branches.

#### Staff IDs, PINs and first login

- Every user (owner, manager, staff) gets an auto-generated `staff_id` (e.g.
  `GAL-001`; prefix is configurable via `STAFF_ID_PREFIX`) at creation time.
- A one-time password is returned once, in the creation response, and given to
  the user.
- On first login, the user authenticates with `staff_id` + one-time password
  **and** a `new_pin` (4–10 chars). That PIN becomes their permanent credential
  and `first_login_done` flips to true.
- `GET /api/staff/:id/first-login` reports whether a user has performed their
  first login and when.
- A manager (or owner) can force a PIN reset with
  `POST /api/staff/:id/reset-pin`: the system generates a new one-time password
  (returned once), the old PIN is invalidated, and the user chooses a new PIN
  at their next login.

#### Login & lockout

- `POST /api/auth/login` with `staff_id` and `pin` (and `new_pin` when a
  one-time password is pending).
- The account **locks after 3 failed login attempts**. A locked staff account
  can be reopened by the staff's branch manager:
  `POST /api/staff/:id/unlock`. The owner can unlock accounts across all
  branches.

### sqlc Code Generation

To regenerate Go code from SQL queries (`backend/internal/db/queries`) and schema:

```sh
make sqlc
```