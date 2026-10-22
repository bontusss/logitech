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
make -C backend s_up
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
make -C backend s_down
```

Backend and frontend logs are written to `backend/.run/be.log` and
`backend/.run/fe.log`.

## Run the servers separately

Install frontend dependencies as above. Start PostgreSQL and apply database
migrations:

```sh
make -C backend db_up m_up
```

Then start each server in its own terminal from the project root:

```sh
make -C backend run-be
```

```sh
make -C backend run-fe
```

Stop PostgreSQL when finished with `make -C backend db_down`. The database
volume is retained.

## Configuration

The backend accepts `BE_PORT` (default `8080`) and `DATABASE_URL`. For example:

```sh
make -C backend run-be BE_PORT=8081 DATABASE_URL='postgres://app:app@localhost:5432/appdb?sslmode=disable'
```

The frontend API base defaults to `http://localhost:8080`; override it with
`NUXT_PUBLIC_API_BASE` when running the frontend against a backend at another
URL.

There is also `STAFF_ID_PREFIX` (default `GAL`), the prefix used for
auto-generated staff IDs, e.g. `GAL-001`, `GAL-002`.

## API Documentation (Swagger)

Interactive Swagger UI documentation is available at:

- `http://localhost:8080/swagger/index.html` (or `http://localhost:8080/swagger`)
- OpenAPI JSON specification: `http://localhost:8080/swagger/doc.json`

To regenerate Swagger documentation after adding or updating endpoints and annotations:

```sh
make -C backend swagger
```

## Authentication & Staff Management

The backend uses JWT tokens for authentication, sqlc for type-safe database queries, and Redis for token revocation (with in-memory fallback).

### Users, Roles & Access Control

There are three kinds of users, plus a system administrator seeded from env:

- **System Administrator** (`ADMIN_USERNAME` and `ADMIN_PASSWORD` in `.env`):
  - Logs in via `POST /api/auth/admin/login` or `POST /api/auth/login`.
  - Can only create `owner` accounts via `POST /api/staff`.
- **Owner**:
  - Created by the System Admin.
  - Can CRUD managers (`POST/PUT/DELETE /api/staff`) and branches
    (`POST/PUT/DELETE /api/branches`), and can manage any user or branch.
- **Manager**:
  - Created by the Owner.
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

- Every user gets an auto-generated `staff_id` (e.g. `GAL-001`; prefix is
  configurable via `STAFF_ID_PREFIX`) at creation time.
- A one-time password is returned once, in the creation response, and given to
  the user.
- On first login, the user authenticates with `staff_id` + one-time password
  **and** a `new_pin` (4–10 chars). That PIN becomes their permanent credential
  when `first_login_done` flips to true.
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
make -C backend sqlc
```