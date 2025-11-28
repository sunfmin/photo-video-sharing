# Technical Research: Photo Video Sharing System

**Feature**: 001-photo-video-sharing  
**Date**: Friday Nov 28, 2025  
**Purpose**: Resolve technical unknowns identified in plan.md before detailed design

## Research Tasks

### 1. Object Storage Solution

**Decision**: **MinIO for production, local filesystem for development/testing**

**Rationale**:
- MinIO is S3-compatible (standard API, easy migration to AWS S3 if needed)
- Self-hosted option keeps data under our control (privacy concern for user photos/videos)
- Excellent performance for binary objects (optimized for media files)
- Easy Docker deployment (fits containerized architecture)
- Free and open-source (no vendor lock-in)
- Go SDK available (`github.com/minio/minio-go/v7`)

**Implementation Approach**:
- Define `StorageService` interface in `services/storage_service.go`
- Implement MinIO backend in `internal/storage/minio.go`
- Implement local filesystem backend in `internal/storage/local.go` for tests
- Use builder pattern: `NewStorageService().WithMinIO(endpoint, credentials).Build()`
- Store files with UUID-based keys to avoid naming conflicts: `{user_id}/{media_id}.{ext}`
- Generate presigned URLs for secure direct download (avoid proxying through API)

**Alternatives Considered**:
- **AWS S3**: More expensive, vendor lock-in, network latency, privacy concerns
- **Local filesystem only**: Not scalable, difficult to replicate, no object versioning
- **Google Cloud Storage**: Similar to S3, same concerns
- **Cloudflare R2**: Good option but less mature ecosystem than MinIO

**Configuration**:
```go
type StorageConfig struct {
    Backend   string // "minio" or "local"
    Endpoint  string // MinIO endpoint
    AccessKey string
    SecretKey string
    Bucket    string
    UseSSL    bool
    LocalPath string // For local backend
}
```

---

### 2. Image Processing Library

**Decision**: **Go native `image` package + `github.com/disintegration/imaging` for thumbnails, `github.com/rwcarlsen/goexif` for EXIF**

**Rationale**:
- Pure Go solution (no external dependencies like ImageMagick or libvips)
- Easy to deploy in containers (no system libraries required)
- `imaging` library provides high-quality resize algorithms (Lanczos, bilinear)
- `goexif` handles EXIF metadata extraction reliably
- Lower attack surface (no CGo, no shell execution)
- Good performance for thumbnail generation (<500ms for 10MB photos)
- Handles JPG, PNG, GIF, HEIC formats

**Implementation Approach**:
- Create `ProcessingService` interface in `services/processing_service.go`
- Implement image processor in `internal/processing/image.go`
- Generate multiple thumbnail sizes: 150x150 (list view), 800x600 (preview), 1920x1080 (lightbox)
- Extract EXIF data: date taken, camera model, GPS coordinates, orientation
- Handle orientation correction (auto-rotate based on EXIF)
- Process asynchronously after upload (don't block upload response)

**Thumbnail Strategy**:
- Store thumbnails in object storage alongside originals: `{user_id}/{media_id}_thumb_{size}.jpg`
- Generate on upload, cache forever (immutable content)
- Fallback to original if thumbnail generation fails (graceful degradation)

**Alternatives Considered**:
- **ImageMagick**: Powerful but requires system installation, security vulnerabilities history, shell execution risk
- **libvips**: Faster than ImageMagick but CGo dependency, deployment complexity
- **Cloud service (e.g., Cloudinary)**: Easy but adds external dependency, cost, latency, privacy concerns

**Libraries**:
```go
import (
    "github.com/disintegration/imaging" // v1.6.2 - image resize/thumbnail
    "github.com/rwcarlsen/goexif/exif"  // v0.0.0 - EXIF extraction
)
```

---

### 3. Video Processing Library

**Decision**: **FFmpeg via `github.com/u2takey/ffmpeg-go` wrapper for thumbnail extraction, no transcoding initially**

**Rationale**:
- FFmpeg is industry standard for video processing (reliable, battle-tested)
- `ffmpeg-go` provides Go-friendly API (no shell escaping needed)
- Thumbnail extraction is fast (~1-2 seconds per video)
- Transcoding is expensive and complex (defer until proven necessary)
- Users upload videos in modern formats (MP4/MOV already web-compatible)
- FFmpeg widely available in Docker images (easy deployment)

**Implementation Approach**:
- Implement video processor in `internal/processing/video.go`
- Extract single frame at 1-second mark for thumbnail
- Generate same thumbnail sizes as images (consistency)
- Validate video format and duration on upload
- Set maximum video length (10 minutes initially to control storage)
- Store original video as-is (no transcoding)

**Thumbnail Extraction**:
```go
// Extract frame at 1 second mark
ffmpeg.Input(videoPath).
    Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", 30)}). // 30 frames ≈ 1sec
    Output(thumbnailPath, ffmpeg.KwArgs{"vframes": 1, "format": "image2"}).
    OverWriteOutput().Run()
```

**Alternatives Considered**:
- **Cloud transcoding (AWS MediaConvert, Mux)**: Adds cost and latency, not needed for MVP (most videos already in MP4)
- **Go native video library**: None mature enough for production use
- **No video support**: Removes significant feature value (user expectation is photo AND video)

**Future Transcoding Consideration**:
If analytics show users uploading incompatible formats (AVI, MKV), add transcoding to MP4 (H.264/AAC) as background job. Not in initial scope.

**Dependencies**:
- System: FFmpeg binary (install via Dockerfile: `apt-get install ffmpeg`)
- Go library: `github.com/u2takey/ffmpeg-go` v0.5.0

---

### 4. Authentication Method

**Decision**: **Session-based authentication with secure HTTP-only cookies**

**Rationale**:
- Simplest secure approach for web application (no token management complexity)
- HTTP-only cookies prevent XSS attacks (JavaScript cannot access tokens)
- SameSite=Strict prevents CSRF attacks
- Server-side session storage allows instant revocation (logout, security breach)
- No need for mobile app support initially (web-first)
- Standard library `crypto/rand` for session ID generation
- Sessions stored in PostgreSQL (reuse existing database, no Redis needed)

**Implementation Approach**:
- Store session ID in HTTP-only, Secure, SameSite=Strict cookie
- Session table in PostgreSQL: `id` (UUID), `user_id`, `created_at`, `expires_at`, `last_used_at`
- Session lifetime: 7 days, extend on activity (sliding window)
- Middleware extracts session from cookie, validates, attaches user context
- Logout invalidates session in database (immediate effect)
- Password hashing: `golang.org/x/crypto/bcrypt` (cost factor 12)

**Session Storage**:
```go
type Session struct {
    ID        string    `gorm:"primaryKey;type:uuid"`
    UserID    string    `gorm:"type:uuid;not null;index"`
    CreatedAt time.Time
    ExpiresAt time.Time `gorm:"index"`
    LastUsedAt time.Time
}
```

**Alternatives Considered**:
- **JWT tokens**: Stateless but cannot revoke before expiry, vulnerable to XSS if stored in localStorage, more complex refresh token flow
- **OAuth2 (third-party)**: Adds external dependencies (Google, GitHub), users want dedicated accounts for private media
- **Basic Auth**: Not secure without HTTPS, no session management, poor UX (browser dialogs)

**Security Measures**:
- Bcrypt password hashing (automatically salted, configurable cost)
- HTTPS enforcement (redirect HTTP → HTTPS)
- Rate limiting on login endpoint (prevent brute force)
- Session cleanup job (delete expired sessions daily)

**Libraries**:
```go
import (
    "golang.org/x/crypto/bcrypt" // Password hashing
    "crypto/rand"                // Session ID generation
)
```

---

## Best Practices

### File Upload Handling

**Multipart Form Processing**:
- Use `r.ParseMultipartForm(maxMemory)` with 32MB memory limit
- Stream large files to disk/storage (don't buffer entire file in memory)
- Validate file size before processing (reject >500MB videos, >50MB photos)
- Validate MIME type from Content-Type header AND file magic bytes (prevent spoofing)
- Generate UUID for media ID (avoid filename collisions, prevent path traversal)

**Upload Flow**:
1. Parse multipart form
2. Validate file type and size
3. Generate media UUID
4. Stream to temporary file
5. Process in background (thumbnails, EXIF)
6. Upload to object storage
7. Save metadata to PostgreSQL
8. Return response (don't wait for processing)

**Progress Tracking**:
For initial version, client polls for processing status. Future enhancement: WebSockets for real-time progress.

---

### Storage Quota Management

**Implementation**:
- Track usage per user in `users.storage_used` column (bytes)
- Update on upload (+file size) and delete (-file size)
- Use database transaction to ensure atomic quota checks
- Reject uploads if `storage_used + file_size > storage_quota`
- Background job to recalculate quota (audit for inconsistencies)

**Quota Check Transaction**:
```go
tx := db.Begin()
var user User
tx.Where("id = ?", userID).First(&user)
if user.StorageUsed + fileSize > user.StorageQuota {
    tx.Rollback()
    return ErrQuotaExceeded
}
user.StorageUsed += fileSize
tx.Save(&user)
tx.Create(&media)
tx.Commit()
```

---

### Media Sharing Permissions

**Access Control Pattern**:
- Each media item has `owner_id` (foreign key to users)
- Share table: `media_id`, `shared_with_user_id`, `created_at`
- Album share table: `album_id`, `shared_with_user_id`, `created_at`
- Check access: `owner_id = current_user OR exists(share where shared_with = current_user)`

**Permission Levels** (initial: view-only):
- Shared users can view but NOT delete or modify
- Future enhancement: Add `permission` column for write/admin access

---

### Performance Optimization

**Database Indexes**:
- `media.owner_id` (filter by user)
- `media.created_at` (sort by date)
- `media.file_type` (filter photos vs videos)
- `album_media.album_id` (album contents)
- `shares.media_id, shares.shared_with_user_id` (access checks)
- `sessions.expires_at` (cleanup expired)

**Pagination**:
- Limit gallery queries to 50 items per page
- Use cursor-based pagination for large collections (more efficient than offset)
- Return total count separately (cached, updated on write)

**Caching Strategy** (future optimization, not initial scope):
- Cache user gallery page 1 in Redis (invalidate on upload/delete)
- Cache album metadata (invalidate on album changes)
- Object storage already provides CDN capabilities

---

## Technology Summary

**Final Stack**:
- **Language**: Go 1.25+
- **Database**: PostgreSQL 15+ (metadata, users, sessions)
- **Object Storage**: MinIO (production), local filesystem (dev/test)
- **Image Processing**: `github.com/disintegration/imaging`, `github.com/rwcarlsen/goexif`
- **Video Processing**: FFmpeg via `github.com/u2takey/ffmpeg-go`
- **Authentication**: Session-based with bcrypt password hashing
- **HTTP Framework**: Standard library `net/http` with `http.ServeMux`
- **ORM**: GORM with PostgreSQL driver
- **Protobuf**: protoc, protoc-gen-go, protoc-gen-validate
- **Testing**: testcontainers-go (PostgreSQL), mock storage backend
- **Tracing**: OpenTracing (NoopTracer for dev, configurable for production)

**External Dependencies** (go.mod):
```
github.com/minio/minio-go/v7
github.com/disintegration/imaging
github.com/rwcarlsen/goexif
github.com/u2takey/ffmpeg-go
golang.org/x/crypto
gorm.io/gorm
gorm.io/driver/postgres
google.golang.org/protobuf
github.com/envoyproxy/protoc-gen-validate
github.com/opentracing/opentracing-go
github.com/testcontainers/testcontainers-go
github.com/google/go-cmp
```

**System Dependencies** (Dockerfile):
```dockerfile
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ffmpeg
```

---

## Deployment Considerations

**Container Architecture**:
- Single API container (Go binary + FFmpeg)
- PostgreSQL container (official postgres:15-alpine)
- MinIO container (official minio/minio)
- Docker Compose for local development
- Kubernetes manifest for production (future)

**Environment Variables**:
```bash
DATABASE_URL=postgres://user:pass@host:5432/dbname
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=...
MINIO_SECRET_KEY=...
MINIO_BUCKET=media
SESSION_SECRET=...
STORAGE_QUOTA_DEFAULT=524288000  # 500MB in bytes
```

**Health Checks**:
- `/health` endpoint checks database connectivity, storage service
- Kubernetes liveness/readiness probes

---

## Next Steps

With all technical unknowns resolved, proceed to **Phase 1**:
1. Generate `data-model.md` (entity relationships, GORM models)
2. Generate API contracts in `contracts/` (OpenAPI spec from protobuf)
3. Generate `quickstart.md` (setup instructions for developers)
4. Update agent context with new technology decisions

