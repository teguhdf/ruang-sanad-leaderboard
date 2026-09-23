# Laju Go

High-performance SaaS boilerplate built with **Go Fiber** + **Svelte 5** + **Inertia.js 3** + **SQLite** (CGO).

Build production-ready web applications faster with clean layered architecture that combines the speed of Go with the DX of modern frontend frameworks. Ships with **Svelte 5** by default, but Inertia.js makes it trivial to swap to **React** or **Vue** without changing any Go code.

## 🚀 Quick Start

The fastest way to start a new project is via the published CLI — it scaffolds the project, installs Go + Node dependencies, and configures `.env` automatically:

```bash
npx create-laju-go my-app
cd my-app
npm run dev:all
```

> `create-laju-go` clones the latest template from this repo, replaces placeholders (`laju-go` → `my-app`), runs `go mod download` + package manager install, and copies `.env.example` to `.env`. See <https://www.npmjs.com/package/create-laju-go>.

Alternatively, clone directly (manual setup):

```bash
git clone https://github.com/maulanashalihin/laju-go.git
cd laju-go
make setup
npm run dev:all
```

Visit `http://localhost:8080` to see your application running.

## ✨ Features

### File Upload (TUS Protocol)

- **TUS Resumable Upload** — Powered by [tusdfiber](https://github.com/maulanashalihin/tusdfiber), native Go Fiber implementation
- **Drag & Drop** — Test page at `/app/upload` with progress bars via `tus-js-client`
- **Large File Support** — Up to 1GB per upload, chunked via PATCH requests
- **Post-Processing** — Completed files copied to `storage/completed/` for easy access

### Authentication & Security

- **Email/Password** — Argon2id hashing, session management
- **Google OAuth 2.0** — One-click social login
- **Password Reset** — Email-based recovery with secure tokens
- **CSRF Protection** — Double-submit cookie pattern (stateless, no session I/O)
- **Rate Limiting** — Configurable throttling for auth, API, upload endpoints
- **Session Fixation Protection** — Session ID regenerated on privilege escalation

### User Management

- **Role-Based Access Control** — Admin/User roles with middleware guards
- **Profile Management** — Update name, email, avatar
- **Avatar Upload** — With file type/size validation

### Development Experience

- **Hot Module Replacement** — Vite HMR for instant frontend updates
- **Go Hot Reload** — Air rebuilds on Go file changes (~1-2s)
- **Clean Architecture** — Handler → Service → Query (sqlc-generated)
- **Full TypeScript** — Every `.svelte` file uses `<script lang="ts">`
- **Type-Safe Templates** — Go HTML via [templ](https://templ.guide/)

### Performance & Database

- **SQLite (mattn/go-sqlite3)** — CGO-based, 2x throughput vs pure-Go drivers
- **WAL Mode + mmap** — Optimized for production workloads
- **In-Memory Session Cache** — Fast sync.RWMutex+map cache for session data
- **Background Cleanup** — Expired sessions & tokens auto-purged every hour

## 📁 Project Structure

```
laju-go/
├── cmd/laju-go/main.go        # Application entry point
├── app/                       # Backend Go code
│   ├── handlers/              # HTTP request handlers
│   ├── services/              # Business logic layer
│   ├── queries/               # sqlc-generated query code
│   ├── middlewares/           # Auth, CSRF, rate limiting
│   ├── cache/                 # In-memory session cache
│   ├── models/                # Data structures + DTOs
│   ├── session/               # Session store (SQLite + cache)
│   └── config/                # Env-based configuration
├── frontend/                  # Svelte 5 frontend
│   └── src/
│       ├── components/        # Header, DarkModeToggle
│       ├── pages/auth/        # Login, Register, ForgotPassword, ResetPassword
│       ├── pages/app/         # Dashboard, Profile, UploadTest
│       └── lib/i18n/          # Internationalization (en/id)
├── queries/                   # SQL source files (write here → sqlc)
├── routes/                    # Route definitions
├── migrations/                # Goose SQL migrations (1 table per file)
├── templates/                 # templ HTML components
├── prompts/                   # AI agent deploy prompts
└── systemd/                   # Production service file
```

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.26+, Fiber v2 |
| **Database** | SQLite via `mattn/go-sqlite3` (CGO) |
| **Query Builder** | sqlc — compile-time type-safe SQL |
| **Migrations** | Goose — embedded in binary via `go run` |
| **Frontend** | Svelte 5 (rune-based) |
| **Build Tool** | Vite 8 |
| **Styling** | Tailwind CSS 4 |
| **Templating** | templ — type-safe Go HTML |
| **SPA Bridge** | Inertia.js 3 via [fiber-inertia](https://github.com/maulanashalihin/fiber-inertia) |
| **Icons** | Lucide Svelte |

### Why `mattn/go-sqlite3` (CGO)?

The `mattn/go-sqlite3` driver delivers **~1.3–1.9x higher throughput** than pure-Go alternatives in real-world benchmarks (100K+ RPS on a $24/mo Vultr instance). The trade-off:

- ✅ **2x faster** on production workloads
- 🛠️ **Cross-compile** via `make build-linux` (requires `brew install zig` for `zig cc`)
- ➡️ **Static binary** still possible — `zig cc` links `libsqlite3` statically

For development (macOS), CGO works out of the box — no extra setup needed.

## ⚡ Quick Reference

```bash
# Development
npm run dev:all                # Vite + Air (hot reload both)

# Build (production)
npm run build:all              # vite build → go build

# Verify (before commit)
npm run verify                 # templ → vite → go build → go vet → go test

# Build for Linux (from macOS)
make build-linux               # requires zig cc

# Database
npm run db:migrate             # run pending migrations
npm run db:generate            # sqlc — generate Go from SQL

# Templates
templ generate                 # regenerate templ Go files

# Verify (before commit/deploy)
npm run verify                 # templ → vite → go build → go vet

# Test
go test ./...                  # backend tests
```

### Testing Strategy

| Approach | For | Command |
|----------|-----|---------|
| Go unit/integration | Services, queries, handlers | `go test ./...` |
| E2E / user flow | Visual regression, auth flows, form submission | `agent_browser` via pi |

> **E2E testing** is done manually with `agent_browser` (open browser, click, fill forms, verify redirects).
> No Cypress/Playwright needed — a real browser is more realistic for a project of this scale.

## 🚀 Deployment

Three paths: **AI agent** (copy-paste a prompt), **Docker**, or **bare VPS** with systemd. Full guides at <https://laju.dev/deployment/>.

```bash
# Build (vite build → go build)
npm run build:all

# Build for Linux (from macOS, requires zig cc for CGO cross-compile)
make build-linux
```

> **CGO note**: `mattn/go-sqlite3` uses C bindings. Cross-compiling from macOS/Windows to Linux requires a C cross-compiler. `make build-linux` uses `zig cc` — install with `brew install zig`. Produces a static ELF binary.

## 🗄️ Database

Migrations run **automatically** on startup. Manual:

```bash
# Run pending migrations
go run github.com/pressly/goose/v3/cmd/goose@latest \
  -dir migrations sqlite3 ./data/app.db up

# Status
go run github.com/pressly/goose/v3/cmd/goose@latest \
  -dir migrations sqlite3 ./data/app.db status
```

Write SQL in `queries/*.sql`, then:

```bash
npm run db:generate    # sqlc generates Go code into app/queries/
```

## 📖 Documentation

| Section | Description |
|---------|-------------|
| [Architecture](https://laju.dev/architecture/overview/) | Layered design, patterns, conventions |
| [Database](https://laju.dev/database/schema-migrations/) | SQLite setup, migrations, sqlc |
| [Frontend](https://laju.dev/frontend/svelte-inertia/) | Svelte 5 + Inertia.js patterns |
| [Deployment](https://laju.dev/deployment/overview/) | Docker, systemd, reverse proxy, configuration |

## 📄 License

MIT — see [LICENSE](LICENSE).
