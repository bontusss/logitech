# BOL — based on logistics

Multi-tenant logistics platform: a public marketplace landing page, a company operations
workspace, and the backend foundation for authentication, company onboarding and company
management.

The platform implements **authentication, onboarding and company management**: profiles, branches,
staff invitations and acceptance, granular roles and permissions, branding, subdomains and
company-branded PWA metadata. Providers can create drivers with private, passwordless portal links
and record deliveries with contacts, addresses, declared value, price, payment details and optional
driver assignment. Live driver GPS, delivery progress updates, proof of delivery, dispatch-board
workflows, manifests, the rate engine and the public marketplace API are not built yet.

## Stack

| Layer | Technology |
| --- | --- |
| Frontend | Nuxt 4, Vue 3 Composition API, TypeScript, Tailwind CSS v4, Lucide |
| Backend | Go, Gin, pgx, sqlc, golang-migrate |
| Data | PostgreSQL 16, Redis 7 |
| Local infra | Docker Compose |

```
.
├── frontend/                  # Nuxt landing page + auth/company workspace
│   ├── app/components         # landing sections, workspace forms and dialogs
│   ├── app/composables        # useApi, useAuth, useCompany, useWorkspace
│   ├── app/layouts            # default, auth, app
│   ├── app/pages              # /, auth, onboarding, dashboard/company…
│   ├── app/plugins            # client branding/PWA metadata
│   └── app/middleware         # authenticated workspace guard
├── backend/
│   ├── cmd/server             # Gin API
│   ├── cmd/migrate            # golang-migrate wrapper (make m_up/m_down)
│   ├── internal/auth          # Argon2id, access JWT, opaque tokens
│   ├── internal/config        # environment configuration
│   ├── internal/db/queries    # sqlc source queries
│   ├── internal/db/sqlc       # generated, type-safe query layer
│   ├── internal/httpapi       # routes, middleware, handlers
│   ├── internal/service       # sessions, registration, invitation acceptance
│   ├── internal/ratelimit     # Redis limiter with in-process fallback
│   ├── internal/mailer        # mail boundary (log implementation in dev)
│   └── migrations             # golang-migrate up/down migrations
└── Makefile
```

## Local setup

Requirements: Go 1.26+, Node 20+, Docker.

```bash
cp .env.example .env

make s_up          # start Postgres (5432) and Redis (6380)
make m_up          # apply migrations
make run-be        # API on http://localhost:8080
make run-fe        # Nuxt on http://localhost:3000
```

`make run-fe` loads the repository-root `.env`, including `NUXT_PUBLIC_API_BASE`, so
frontend requests go to the local API at `http://localhost:8080` during development.

Health check: `curl http://localhost:8080/health`.

Create the integration-test database once:

```bash
docker exec app_postgres createdb -U app appdb_test
```

## Environment

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis URL (host port **6380** by default to avoid collisions) |
| `APP_ENV` | Set to `production` for production deployments |
| `JWT_SECRET` | HS256 signing key, minimum 32 characters; required in production |
| `ACCESS_TOKEN_TTL` | Access token lifetime (default `15m`) |
| `REFRESH_TOKEN_TTL` | Refresh family lifetime (default `720h`) |
| `APP_URL` | Frontend origin used in email links and CORS defaults |
| `CORS_ORIGINS` | Comma-separated allowed browser origins |
| `COOKIE_SECURE` | Set true behind HTTPS (forced in production) |
| `SUPER_ADMIN_EMAIL` | Bootstrap platform super administrator email; created at startup if absent |
| `SUPER_ADMIN_PASSWORD` | Bootstrap super administrator password; change before deploying |
| `PLUNK_SECRET_KEY` | Server-side Plunk secret key used for transactional email |
| `PLUNK_FROM_EMAIL` | Sender address on a verified Plunk sending domain |
| `PLUNK_PUBLIC_KEY` | Optional; only needed for browser-side Plunk event tracking |
| `WHATSAPP_ACCESS_TOKEN` | Meta WhatsApp Cloud API access token |
| `WHATSAPP_PHONE_NUMBER_ID` | Meta WhatsApp business phone-number ID |
| `WHATSAPP_GRAPH_API_VERSION` | Graph API version (default `v23.0`) |
| `WHATSAPP_OTP_TEMPLATE` | Approved WhatsApp authentication template name |
| `WHATSAPP_TEMPLATE_LANGUAGE` | Template language code (default `en_US`) |
| `UPLOAD_DIR` | Local directory for logo/cover uploads (default `uploads`) |
| `API_PUBLIC_URL` | Public API origin used to build uploaded image URLs |
| `APP_DOMAIN` | Tenant base domain used for host/subdomain resolution |
| `NUXT_PUBLIC_API_BASE` | API base used by the Nuxt client |
| `NUXT_PUBLIC_APP_DOMAIN` | Frontend copy of the tenant base domain |
| `NUXT_PUBLIC_DEMO_URL` | Booking CTA destination |
| `NUXT_PUBLIC_WHATSAPP_URL` | Full `https://wa.me/<number>` chat URL |
| `NUXT_PUBLIC_MAPBOX_ACCESS_TOKEN` | Public Mapbox token for delivery address lookup and map previews |
| `NUXT_PUBLIC_CLOUDINARY_CLOUD_NAME` | Cloudinary cloud name for optional booking-page package photos |
| `NUXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET` | Cloudinary unsigned upload preset for package photos |

Public booking pages use Mapbox autocomplete/reverse geocoding for verified location pins.
Optional package photos use Cloudinary's unsigned image upload; configure the preset to allow
images only and cap file size (the booking form accepts files up to 10 MB). If Cloudinary is not
configured, customers can still submit bookings without a photo.

`JWT_SECRET` falls back to a documented development value only when
`APP_ENV != production`; production startup fails without it.

## Container deployment

The root `Dockerfile` generates the Nuxt site, embeds its static output in the Go server, and
serves both the frontend and `/api/v1` from one web service. Deploy it as a Render Web Service
using the repository's Dockerfile; the server listens on Render's `PORT` environment variable.
For this same-origin deployment, set `APP_ENV=production`, `APP_URL` to the public HTTPS
frontend URL, and `CORS_ORIGINS` to that URL (without a trailing slash). The frontend defaults
to the relative `/api/v1` path. If the API is hosted on a different origin, set
`NUXT_PUBLIC_API_BASE` to its public API URL in the frontend build environment; this value is embedded when the static
frontend is generated, so changing only the runtime environment requires a rebuild/redeploy.
The Mapbox token is fetched from the API at runtime, so configure
`NUXT_PUBLIC_MAPBOX_ACCESS_TOKEN` on the API service; changing it does not require rebuilding the
static frontend.

### Transactional email and WhatsApp verification

The API sends verification, password-reset, invitation, welcome and provider-onboarding
emails through Plunk, using the sender name **Nnanna from BOL**. Configure a Plunk secret
key and a sender email whose domain is verified in Plunk. Production startup requires
`PLUNK_SECRET_KEY` and `PLUNK_FROM_EMAIL`.

Welcome emails are queued for 30 minutes after registration, with separate customer/shipper
and provider content. Provider onboarding reminders are queued for 24 and 72 hours after
registration; they are cancelled after the provider submits the profile and skipped if the
company is no longer incomplete. The queue is stored in PostgreSQL and processed by the API.

WhatsApp verification uses Meta's WhatsApp Cloud API directly. Create an approved
authentication message template named to match `WHATSAPP_OTP_TEMPLATE`, with one body
parameter for the six-digit code (for example, `Your BOL verification code is {{1}}. It expires in 10 minutes.`).
Configure its language, a token with
`whatsapp_business_messaging` permission, and the business phone-number ID. The provider
enters their phone in international E.164 format during onboarding. Codes expire after
10 minutes, can be resent after one minute, and are limited to five per hour. Meta offers a
test number for development, but production authentication messages may be billable; check
Meta's current country-specific pricing and messaging requirements.

## Authentication flow

1. **Registration**
   - `POST /auth/register` creates a marketplace customer.
   - `POST /companies` creates the owner user, company, membership, settings and branding in a
     single PostgreSQL transaction.
2. **Passwords** are hashed with Argon2id (64 MiB, 3 iterations, 2 lanes) and stored as PHC strings.
3. **Access tokens** are short-lived HS256 JWTs with `sub`, `jti`, `iss`, `iat` and `exp`, returned
   in the response and set as an HttpOnly `bol_access` cookie.
4. **Refresh tokens** are 256-bit opaque values stored only as SHA-256 hashes. Rotation revokes the
   previous token; reuse revokes the whole family.
5. **Logout** revokes the refresh family, denylists the access-token `jti` in Redis for its
   remaining lifetime, and clears both cookies.
6. **Email verification / password reset** use single-use opaque tokens stored as hashes.
   Plunk delivers verification links and reset links; only development exposes verification tokens
   in the API response.
7. **Invitation tokens** are 256-bit opaque values stored hashed, expire after seven days and are
   claimed atomically so a link can only be accepted once.

## Tenant isolation

- Every company-owned row carries `company_id`.
- The client may select a company with `X-Company-Id`, but the middleware resolves it through the
  authenticated user's **active `company_members` row**. A supplied id alone grants nothing.
- Queries are scoped by `company_id` in SQL; handlers never accept a tenant id as proof of access.
- Branch managers see only branches in `branch_members`. Owners and company administrators see all
  company branches.
- Platform administrators hold an explicit `platform_role` (`super_admin`, `admin`, `support`).
  Support accounts are read-only for company mutations.
- Removing or suspending a member revokes their refresh tokens and branch assignments.
- The owner cannot be removed, demoted or suspended through company endpoints.

## Database schema

```
users ─────┐
           │ owner_user_id
           ▼
        companies ──1:1── company_settings
           │        ├──1:1── company_branding
           │        └──1:1── company_subdomains
           │
           ├──< branches ──< branch_members >── company_members
           │
           ├──< company_members >── users
           │         │
           │         └──< member_roles >── roles ──< role_permissions >── permissions
           │
           └──< company_invitations ──< invitation_branches >── branches

users ──< email_verifications
      ──< password_resets
      ──< refresh_tokens
      ──< audit_logs
```

- **companies** — profile, description, cover image, operating hours, verification and onboarding
  status, owner.
- **company_branding** — logo, favicon, cover, colours, public name/description and PWA short name.
- **company_subdomains** — one unique, reserved-word-checked subdomain per company.
- **branches** — per-company unique code, address, contact, manager, hours and active/inactive
  status. Branches are never deleted through the API; they are deactivated.
- **branch_members** — staff-to-branch assignments used for branch-manager access.
- **company_invitations** — hashed token, role, expiry, revocation and acceptance metadata.
- **invitation_branches** — branch access granted when an invitation is accepted.
- **roles / permissions** — system templates plus company-scoped custom roles, with granular keys.

Primary keys are UUIDs. All mutable tables carry `created_at`/`updated_at`; tenant tables use
soft deletion where history matters.

## API (v1)

All routes are under `/api/v1`. Authentication accepts `Authorization: Bearer <access_token>` or
the HttpOnly access cookie. Company routes accept `X-Company-Id`.

### Public

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/auth/register` | customer registration, rate limited |
| POST | `/auth/login` | sign in, rate limited |
| POST | `/auth/refresh` | rotate refresh token |
| POST | `/auth/logout` | revoke session |
| POST | `/auth/forgot-password` | request reset link |
| POST | `/auth/reset-password` | reset password |
| POST | `/auth/verify-email` | verify email |
| POST | `/companies` | company onboarding |
| GET | `/invitations/:token` | invitation preview |
| POST | `/invitations/:token/accept` | accept invitation (session optional) |
| GET | `/public/branding` | host/query resolved public branding |
| GET | `/public/companies/:subdomain` | public company profile |
| GET | `/public/manifest` | company-branded PWA manifest |

### Platform administration

Platform routes require a signed-in account whose `account_type` is `platform`. `admin` and
`super_admin` can approve providers; `support` is read-only. Creating admin accounts is
restricted to `super_admin`.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/admin/overview` | customers, provider companies and admin accounts |
| POST | `/admin/providers/:id/approve` | verify a provider company |
| POST | `/admin/admins` | create an `admin` or `support` account (super admin only) |

### Company workspace

| Method | Path | Permission |
| --- | --- | --- |
| GET | `/companies/me` | active member |
| PATCH | `/companies/me` | `company.update` |
| GET / PATCH | `/companies/me/settings` | member / `company.settings.manage` |
| PATCH | `/companies/me/branding` | `branding.manage` |
| POST | `/companies/me/branding/upload` | `branding.manage` |
| POST | `/companies/me/submit-verification` | `company.update` + verified owner email |
| GET / PUT | `/companies/me/subdomain` | `subdomain.manage` |
| GET | `/companies/me/subdomain/availability` | `subdomain.manage` |
| GET | `/companies/me/branches` | `branches.view` (access-filtered) |
| POST | `/companies/me/branches` | `branches.create` |
| GET / PATCH | `/companies/me/branches/:id` | `branches.view` / `branches.update` |
| POST | `/companies/me/branches/:id/activate` | `branches.deactivate` |
| POST | `/companies/me/branches/:id/deactivate` | `branches.deactivate` |
| GET / PUT | `/companies/me/branches/:id/staff` | `branches.view` / `staff.update` |
| GET | `/companies/me/members` | `staff.view` |
| PATCH | `/companies/me/members/:id` | `staff.update` |
| PUT | `/companies/me/members/:id/branches` | `staff.update` |
| PATCH | `/companies/me/members/:id/role` | `staff.roles.manage` |
| DELETE | `/companies/me/members/:id` | `staff.deactivate` |
| GET / POST | `/companies/me/invitations` | `staff.view` / `staff.invite` |
| DELETE | `/companies/me/invitations/:id` | `staff.invite` |
| POST | `/companies/me/invitations/:id/resend` | `staff.invite` |
| GET / POST | `/companies/me/roles` | `staff.roles.manage` |
| PATCH / DELETE | `/companies/me/roles/:id` | `staff.roles.manage` |

Inactive branches remain available for historical records but are excluded from new assignments.
There is deliberately no branch delete endpoint.

## Permissions

| Role | Capabilities |
| --- | --- |
| `owner` | All company permissions, including ownership; cannot be delegated |
| `company_admin` | Profile, settings, branches, staff, roles, branding and subdomain management |
| `branch_manager` | View/update assigned branches, view staff, shipments and reports |
| `operations_staff` | Company/branch read access and shipment operations |
| `customer_service` | Company/branch/staff read access, shipments, rates and reports |
| Custom roles | Any subset of the actor's own permissions, managed by `staff.roles.manage` |

Granular keys include `company.*`, `branches.*`, `staff.*`, `shipments.*`, `rates.*`,
`reports.*`, `branding.manage` and `subdomain.manage`. Legacy Task 1 permission keys remain in the
database for compatibility; new endpoints use the granular keys.

## Frontend

- `/login`, `/register` (customer or logistics company), `/forgot-password`, `/verify-email`
- `/admin/login` — platform administrator sign-in
- `/admin` — customers, provider approvals and admin account creation
- `/onboarding` — company profile, contact, location, branding and CAC business name
- `/dashboard` — account/company status and onboarding checklist
- `/dashboard/company` — profile, operating hours and verification
- `/dashboard/company/branches` — search/filter, create, activate/deactivate
- `/dashboard/company/branches/[id]` — details, edit and staff assignment
- `/dashboard/company/staff` — directory, invitations, role/branch/status updates
- `/dashboard/company/roles` — system roles and custom permission editor
- `/dashboard/company/branding` — brand identity, image uploads and subdomain
- `/dashboard/company/settings` — timezone, currency, weight unit and notifications
- `/invitations/[token]` — public invitation preview and acceptance

The workspace is client-authenticated against the API using HttpOnly cookies; no tokens are stored
in `localStorage`.

### Platform administration

Sign in at **`/admin/login`**. On API startup the server creates the super administrator from
`SUPER_ADMIN_EMAIL` / `SUPER_ADMIN_PASSWORD` (see `.env.example`) if that email does not exist yet;
an existing account is never overwritten. The dashboard lists customers and provider companies,
lets an admin approve a provider, and lets the super admin create additional `admin` or
`support` accounts. Success, error, loading and empty states are shared through `useWorkspace`,
`ToastHost` and `ConfirmDialog`.

### Branding and PWA

Companies get a unique subdomain (`company_subdomains`). The backend resolves the subdomain from
the request Host (for public branding/manifest) and the Nuxt client plugin applies the company's
colours, title, favicon and manifest link when the app is loaded on a company subdomain. The
manifest endpoint returns the dynamic app name, short name, icon, theme colour and background
colour, so a single frontend deployment serves every company.

## Tests

```bash
# Unit tests (Argon2id and JWT)
cd backend && go test ./internal/auth

# Full suite, including integration tests against appdb_test
TEST_DATABASE_URL='postgres://app:app@localhost:5432/appdb_test?sslmode=disable' go test ./...

# Frontend production build
cd frontend && npx nuxt generate
```

Integration coverage includes:

- company registration, tenant isolation and refresh-token rotation;
- profile, operating hours and branding updates, including image upload/serving;
- branch creation, duplicate codes, manager-only branch access, staff assignment and deactivation;
- invitation preview, sign-in-required acceptance, new-account acceptance, single use, expiry,
  resend and revocation;
- suspension immediately removing company access;
- custom role creation/update/deletion, authority checks and owner-role protection;
- subdomain validation, reserved words, uniqueness and host-based public resolution;
- manifest content type and platform fallback.

## Known limitations and deferred work

- Row-level security is not enabled. Tenant isolation is enforced in SQL plus membership
  middleware; RLS is a planned defence-in-depth addition.
- Uploaded images are stored on the local filesystem (`UPLOAD_DIR`). A storage provider and image
  processing pipeline are not configured.
- Subdomain DNS and wildcard TLS are deployment concerns and are not provisioned by the app.
- Provider verification collects the registered CAC business name; document uploads and automated
  CAC lookups are not implemented, so an administrator checks the name manually before approving.
- Shipments, rate engine, manifests and marketplace booking are out of scope for this stage.

## Recommended next task

Build the shipment and rate engine on top of these branches and permissions, using the existing
tenant middleware, branch access rules and audit trail.
