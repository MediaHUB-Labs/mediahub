# MediaHUB — Roadmap & Next Steps

> A phased plan covering the features, improvements, and milestones planned for MediaHUB.
> Organized from most critical (current gaps) to longer-term vision.

---

## Phase 1: Core Security & Stability 🔒

**Goal:** Lock down the existing features so they're production-safe.

- [ ] **JWT Auth Middleware**
  - Create a Gin middleware that extracts the `Authorization: Bearer <token>` header, validates the JWT, and injects `user_id` into the Gin context.
  - Apply it to all protected routes (user CRUD, media upload, progress, etc.).
  - Keep `/api/health`, `/api/auth/login`, and `/api/auth/signup` as public routes.

- [ ] **Use `JWT_EXPIRY_MINUTES` from `.env`**
  - Replace the hardcoded `24 * time.Hour` in `auth/jwt.go` with the env-configured value.

- [ ] **Fix nil-error bug in Register**
  - In `auth/service/user.go` line 52, `return nil, err` should be `return nil, tokenErr`.

- [ ] **File Upload Limits**
  - Set a max upload size in Gin (e.g. `router.MaxMultipartMemory = 512 << 20` for 512 MB).
  - Validate allowed MIME types server-side (reject unexpected file types).

- [ ] **Absolute Upload Paths**
  - Resolve `"uploads"` to an absolute path based on the executable directory (similar to `LogToFile`).
  - Optionally make the base upload path configurable via `.env` (`UPLOAD_PATH`).

- [ ] **Improve Error Handling**
  - Replace fragile `strings.Contains` error matching with GORM error types or custom error wrappers.
  - Return consistent `dto.ApiResponse` from all media endpoints (currently uses raw `gin.H`).

---

## Phase 2: Media Library — CRUD & Browsing 📚

**Goal:** Build out the full media management API so users can browse, search, and manage their library.

- [ ] **Media Repository Expansion**
  - `FindByID(id)` — fetch single media item with full details
  - `FindAll(limit, offset, filters)` — paginated listing with optional filters (category, genre, MIME type)
  - `Search(query)` — full-text search on title/description
  - `Update(media)` — update metadata (title, description, category, genres)
  - `Delete(id)` — soft-delete media + optionally remove file from disk
  - `GetCategories()` — list distinct categories
  - `GetByChecksum(checksum)` — duplicate lookup

- [ ] **Wire Commented-Out Routes**
  - Uncomment and implement handlers for:
    - `GET  /api/media/list` — paginated listing
    - `GET  /api/media/search?q=<query>` — search
    - `GET  /api/media/categories` — category list
    - `POST /api/media/details` — single item details
    - `PUT  /api/media/metadata` — update metadata
    - `DELETE /api/media/item` — delete media

- [ ] **Thumbnail Generation**
  - Auto-generate thumbnails for video files on upload (via FFmpeg).
  - Store thumbnail path in `media.thumbnail_path`.
  - Serve thumbnails via a static route (e.g. `/api/media/thumbnail/:id`).

---

## Phase 3: User Progress & Continue Watching ▶️

**Goal:** Track playback state so users can pick up where they left off.

- [ ] **Progress Handlers & Service**
  - Create `progress/handler`, `progress/service`, `progress/repository` following the existing layered pattern.
  - Implement the commented-out routes:
    - `POST   /api/progress/save` — upsert playhead position
    - `GET    /api/progress/continue` — list "Continue Watching" items (sorted by `last_watched_at`)
    - `DELETE /api/progress/clear` — remove progress for a media item

- [ ] **Auto-Complete Detection**
  - Mark `is_completed = true` when `playhead_position_sec >= 95%` of `duration_sec`.

- [ ] **Progress in Media Details**
  - When fetching media details, include the requesting user's progress data if available.

---

## Phase 4: Media Streaming 🎬

**Goal:** Stream video/audio files efficiently instead of downloading them entirely.

- [ ] **Range Request Support (HTTP 206)**
  - Implement byte-range serving for video/audio files.
  - This enables native `<video>` seek without downloading the full file.

- [ ] **Streaming Endpoint**
  - `POST /api/media/play` or `GET /api/media/stream/:id`
  - Validate auth, resolve file path, serve with proper headers (`Content-Range`, `Accept-Ranges`).

- [ ] **HLS/DASH Transcoding (Future)**
  - Integrate FFmpeg to transcode uploaded videos into HLS segments.
  - Serve `.m3u8` manifests for adaptive bitrate streaming.
  - Update `is_transcoded` flag once processing is complete.

---

## Phase 5: UI Enhancements (mediahub-ui) 🎨

**Goal:** Build a rich browsing and playback experience in the frontend.

- [ ] **Media Library Page**
  - Grid/list view of all media with thumbnails.
  - Filter by category, genre, and type (video/image/audio/document).
  - Search bar with debounced API calls.

- [ ] **Media Player Page**
  - Built-in video/audio player with playback controls.
  - Auto-save progress on pause/close (via `/api/progress/save`).

- [ ] **Continue Watching Row**
  - Home page section showing recently watched, incomplete items.

- [ ] **Upload UI**
  - Drag-and-drop upload form with progress indicator.
  - Metadata input fields (title, category, genres, etc.).

- [ ] **User Profile / Settings**
  - View and edit profile info.
  - Dark/light theme toggle (already partially implemented).

- [ ] **Responsive Design**
  - Ensure mobile-friendly layouts for phone/tablet browsing on the local network.

---

## Phase 6: Infrastructure & DevOps 🛠️

**Goal:** Make MediaHUB easy to deploy, run, and maintain.

- [ ] **Docker Support**
  - `Dockerfile` for the Go backend (multi-stage build).
  - `docker-compose.yml` that:
    - Builds the server
    - Mounts a volume for `uploads/` and `mediahub.db`
    - Clones and serves `mediahub-ui`

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

## Phase 7: Advanced Features 🚀

**Goal:** Features that elevate MediaHUB beyond a basic file server.

- [ ] **Folder Scan / Auto-Import**
  - Watch a configurable folder on disk for new media files.
  - Auto-import new files into the database with metadata extraction.
  - Detect removed files and mark them accordingly.

- [ ] **Media Metadata Extraction**
  - Use FFprobe / FFmpeg to extract video duration, resolution, codec info on upload.
  - Auto-populate `duration_sec`, `resolution`, and other fields.

- [ ] **Multi-User Support**
  - User roles: Admin vs. Viewer.
  - Admin can upload, delete, and manage; Viewer can only browse and stream.
  - Per-user progress tracking is already modeled.

- [ ] **Collections / Playlists**
  - Group media items into named collections (e.g. "Family Vacation 2025").
  - Auto-generated collections by genre or category.

- [ ] **Image Gallery Mode**
  - Lightbox viewer for image files.
  - Slideshow mode.

- [ ] **Document Viewer**
  - In-browser PDF viewing.

- [ ] **Network Discovery**
  - mDNS/Bonjour so MediaHUB auto-appears on the local network as `mediahub.local`.

---

## Priority Summary

| Priority | Phase                           | Status     |
| -------- | ------------------------------- | ---------- |
| 🔴 High  | Phase 1: Security & Stability   | **Next Up** |
| 🔴 High  | Phase 2: Media CRUD & Browsing  | Not Started |
| 🟡 Medium | Phase 3: Progress Tracking     | Not Started |
| 🟡 Medium | Phase 4: Streaming             | Not Started |
| 🟡 Medium | Phase 5: UI Enhancements       | Not Started |
| 🟢 Low   | Phase 6: DevOps                 | Not Started |
| 🟢 Low   | Phase 7: Advanced Features      | Not Started |

---

*Last updated: April 2026*
