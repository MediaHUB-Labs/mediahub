# MediaHUB

**A self-hosted local media hub for organizing, uploading, and streaming your personal media collection.**

MediaHUB is a lightweight Go backend that turns any machine into a personal media server. Upload videos, images, audio, and documents — organize them by category and genre — and stream them from anywhere on your local network via a clean web UI.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Environment Configuration](#environment-configuration)
  - [Running the Server](#running-the-server)
- [Setting Up the UI (mediahub-ui)](#setting-up-the-ui-mediahub-ui)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [Codebase Analysis](#codebase-analysis)
- [License](#license)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                     MediaHUB Server                     │
│                       (Go / Gin)                        │
│                                                         │
│  ┌───────────┐   ┌───────────┐   ┌──────────────────┐  │
│  │   Auth     │   │   Media   │   │     Upload       │  │
│  │  Module    │   │  Module   │   │     Module       │  │
│  │           │   │           │   │                  │  │
│  │ Handler   │   │ Handler   │   │ Handler          │  │
│  │ Service   │   │ Service   │   │ Service          │  │
│  │ Repository│   │ Repository│   │ Repository       │  │
│  └─────┬─────┘   └─────┬─────┘   └────────┬─────────┘  │
│        │               │                   │            │
│        └───────────────┼───────────────────┘            │
│                        │                                │
│                 ┌──────┴──────┐                          │
│                 │  SQLite DB  │                          │
│                 │ (GORM ORM)  │                          │
│                 └─────────────┘                          │
│                                                         │
│  ┌─────────────────────────────────────────────────┐    │
│  │         Static File Server (mediahub-ui)        │    │
│  │      Served at / via Gin's Static middleware     │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

Each domain module (Auth, Media, Upload) follows a clean **3-layer architecture**:

| Layer          | Responsibility                                              |
| -------------- | ----------------------------------------------------------- |
| **Handler**    | HTTP request/response handling, input validation, status codes |
| **Service**    | Business logic, orchestration between repositories          |
| **Repository** | Direct database access via GORM                             |

---

## Tech Stack

| Component    | Technology                                                   |
| ------------ | ------------------------------------------------------------ |
| Language     | Go 1.25                                                      |
| HTTP Router  | [Gin](https://github.com/gin-gonic/gin) v1.11               |
| Database     | SQLite (via [glebarez/sqlite](https://github.com/glebarez/sqlite) — pure Go, no CGO) |
| ORM          | [GORM](https://gorm.io/) v1.30                              |
| Auth         | JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt) v5) + bcrypt |
| Validation   | [go-playground/validator](https://github.com/go-playground/validator) v10 |
| Config       | `.env` files via [godotenv](https://github.com/joho/godotenv) |
| Frontend     | Vanilla JS + TailwindCSS (separate repo: `mediahub-ui`)     |

---

## Project Structure

```
mediahub/
├── main.go                  # Entry point — DB init, DI wiring, Gin setup, static serving
├── go.mod / go.sum          # Go module dependencies
├── .env.example             # Environment variable template
├── .gitignore
├── LICENSE                  # MIT License
│
├── auth/                    # Authentication module
│   ├── jwt.go               #   JWT token generation & validation
│   ├── handler/
│   │   └── user.go          #   HTTP handlers: Signup, Login, GetUser, UpdateUser, DeleteUser
│   ├── service/
│   │   └── user.go          #   Business logic: registration, login, password hashing
│   └── repository/
│       └── user.go          #   DB CRUD: Create, FindByEmail, FindByID, Update, Delete, FindAll
│
├── media/                   # Media management module
│   ├── handler/
│   │   └── media.go         #   HTTP handlers: MediaHealth, UploadMedia
│   ├── service/
│   │   └── media.go         #   Business logic: file upload orchestration, DB insert with rollback
│   └── repository/
│       └── media.go         #   DB CRUD: Create
│
├── upload/                  # File upload module
│   ├── handler/
│   │   └── upload.go        #   Handler struct (scaffold, not yet wired)
│   ├── service/
│   │   └── upload.go        #   Service struct (scaffold)
│   └── repository/
│       └── upload.go        #   Core upload logic: save-to-disk, MIME detection, SHA-256 checksum
│
├── models/                  # GORM data models
│   ├── user.go              #   User model
│   ├── media.go             #   Media model (videos, images, docs, audio)
│   └── userMediaProgress.go #   Watch progress tracking model
│
├── dto/                     # Data Transfer Objects
│   ├── request.go           #   Request structs with validation tags
│   └── response.go          #   Response structs (ApiResponse, AuthResponse, etc.)
│
├── routes/
│   └── routes.go            #   Route registration: /api/auth/* and /api/media/*
│
├── utils/
│   ├── helper.go            #   File-based logging utility
│   └── validation.go        #   Validation error parser for clean API messages
│
├── uploads/                 # File storage (auto-created, gitignored)
│   ├── videos/
│   ├── images/
│   ├── audios/
│   └── documents/
│
├── Logs/                    # Debug logs directory (date-stamped files)
│
├── mediahub-ui/             # Frontend static files (cloned from mediahub-ui repo)
│   ├── index.html
│   ├── App.js               #   SPA router & component renderer
│   ├── assets/
│   │   └── output.css       #   Compiled TailwindCSS
│   ├── view/                #   Pages & components
│   │   ├── Home.js
│   │   ├── pages/
│   │   └── components/
│   ├── src/
│   │   ├── global/config.js #   API base URL config
│   │   ├── connection/      #   Server health check
│   │   └── utils/           #   Theme toggle, etc.
│   └── README.md
│
└── mediahub-server          # Compiled binary (gitignored recommended)
```

---

## Getting Started

### Prerequisites

- **Go** 1.21+ installed ([download](https://go.dev/dl/))
- **Git**

### Installation

```bash
# Clone the repository
git clone https://github.com/MediaHUB-Labs/mediahub.git
cd mediahub

# Download Go dependencies
go mod download
```

### Environment Configuration

```bash
# Copy the example env file
cp .env.example .env
```

Edit `.env` with your values:

```env
# Server port (default: 9123)
APP_PORT=9123

# SQLite database file path
DB_PATH=mediahub.db

# Gin mode: "debug" for development, "release" for production
GIN_MODE=debug

# Enable file logging: "true" or "false"
LOGGER=true

# JWT secret key — CHANGE THIS in production!
JWT_SECRET=your-secret-key-here

# JWT token expiry in minutes (1440 = 24 hours)
JWT_EXPIRY_MINUTES=1440
```

### Running the Server

```bash
# Development (with hot output)
go run main.go

# Build & run a production binary
go build -o mediahub-server .
./mediahub-server
```

The server starts at `http://localhost:9123` (or your configured `APP_PORT`).

Verify it's running:
```bash
curl http://localhost:9123/api/health
# → {"message":"OK"}
```

---

## Setting Up the UI (mediahub-ui)

The frontend lives in a **separate repository**: [`mediahub-ui`](https://github.com/MediaHUB-Labs/mediahub-ui). The backend serves it as static files from the `./mediahub-ui/` directory.

### Steps

1. **Clone the UI repo into the backend project root:**

   ```bash
   cd /path/to/mediahub
   git clone https://github.com/MediaHUB-Labs/mediahub-ui.git
   ```

   This creates the `mediahub-ui/` folder that the Go server expects.

2. **Configure the API base URL:**

   Edit `mediahub-ui/src/global/config.js`:

   ```javascript
   export const CONFIG = {
       API_BASE_URL: "http://<YOUR-SERVER-IP>:9123/api",
       ENDPOINTS: {
           HEALTH: "/health",
           LOGIN: "/auth/login",
           SIGNUP: "/auth/signup"
       }
   };
   ```

   Replace `<YOUR-SERVER-IP>` with:
   - `localhost` — if accessing from the same machine
   - Your machine's LAN IP (e.g. `192.168.0.105`) — if accessing from other devices on the network

3. **(Optional) Rebuild TailwindCSS** — if you modify any UI styles:

   ```bash
   cd mediahub-ui
   chmod +x tailwindBuild.sh
   ./tailwindBuild.sh
   ```

4. **Start the backend server** — the UI is automatically served:

   ```bash
   # From the mediahub root
   go run main.go
   ```

   Open `http://<YOUR-SERVER-IP>:9123` in your browser. The Go server handles:

   | URL Path    | Serves From                  |
   | ----------- | ---------------------------- |
   | `/`         | `mediahub-ui/index.html`     |
   | `/view/*`   | `mediahub-ui/view/`          |
   | `/src/*`    | `mediahub-ui/src/`           |
   | `/assets/*` | `mediahub-ui/assets/`        |
   | `/App.js`   | `mediahub-ui/App.js`         |
   | `/output.css` | `mediahub-ui/assets/output.css` |
   | Any other path (SPA fallback) | `mediahub-ui/index.html` |

> **Note:** The `mediahub-ui/` directory is gitignored in this repo since it's managed as a separate repository. Always clone it fresh when setting up a new environment.

---

## API Reference

### Health Check

| Method | Endpoint       | Description          |
| ------ | -------------- | -------------------- |
| `GET`  | `/api/health`  | Server health check  |

### Auth (`/api/auth`)

| Method   | Endpoint          | Description           | Auth Required |
| -------- | ----------------- | --------------------- | ------------- |
| `POST`   | `/api/auth/signup` | Register a new user   | No            |
| `POST`   | `/api/auth/login`  | Login & get JWT token | No            |
| `POST`   | `/api/auth/user`   | Get user by ID        | No*           |
| `PUT`    | `/api/auth/user`   | Update user profile   | No*           |
| `DELETE` | `/api/auth/user`   | Delete user account   | No*           |

> *Auth middleware is not yet applied to these routes — see [Roadmap](./ROADMAP.md).

#### `POST /api/auth/signup`
```json
{
  "email": "user@example.com",
  "password": "secret123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### `POST /api/auth/login`
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

#### Response format (all endpoints):
```json
{
  "success": true,
  "message": "Login successful",
  "data": { ... },
  "error": ""
}
```

### Media (`/api/media`)

| Method | Endpoint          | Description                    | Auth Required |
| ------ | ----------------- | ------------------------------ | ------------- |
| `GET`  | `/api/media/health` | Media module health check    | No            |
| `POST` | `/api/media/add`  | Upload a media file (multipart) | No*          |

#### `POST /api/media/add` (multipart/form-data)

| Field          | Type   | Description                      |
| -------------- | ------ | -------------------------------- |
| `file`         | File   | The media file (required)        |
| `title`        | string | Display title                    |
| `description`  | string | Description text                 |
| `category`     | string | e.g. "Movie", "Home Video"       |
| `genres`       | string | Comma-separated, e.g. "Action,Sci-Fi" |
| `duration_sec` | uint   | Playback duration in seconds     |
| `resolution`   | string | e.g. "1920x1080"                 |

---

## Database Schema

MediaHUB uses **SQLite** with GORM auto-migration. Three tables are created:

### `users`
| Column       | Type     | Constraints          |
| ------------ | -------- | -------------------- |
| id           | uint     | Primary Key, Auto    |
| created_at   | datetime |                      |
| updated_at   | datetime |                      |
| deleted_at   | datetime | Soft Delete index    |
| first_name   | string   | nullable             |
| last_name    | string   | nullable             |
| email        | string   | unique, not null     |
| password     | string   | not null (bcrypt)    |
| is_active    | bool     | default: true        |

### `media`
| Column          | Type   | Constraints            |
| --------------- | ------ | ---------------------- |
| id              | uint   | Primary Key, Auto      |
| created_at      | datetime |                      |
| updated_at      | datetime |                      |
| deleted_at      | datetime | Soft Delete index    |
| file_path       | string | unique, not null       |
| mime_type       | string | not null               |
| title           | string | not null               |
| description     | string |                        |
| category        | string |                        |
| genres          | string |                        |
| file_size_kb    | uint   | not null               |
| checksum        | string | unique (SHA-256)       |
| duration_sec    | uint   |                        |
| resolution      | string |                        |
| thumbnail_path  | string |                        |
| is_transcoded   | bool   | default: false         |
| is_new          | bool   | default: true          |

### `user_media_progresses`
| Column                | Type    | Constraints                          |
| --------------------- | ------- | ------------------------------------ |
| id                    | uint    | Primary Key, Auto                    |
| user_id               | uint    | FK → users, composite unique         |
| media_id              | uint    | FK → media, composite unique         |
| playhead_position_sec | float64 | default: 0                           |
| last_watched_at       | datetime | not null                            |
| is_completed          | bool    | default: false                       |

---

## Codebase Analysis

### What's Working ✅

- **User Authentication** — Full CRUD: signup (with bcrypt), login (with JWT), get/update/delete user
- **Media Upload** — Multipart file upload with automatic MIME detection, sub-folder routing (`videos/`, `images/`, `audios/`, `documents/`), SHA-256 checksum for deduplication, and DB rollback on failure
- **Static UI Serving** — The Go server serves the `mediahub-ui` SPA with proper SPA fallback routing
- **File Logging** — Date-stamped debug logs written to `Logs/` directory
- **Validation** — Clean validation error messages via `go-playground/validator`
- **CORS** — Permissive CORS middleware for local network access

### What's Scaffolded / In Progress 🚧

- **Media listing, search, details, streaming** — Routes defined but commented out in `routes/routes.go`
- **User progress tracking** — Model exists (`UserMediaProgress`) but routes/handlers are not wired
- **Upload handler/service** — Struct scaffolded but upload logic lives directly in the repository layer
- **JWT middleware** — Token generation/validation functions exist but no Gin middleware to protect routes

### Areas for Improvement ⚠️

- **No auth middleware on protected routes** — User CRUD and media endpoints are currently unprotected
- **JWT expiry env var unused** — `JWT_EXPIRY_MINUTES` is defined in `.env` but the code hardcodes 24 hours
- **Error string matching** — Duplicate detection uses `strings.Contains` on error messages (fragile)
- **No file size limits** — Upload endpoint has no max file size enforcement
- **Upload path is relative** — Uses `"uploads"` relative path; may break depending on working directory
- **Potential nil error in Register** — Line 52 in `auth/service/user.go` returns `err` instead of `tokenErr`

---

## License

[MIT](./LICENSE) — Copyright (c) 2025 MediaHUB-Labs
