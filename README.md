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

## API Documentation (Swagger)

Interactive Swagger UI documentation is available at:

- `http://localhost:8080/swagger/index.html` (or `http://localhost:8080/swagger`)
- OpenAPI JSON specification: `http://localhost:8080/swagger/doc.json`

To regenerate Swagger documentation after adding or updating endpoints and annotations:

```sh
make -C backend swagger
```