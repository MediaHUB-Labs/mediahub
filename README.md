# MediaHUB

**A self-hosted local media hub for organizing, uploading, and streaming your personal media collection.**

MediaHUB is a lightweight Go backend that turns any machine into a personal media server. Upload videos, images, audio, and documents — organize them by category and genre — and stream them from anywhere on your local network via a clean web UI.

### Key Features

- **Movies & Music** — Shared across all users, browse and stream from any device
- **Photo & Document Vault** — Private per-user storage, only the owner can access
- **Video Streaming** — HTTP 206 Range Requests for native `<video>` seeking
- **HLS Transcoding** — CPU-only (no GPU required) video transcoding with FFmpeg
- **Music Playlists** — Create and manage personal playlists
- **Continue Watching** — Pick up where you left off with automatic progress tracking
- **Thumbnails** — Auto-generated video thumbnails via FFmpeg
- **JWT Authentication** — Secure API with token-based auth on all protected routes

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
- [Access Control Model](#access-control-model)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [License](#license)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                         MediaHUB Server                              │
│                           (Go / Gin)                                 │
│                                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │   Auth   │  │  Media   │  │ Progress │  │ Playlist │            │
│  │  Module  │  │  Module  │  │  Module  │  │  Module  │            │
│  │          │  │          │  │          │  │          │            │
│  │ Handler  │  │ Handler  │  │ Handler  │  │ Handler  │            │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │            │
│  │ Repos.   │  │ Repos.   │  │ Repos.   │  │ Repos.   │            │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘            │
│       │              │              │              │                  │
│       └──────────────┼──────────────┼──────────────┘                  │
│                      │              │                                 │
│  ┌───────────┐ ┌─────┴──────┐ ┌────┴────────┐ ┌──────────────┐     │
│  │ Upload    │ │  SQLite DB │ │ Transcode   │ │  FFmpeg /    │     │
│  │ Module    │ │  (GORM)    │ │ Service     │ │  FFprobe     │     │
│  └───────────┘ └────────────┘ └─────────────┘ └──────────────┘     │
│                                                                      │
│  ┌───────────────────┐  ┌──────────────────────────────────────┐    │
│  │  JWT Middleware    │  │   Static File Server (mediahub-ui)  │    │
│  │  (Auth Gate)       │  │   Served at / via Gin middleware    │    │
│  └───────────────────┘  └──────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────┘
```

Each domain module follows a clean **3-layer architecture**:

| Layer          | Responsibility                                              |
| -------------- | ----------------------------------------------------------- |
| **Handler**    | HTTP request/response handling, input validation, status codes |
| **Service**    | Business logic, orchestration, access control               |
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
| Transcoding  | FFmpeg (optional, CPU-only — `libx264` software encoder)     |
| Frontend     | Vanilla JS + TailwindCSS (separate repo: `mediahub-ui`)     |

---

## Project Structure

```
mediahub/
├── main.go                          # Entry point — DB init, DI wiring, Gin setup, static serving
├── go.mod / go.sum                  # Go module dependencies
├── .env.example                     # Environment variable template
├── .gitignore
├── LICENSE                          # MIT License
│
├── auth/                            # Authentication module
│   ├── jwt.go                       #   JWT token generation & validation (env-based expiry)
│   ├── middleware.go                #   JWT auth middleware (Bearer token → user_id in context)
│   ├── handler/
│   │   └── user.go                  #   HTTP handlers: Signup, Login, GetUser, UpdateUser, DeleteUser
│   ├── service/
│   │   └── user.go                  #   Business logic: registration, login, password hashing
│   └── repository/
│       └── user.go                  #   DB CRUD: Create, FindByEmail, FindByID, Update, Delete, FindAll
│
├── media/                           # Media management module
│   ├── handler/
│   │   ├── media.go                 #   HTTP handlers: Upload, List, Search, Details, Update, Delete, Vault
│   │   └── stream.go               #   HTTP 206 streaming, thumbnail serving, HLS manifest
│   ├── service/
│   │   └── media.go                 #   Business logic: visibility rules, FFmpeg probing, auto-thumbnails
│   └── repository/
│       └── media.go                 #   DB CRUD: FindAll, Search, FindByUser, GetCategories, etc.
│
├── progress/                        # Playback progress tracking module
│   ├── handler/
│   │   └── handler.go              #   HTTP handlers: SaveProgress, ContinueWatching, ClearProgress
│   ├── service/
│   │   └── service.go              #   Business logic: upsert progress, auto-complete at ≥95%
│   └── repository/
│       └── repository.go           #   DB CRUD: Upsert, FindByUser, Delete, MarkCompleted
│
├── playlist/                        # Playlist management module
│   ├── handler/
│   │   └── handler.go              #   HTTP handlers: CRUD playlists, add/remove items
│   ├── service/
│   │   └── service.go              #   Business logic: ownership checks, audio-only validation
│   └── repository/
│       └── repository.go           #   DB CRUD: playlists + items, auto-position, cascading deletes
│
├── upload/                          # File upload module
│   ├── handler/
│   │   └── upload.go               #   Handler struct (scaffold)
│   ├── service/
│   │   └── upload.go               #   Service struct (scaffold)
│   └── repository/
│       └── upload.go               #   Upload logic: save-to-disk, MIME validation, SHA-256 checksum
│
├── transcode/                       # Video transcoding module
│   └── transcode.go                #   CPU-only HLS transcoding (libx264, veryfast preset)
│
├── models/                          # GORM data models
│   ├── user.go                      #   User model
│   ├── media.go                     #   Media model (with ownership & visibility fields)
│   ├── playlist.go                  #   Playlist & PlaylistItem models
│   └── userMediaProgress.go         #   Watch progress tracking model
│
├── dto/                             # Data Transfer Objects
│   ├── request.go                   #   Request structs: auth, media, progress, playlist
│   └── response.go                  #   Response structs: API, media, progress, playlist, FFmpeg
│
├── routes/
│   └── routes.go                    #   Route registration with JWT middleware on protected routes
│
├── utils/
│   ├── helper.go                    #   File-based logging utility
│   ├── validation.go                #   Validation error parser for clean API messages
│   ├── ffmpeg.go                    #   FFmpeg/FFprobe detection & media metadata probing
│   └── thumbnail.go                 #   Video thumbnail generation via FFmpeg
│
├── uploads/                         # File storage (auto-created, gitignored)
│   ├── videos/
│   ├── images/
│   ├── audio/
│   ├── documents/
│   ├── thumbnails/                  #   Auto-generated video thumbnails
│   └── transcoded/                  #   HLS transcoded segments & manifests
│
├── Logs/                            # Debug logs directory (date-stamped files)
│
└── mediahub-ui/                     # Frontend static files (cloned from mediahub-ui repo)
    ├── index.html
    ├── App.js                       #   SPA router & component renderer
    ├── assets/
    │   └── output.css               #   Compiled TailwindCSS
    ├── view/                        #   Pages & components
    │   ├── Home.js
    │   ├── pages/
    │   └── components/
    ├── src/
    │   ├── global/config.js         #   API base URL config
    │   ├── connection/              #   Server health check
    │   └── utils/                   #   Theme toggle, etc.
    └── README.md
```

---

## Getting Started

### Prerequisites

- **Go** 1.21+ installed ([download](https://go.dev/dl/))
- **Git**
- **FFmpeg** (optional) — for video thumbnails and HLS transcoding
  ```bash
  # Ubuntu/Debian
  sudo apt install ffmpeg

  # macOS
  brew install ffmpeg

  # Verify installation
  ffmpeg -version
  ```
  > MediaHUB works without FFmpeg — thumbnails and transcoding will simply be skipped.

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
JWT_SECRET=your-strong-secret-key-here

# JWT token expiry in minutes (1440 = 24 hours)
JWT_EXPIRY_MINUTES=1440

# Absolute path for file storage (leave empty = ./uploads relative to executable)
UPLOAD_PATH=

# Transcoding settings (CPU-only, no GPU needed)
MAX_TRANSCODE_JOBS=1
FFMPEG_PRESET=veryfast    # ultrafast | superfast | veryfast | faster | fast | medium
FFMPEG_CRF=23             # 0-51, lower = better quality, 23 = default
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

```
🚀 MediaHUB server starting on http://localhost:9123
📁 Upload path: /path/to/uploads
🔑 JWT secret: your****
✅ FFmpeg detected — transcoding and thumbnails enabled
```

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
   | `/uploads/*` | Upload directory (media files) |
   | `/transcoded/*` | HLS transcoded segments |
   | Any other path (SPA fallback) | `mediahub-ui/index.html` |

> **Note:** The `mediahub-ui/` directory is gitignored in this repo since it's managed as a separate repository. Always clone it fresh when setting up a new environment.

---

## Access Control Model

MediaHUB implements a visibility-based access control system:

| Media Type | Visibility | Who Can See | Who Can Manage |
|------------|-----------|-------------|----------------|
| **Movies** (video/*) | `public` | All users | Uploader only |
| **Music** (audio/*) | `public` | All users | Uploader only |
| **Photos** (image/*) | `private` | Owner only | Owner only |
| **Documents** (pdf, docx, txt) | `private` | Owner only | Owner only |

- **Public media** (movies, music) is visible to everyone and can be browsed, searched, and streamed by any authenticated user.
- **Private media** (photos, documents) acts as a personal vault — only the user who uploaded a file can see or manage it.
- **Playlists** are personal by default but can be made public by the owner.
- Visibility is **automatically assigned** based on the file's MIME type during upload.

---

## API Reference

All requests to protected endpoints require the header:
```
Authorization: Bearer <jwt_token>
```

### Health Check

| Method | Endpoint       | Auth | Description          |
| ------ | -------------- | ---- | -------------------- |
| `GET`  | `/api/health`  | No   | Server health check  |

### Auth (`/api/auth`)

| Method   | Endpoint           | Auth | Description           |
| -------- | ------------------ | ---- | --------------------- |
| `POST`   | `/api/auth/signup` | No   | Register a new user   |
| `POST`   | `/api/auth/login`  | No   | Login & get JWT token |
| `POST`   | `/api/auth/user`   | Yes  | Get user by ID        |
| `PUT`    | `/api/auth/user`   | Yes  | Update user profile   |
| `DELETE` | `/api/auth/user`   | Yes  | Delete user account   |

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

| Method   | Endpoint                    | Auth | Description                         |
| -------- | --------------------------- | ---- | ----------------------------------- |
| `GET`    | `/api/media/health`         | No   | Media module health check           |
| `POST`   | `/api/media/add`            | Yes  | Upload a media file (multipart)     |
| `GET`    | `/api/media/list`           | Yes  | Paginated media listing             |
| `GET`    | `/api/media/search?q=`      | Yes  | Search media by title/description   |
| `GET`    | `/api/media/categories`     | Yes  | List distinct categories            |
| `POST`   | `/api/media/details`        | Yes  | Get single media item details       |
| `PUT`    | `/api/media/metadata`       | Yes  | Update media metadata (owner only)  |
| `DELETE` | `/api/media/item`           | Yes  | Delete media (owner only)           |
| `GET`    | `/api/media/vault`          | Yes  | Personal vault (photos/docs)        |
| `GET`    | `/api/media/stream/:id`     | Yes  | Stream media (HTTP 206)             |
| `GET`    | `/api/media/thumbnail/:id`  | Yes  | Serve video thumbnail               |
| `GET`    | `/api/media/hls/:id`        | Yes  | Serve HLS manifest (.m3u8)          |
| `POST`   | `/api/media/transcode/:id`  | Yes  | Trigger HLS transcoding             |
| `GET`    | `/api/media/transcode/status` | Yes | Transcode service status           |

#### `POST /api/media/add` (multipart/form-data)

| Field          | Type   | Description                      |
| -------------- | ------ | -------------------------------- |
| `file`         | File   | The media file (required)        |
| `title`        | string | Display title                    |
| `description`  | string | Description text                 |
| `category`     | string | e.g. "Movie", "Home Video"       |
| `genres`       | string | Comma-separated, e.g. "Action,Sci-Fi" |
| `duration_sec` | uint   | Playback duration in seconds (auto-detected if FFmpeg available) |
| `resolution`   | string | e.g. "1920x1080" (auto-detected if FFmpeg available) |

#### `GET /api/media/list` (query parameters)

| Param      | Type   | Description                                |
| ---------- | ------ | ------------------------------------------ |
| `limit`    | int    | Items per page (default: 20, max: 100)     |
| `offset`   | int    | Pagination offset (default: 0)             |
| `category` | string | Filter by category                         |
| `genres`   | string | Filter by genre                            |
| `type`     | string | Filter: `video`, `audio`, `image`, `document` |

#### `GET /api/media/search?q=<query>` (query parameters)

| Param    | Type   | Description                        |
| -------- | ------ | ---------------------------------- |
| `q`      | string | Search term (required)             |
| `limit`  | int    | Items per page (default: 20)       |
| `offset` | int    | Pagination offset (default: 0)     |

#### `GET /api/media/vault` (query parameters)

| Param    | Type   | Description                        |
| -------- | ------ | ---------------------------------- |
| `type`   | string | Filter: `image` or `document`     |
| `limit`  | int    | Items per page (default: 20)       |
| `offset` | int    | Pagination offset (default: 0)     |

### Progress (`/api/progress`)

| Method   | Endpoint               | Auth | Description                         |
| -------- | ---------------------- | ---- | ----------------------------------- |
| `POST`   | `/api/progress/save`   | Yes  | Save/update playback position       |
| `GET`    | `/api/progress/continue` | Yes | List "Continue Watching" items     |
| `DELETE` | `/api/progress/clear`  | Yes  | Clear progress for a media item     |

#### `POST /api/progress/save`
```json
{
  "media_id": 1,
  "position_sec": 342.5
}
```
> Progress is automatically marked as completed when position reaches ≥95% of the media's total duration.

#### `GET /api/progress/continue?limit=10`
Returns a list of media items the user hasn't finished, sorted by most recently watched.

#### `DELETE /api/progress/clear`
```json
{
  "media_id": 1
}
```

### Playlists (`/api/playlist`)

| Method   | Endpoint                     | Auth | Description                |
| -------- | ---------------------------- | ---- | -------------------------- |
| `POST`   | `/api/playlist/create`       | Yes  | Create a playlist          |
| `GET`    | `/api/playlist/list`         | Yes  | List user's playlists      |
| `GET`    | `/api/playlist/:id`          | Yes  | Get playlist with items    |
| `PUT`    | `/api/playlist/:id`          | Yes  | Update playlist metadata   |
| `DELETE` | `/api/playlist/:id`          | Yes  | Delete playlist            |
| `POST`   | `/api/playlist/:id/add`      | Yes  | Add song to playlist       |
| `DELETE` | `/api/playlist/:id/remove`   | Yes  | Remove song from playlist  |

#### `POST /api/playlist/create`
```json
{
  "name": "My Chill Mix",
  "description": "Relaxing tunes",
  "is_public": false
}
```

#### `POST /api/playlist/:id/add`
```json
{
  "media_id": 5
}
```
> Only audio/music media can be added to playlists.

---

## Database Schema

MediaHUB uses **SQLite** with GORM auto-migration. Five tables are created:

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
| Column              | Type   | Constraints            |
| ------------------- | ------ | ---------------------- |
| id                  | uint   | Primary Key, Auto      |
| created_at          | datetime |                      |
| updated_at          | datetime |                      |
| deleted_at          | datetime | Soft Delete index    |
| file_path           | string | unique, not null       |
| mime_type           | string | not null               |
| uploaded_by_user_id | uint   | FK → users, indexed    |
| visibility          | string | "public" or "private", indexed |
| title               | string | not null               |
| description         | string |                        |
| category            | string | indexed                |
| genres              | string |                        |
| file_size_kb        | uint   | not null               |
| checksum            | string | unique (SHA-256)       |
| duration_sec        | uint   |                        |
| resolution          | string |                        |
| thumbnail_path      | string |                        |
| is_transcoded       | bool   | default: false         |
| transcoded_path     | string |                        |
| is_new              | bool   | default: true          |

### `user_media_progresses`
| Column                | Type    | Constraints                          |
| --------------------- | ------- | ------------------------------------ |
| id                    | uint    | Primary Key, Auto                    |
| user_id               | uint    | FK → users, composite unique         |
| media_id              | uint    | FK → media, composite unique         |
| playhead_position_sec | float64 | default: 0                           |
| last_watched_at       | datetime | not null                            |
| is_completed          | bool    | default: false                       |

### `playlists`
| Column      | Type     | Constraints              |
| ----------- | -------- | ------------------------ |
| id          | uint     | Primary Key, Auto        |
| created_at  | datetime |                          |
| updated_at  | datetime |                          |
| deleted_at  | datetime | Soft Delete index        |
| name        | string   | not null                 |
| description | string   |                          |
| user_id     | uint     | FK → users, indexed      |
| is_public   | bool     | default: false           |

### `playlist_items`
| Column      | Type | Constraints                        |
| ----------- | ---- | ---------------------------------- |
| id          | uint | Primary Key, Auto                  |
| created_at  | datetime |                              |
| updated_at  | datetime |                              |
| deleted_at  | datetime | Soft Delete index            |
| playlist_id | uint | FK → playlists, composite unique   |
| media_id    | uint | FK → media, composite unique       |
| position    | uint | not null, default: 0 (ordering)    |

---

## Streaming & Transcoding

### HTTP 206 Streaming

MediaHUB serves media files with **HTTP Range Request** support (`206 Partial Content`). This enables:
- Native `<video>` and `<audio>` element seeking in the browser
- Progressive download without loading the entire file
- Efficient bandwidth usage on the local network

Stream a file: `GET /api/media/stream/:id` with `Authorization` header.

### HLS Transcoding (Optional)

For adaptive bitrate streaming, MediaHUB can transcode videos into **HLS format** (`.m3u8` + `.ts` segments) using FFmpeg's software encoder — **no GPU required**.

| Setting | Default | Description |
|---------|---------|-------------|
| `MAX_TRANSCODE_JOBS` | 1 | Max concurrent transcode jobs |
| `FFMPEG_PRESET` | `veryfast` | FFmpeg encoding speed (use `ultrafast` on very slow systems) |
| `FFMPEG_CRF` | `23` | Quality level (0-51, lower = better) |

Trigger transcoding: `POST /api/media/transcode/:id`

The transcoding runs in the background and scales the video to **720p** for efficient streaming. On a typical i5/Ryzen 5, a 1080p movie transcodes at roughly 0.5-1x realtime speed.

Check status: `GET /api/media/transcode/status`

---

## License

[MIT](./LICENSE) — Copyright (c) 2025 MediaHUB-Labs
