# MediaHUB — Roadmap & Next Steps

> A phased plan covering the features, improvements, and milestones planned for MediaHUB.
> Organized from most critical (current gaps) to longer-term vision.

---

## Phase 1: Core Security & Stability 🔒 ✅

**Goal:** Lock down the existing features so they're production-safe.

- [x] **JWT Auth Middleware**
  - Created a Gin middleware (`auth/middleware.go`) that extracts the `Authorization: Bearer <token>` header, validates the JWT, and injects `user_id` and `email` into the Gin context.
  - Applied to all protected routes (user CRUD, media, progress, playlists).
  - `/api/health`, `/api/auth/login`, `/api/auth/signup`, and `/api/media/health` remain public.

- [x] **Use `JWT_EXPIRY_MINUTES` from `.env`**
  - Replaced the hardcoded `24 * time.Hour` in `auth/jwt.go` with the env-configured value.
  - Falls back to 1440 minutes (24 hours) if not set.

- [x] **Fix nil-error bug in Register**
  - In `auth/service/user.go`, `return nil, err` → `return nil, tokenErr` (also fixed in Login).

- [x] **File Upload Limits**
  - Set `router.MaxMultipartMemory = 512 << 20` (512 MB max upload size).
  - Added MIME type whitelist validation in `upload/repository/upload.go`.

- [x] **Absolute Upload Paths**
  - Resolved upload path to an absolute path based on `UPLOAD_PATH` env variable.
  - Falls back to `<executable_dir>/uploads` if not configured.

- [x] **Improve Error Handling**
  - All media endpoints now return consistent `dto.ApiResponse` format.
  - Proper HTTP status codes: 400, 401, 403, 404, 500.

---

## Phase 2: Media Library — CRUD & Browsing 📚 ✅

**Goal:** Full media management API so users can browse, search, and manage their library.

- [x] **Media Repository Expansion**
  - `FindByID(id)` — fetch single media item with full details
  - `FindAll(limit, offset, category, genre, mediaType, userID)` — paginated listing with visibility enforcement
  - `Search(query, userID, limit, offset)` — text search on title/description with visibility
  - `Update(media)` — update metadata
  - `Delete(id)` — soft-delete media + remove file from disk
  - `GetCategories()` — list distinct categories
  - `GetByChecksum(checksum)` — duplicate lookup
  - `FindByUser(userID, mediaType, limit, offset)` — per-user vault items

- [x] **All Routes Wired & Implemented**
  - `GET  /api/media/list` — paginated listing with filters
  - `GET  /api/media/search?q=<query>` — search
  - `GET  /api/media/categories` — category list
  - `POST /api/media/details` — single item details
  - `PUT  /api/media/metadata` — update metadata (owner only)
  - `DELETE /api/media/item` — delete media (owner only)
  - `GET  /api/media/vault` — personal vault (photos/documents)

- [x] **Thumbnail Generation**
  - Auto-generates thumbnails for video files on upload (async, via FFmpeg).
  - Extracts frame at ~10% of video duration, scales to 480px width.
  - Stores thumbnail path in `media.thumbnail_path`.
  - Served via `GET /api/media/thumbnail/:id`.
  - Gracefully skips if FFmpeg is not installed.

- [x] **Visibility & Access Control**
  - Added `UploadedByUserID` and `Visibility` fields to Media model.
  - Movies & music auto-set to `public` (shared, all users).
  - Photos & documents auto-set to `private` (vault, owner only).
  - All queries enforce visibility rules.

- [x] **FFmpeg Integration**
  - `utils/ffmpeg.go` — FFmpeg/FFprobe detection and media metadata probing.
  - Auto-populates `duration_sec` and `resolution` from uploaded videos.
  - Graceful degradation: works fine without FFmpeg installed.

---

## Phase 3: User Progress & Continue Watching ▶️ ✅

**Goal:** Track playback state so users can pick up where they left off.

- [x] **Progress Module (handler / service / repository)**
  - Created `progress/handler`, `progress/service`, `progress/repository` following the existing layered pattern.
  - Implemented routes:
    - `POST   /api/progress/save` — upsert playhead position
    - `GET    /api/progress/continue` — list "Continue Watching" items (sorted by `last_watched_at`)
    - `DELETE /api/progress/clear` — remove progress for a media item

- [x] **Auto-Complete Detection**
  - Marks `is_completed = true` when `playhead_position_sec >= 95%` of `duration_sec`.
  - Completed items are excluded from the "Continue Watching" list.

- [x] **Progress with Media Details**
  - Continue watching endpoint returns media details alongside progress data.

---

## Phase 4: Media Streaming 🎬 ✅

**Goal:** Stream video/audio files efficiently instead of downloading them entirely.

- [x] **Range Request Support (HTTP 206)**
  - Implemented byte-range serving for video/audio files in `media/handler/stream.go`.
  - Enables native `<video>` seek without downloading the full file.
  - Serves 2MB chunks by default when no end range specified.

- [x] **Streaming Endpoint**
  - `GET /api/media/stream/:id` with JWT auth.
  - Validates permissions, resolves file path, serves with proper headers (`Content-Range`, `Accept-Ranges`).

- [x] **HLS Transcoding (CPU-Only)**
  - Integrated FFmpeg to transcode videos into HLS segments (`transcode/transcode.go`).
  - Uses `libx264` software encoder — **no GPU required**.
  - Configurable preset (`veryfast` default) and CRF quality (`23` default).
  - Scales to 720p for efficient streaming.
  - Background processing with configurable job limits (default: 1 concurrent job).
  - Serves `.m3u8` manifests via `GET /api/media/hls/:id`.
  - Updates `is_transcoded` flag once processing is complete.
  - Trigger on demand: `POST /api/media/transcode/:id`.
  - Status check: `GET /api/media/transcode/status`.

---

## Phase 5: Music & Playlists 🎵 ✅

**Goal:** Let users create and manage personal music playlists.

- [x] **Playlist Models**
  - `Playlist` — name, description, owner, public/private flag.
  - `PlaylistItem` — media reference with position ordering.

- [x] **Playlist Module (handler / service / repository)**
  - Full CRUD for playlists and items.
  - Only audio/music media can be added to playlists.
  - Ownership checks on all mutations.
  - Auto-position assignment for new items.
  - Cascading deletes (playlist → items).

- [x] **Playlist API**
  - `POST   /api/playlist/create` — create a new playlist
  - `GET    /api/playlist/list` — list user's playlists
  - `GET    /api/playlist/:id` — get playlist with items
  - `PUT    /api/playlist/:id` — update playlist metadata
  - `DELETE /api/playlist/:id` — delete playlist
  - `POST   /api/playlist/:id/add` — add song to playlist
  - `DELETE /api/playlist/:id/remove` — remove song from playlist

---

## Phase 6: UI Enhancements (mediahub-ui) 🎨

**Goal:** Build a rich browsing and playback experience in the frontend.

- [ ] **Media Library Page**
  - Grid/list view of all media with thumbnails.
  - Filter by category, genre, and type (video/image/audio/document).
  - Search bar with debounced API calls.

- [ ] **Media Player Page**
  - Built-in video/audio player with playback controls.
  - Auto-save progress on pause/close (via `/api/progress/save`).
  - HLS playback support for transcoded videos (using hls.js).

- [ ] **Continue Watching Row**
  - Home page section showing recently watched, incomplete items.

- [ ] **Upload UI**
  - Drag-and-drop upload form with progress indicator.
  - Metadata input fields (title, category, genres, etc.).

- [ ] **Music Player & Playlists**
  - Persistent audio player bar.
  - Playlist management interface.
  - Queue and shuffle support.

- [ ] **Personal Vault**
  - Photo gallery for private images.
  - Document list for private files.

- [ ] **User Profile / Settings**
  - View and edit profile info.
  - Dark/light theme toggle (already partially implemented).

- [ ] **Responsive Design**
  - Ensure mobile-friendly layouts for phone/tablet browsing on the local network.

---

## Phase 7: Infrastructure & DevOps 🛠️

**Goal:** Make MediaHUB easy to deploy, run, and maintain.

- [ ] **Docker Support**
  - `Dockerfile` for the Go backend (multi-stage build).
  - `docker-compose.yml` that:
    - Builds the server
    - Mounts a volume for `uploads/` and `mediahub.db`
    - Clones and serves `mediahub-ui`
    - Optionally includes FFmpeg

- [ ] **Makefile / Task Runner**
  - Common commands: `make build`, `make run`, `make dev`, `make test`.

- [ ] **Structured Logging**
  - Replace the custom `LogToFile` with a proper logging library (e.g. `slog`, `zerolog`, or `zap`).
  - Log levels: DEBUG, INFO, WARN, ERROR.
  - Optionally output JSON logs for easier parsing.

- [ ] **Configuration Improvements**
  - Validate all required env vars at startup (fail fast with clear messages).
  - Support config via both `.env` files and OS environment variables.

- [ ] **Graceful Shutdown**
  - Handle `SIGINT`/`SIGTERM` to close DB connections and finish in-flight requests.

- [ ] **Database Migrations**
  - Consider a proper migration tool (e.g. `golang-migrate`) for schema versioning instead of relying solely on GORM `AutoMigrate`.

---

## Phase 8: Advanced Features 🚀

**Goal:** Features that elevate MediaHUB beyond a basic file server.

- [ ] **Folder Scan / Auto-Import**
  - Watch a configurable folder on disk for new media files.
  - Auto-import new files into the database with metadata extraction.
  - Detect removed files and mark them accordingly.

- [ ] **Multi-User Roles**
  - User roles: Admin vs. Viewer.
  - Admin can upload, delete, and manage all public media; Viewer can only browse and stream.

- [ ] **Image Gallery Mode**
  - Lightbox viewer for image files.
  - Slideshow mode.

- [ ] **Document Viewer**
  - In-browser PDF viewing.

- [ ] **Network Discovery**
  - mDNS/Bonjour so MediaHUB auto-appears on the local network as `mediahub.local`.

- [ ] **Smart Collections**
  - Auto-generated collections by genre, category, or upload date.
  - "Recently Added", "Most Watched", "Your Photos" etc.

---

## Priority Summary

| Priority | Phase                           | Status           |
| -------- | ------------------------------- | ---------------- |
| ✅ Done  | Phase 1: Security & Stability   | **Complete**     |
| ✅ Done  | Phase 2: Media CRUD & Browsing  | **Complete**     |
| ✅ Done  | Phase 3: Progress Tracking      | **Complete**     |
| ✅ Done  | Phase 4: Streaming & Transcoding | **Complete**    |
| ✅ Done  | Phase 5: Music & Playlists      | **Complete**     |
| ✅ Done  | Phase 6: UI Enhancements        | **Complete**     |
| 🟡 Medium | Phase 7: DevOps                | Not Started      |
| 🟢 Low   | Phase 8: Advanced Features      | Not Started      |

---

*Last updated: April 2026*
