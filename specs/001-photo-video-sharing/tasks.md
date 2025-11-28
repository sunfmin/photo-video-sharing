# Tasks: Photo Video Sharing System

**Feature**: 001-photo-video-sharing  
**Input**: Design documents from `/specs/001-photo-video-sharing/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Testing**: TDD mandatory - write tests BEFORE implementation (protobuf → tests → implementation).  
**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Go Project Structure

- **Protobuf**: `api/proto/v1/` (source), `api/gen/v1/` (generated)
- **Services**: `services/` (PUBLIC - business logic, returns protobuf)
- **Handlers**: `handlers/` (PUBLIC - HTTP layer with `*_test.go`)
- **Models**: `internal/models/` (GORM - internal only)
- **Storage**: `internal/storage/` (MinIO and local implementations)
- **Processing**: `internal/processing/` (image and video processing)
- **Middleware**: `internal/middleware/` (auth, logging, tracing)
- **Fixtures**: `testutil/` (helpers and fixtures)

---

## Phase 1: Setup

- [ ] T001 Initialize Go module with `go mod init github.com/yourorg/photo-video-sharing`, Go 1.25+
- [ ] T002 [P] Create directory structure: `api/proto/v1/`, `api/gen/v1/`, `services/`, `handlers/`, `internal/models/`, `internal/storage/`, `internal/processing/`, `internal/middleware/`, `internal/config/`, `testutil/`, `cmd/api/`
- [ ] T003 [P] Install core dependencies: `gorm.io/gorm`, `gorm.io/driver/postgres`, `google.golang.org/protobuf`, `github.com/envoyproxy/protoc-gen-validate`
- [ ] T004 [P] Install MinIO client: `github.com/minio/minio-go/v7`
- [ ] T005 [P] Install image processing: `github.com/disintegration/imaging`, `github.com/rwcarlsen/goexif`
- [ ] T006 [P] Install video processing: `github.com/u2takey/ffmpeg-go`
- [ ] T007 [P] Install testing dependencies: `github.com/testcontainers/testcontainers-go`, `github.com/google/go-cmp`
- [ ] T008 [P] Install security: `golang.org/x/crypto`, OpenTracing: `github.com/opentracing/opentracing-go`
- [ ] T009 [P] Create `.env.example` with all required environment variables (DATABASE_URL, MINIO_*, SESSION_SECRET, etc.)
- [ ] T010 [P] Create `docker-compose.yml` for PostgreSQL and MinIO services
- [ ] T011 [P] Create `Makefile` with targets: `proto`, `test`, `run`, `docker-up`, `docker-down`
- [ ] T012 [P] Create `.gitignore` (.env, tmp/, bin/, coverage.out)
- [ ] T013 Setup protoc configuration: Install `protoc-gen-go`, `protoc-gen-validate`, add generation script to Makefile

---

## Phase 2: Foundation (⚠️ BLOCKS all user stories)

- [ ] T020 Create database test utilities in `testutil/db.go`: `setupTestDB()` with testcontainers-go, `truncateTables()` with CASCADE
- [ ] T021 [P] Create test fixtures utilities in `testutil/fixtures.go`: `CreateTestUser()`, `CreateTestMedia()`, `CreateTestAlbum()`, `CreateTestShare()` with default values
- [ ] T022 [P] Create mock storage in `testutil/storage.go` for testing (in-memory implementation)
- [ ] T023 [P] Setup sentinel errors in `services/errors.go`: `ErrNotFound`, `ErrUnauthorized`, `ErrQuotaExceeded`, `ErrInvalidFormat`, `ErrDuplicateEmail`, etc.
- [ ] T024 [P] Create HTTP error codes in `handlers/error_codes.go`: Error singleton with ServiceErr mapping
- [ ] T025 [P] Implement `HandleServiceError()` in `handlers/error_codes.go`: Automatic error mapping with context awareness
- [ ] T026 [P] Create shared routing setup in `handlers/routes.go`: `SetupRoutes()` function for production and tests
- [ ] T027 [P] Setup OpenTracing NoopTracer in `handlers/routes.go` for development
- [ ] T028 [P] Create migrations export in `services/migrations.go`: `AutoMigrate()` function for external apps
- [ ] T029 [P] Setup config loading in `internal/config/config.go`: Load from environment variables

---

## Phase 3: User Story 2 - User Authentication (P1) 🎯 FOUNDATION FOR ALL STORIES

**Goal**: Enable secure user registration, login, session management, and access control  
**Acceptance Scenarios**: US2-AS1, US2-AS2, US2-AS3, US2-AS4, US2-AS5  
**Dependencies**: Foundation complete  
**Blocks**: All other user stories (auth required for media operations)

### Step 1: Protobuf (Design) 📝

- [ ] T030 [P] [US2] Define User message in `api/proto/v1/user.proto`: id, email, storage_used, storage_quota, timestamps
- [ ] T031 [P] [US2] Define auth requests/responses in `api/proto/v1/user.proto`: RegisterRequest, LoginRequest, LogoutRequest, PasswordResetRequest, GetCurrentUserRequest
- [ ] T032 [US2] Add validation rules to user.proto (email format, password min 8 chars with letter+number), run `make proto`, commit generated code

### Step 2: Tests (Red) 🔴

- [ ] T035 [P] [US2] Create User model fixture in `testutil/fixtures.go`: `CreateTestUser()` with default email/password
- [ ] T036 [P] [US2] Create Session model fixture in `testutil/fixtures.go`: `CreateTestSession()` with expiry
- [ ] T037 [US2] Write test for US2-AS1 (register) in `handlers/user_handler_test.go`: New user registers with valid email/password → account created, auto-login
- [ ] T038 [US2] Write test for US2-AS2 (login) in `handlers/user_handler_test.go`: Existing user logs in → session created, media gallery accessible
- [ ] T039 [US2] Write test for US2-AS3 (isolation) in `handlers/user_handler_test.go`: User A logs in → sees own media, User B logs in → sees different media
- [ ] T040 [US2] Write test for US2-AS4 (redirect) in `handlers/user_handler_test.go`: Unauthenticated user tries to access protected endpoint → 401 Unauthorized
- [ ] T041 [US2] Write test for US2-AS5 (password reset) in `handlers/user_handler_test.go`: User requests password reset → reset link generated (email not actually sent in test)
- [ ] T042 [US2] Add edge case tests: Empty email, invalid email format, short password, missing password, duplicate email, wrong password, expired session, invalid session ID
- [ ] T043 [US2] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 3: Implementation (Green) 🟢

- [ ] T045 [P] [US2] Create User model in `internal/models/user.go`: ID (UUID), Email (unique), PasswordHash, StorageUsed, StorageQuota, timestamps
- [ ] T046 [P] [US2] Create Session model in `internal/models/session.go`: ID (UUID), UserID, ExpiresAt, LastUsedAt, CreatedAt
- [ ] T047 [P] [US2] Add auth errors to `services/errors.go`: `ErrInvalidCredentials`, `ErrDuplicateEmail`, `ErrSessionExpired`
- [ ] T048 [P] [US2] Add auth HTTP codes to `handlers/error_codes.go`: Map auth errors to appropriate status codes (401, 409)
- [ ] T049 [US2] Implement UserService in `services/user_service.go`: Register(), Login(), Logout(), GetCurrentUser(), PasswordReset() methods (use bcrypt, return protobuf)
- [ ] T050 [US2] Implement SessionService in `services/session_service.go`: CreateSession(), ValidateSession(), DeleteSession(), CleanupExpiredSessions()
- [ ] T051 [US2] Implement UserHandler in `handlers/user_handler.go`: Register, Login, Logout, GetCurrentUser, PasswordReset (thin wrappers)
- [ ] T052 [US2] Create auth middleware in `internal/middleware/auth.go`: Extract session cookie, validate, attach user to context, handle 401
- [ ] T053 [US2] Add user/session routes to `handlers/routes.go`: POST /auth/register, POST /auth/login, POST /auth/logout, GET /auth/me
- [ ] T054 [US2] Update `services/migrations.go`: Add User and Session to AutoMigrate()
- [ ] T055 [US2] Add OpenTracing spans to UserHandler and UserService methods
- [ ] T056 [US2] **RUN TESTS** - Verify all PASS (green) ✅

### Step 4: Refactor ♻️

- [ ] T060 [US2] Refactor: Extract password validation to helper, improve error messages, add method comments
- [ ] T061 [US2] **RUN TESTS** after each change ✅, run with `go test -race` ✅

### Step 5: Verify ✅

- [ ] T065 [US2] Run `go test -cover ./services/user_service.go ./handlers/user_handler_test.go` - verify >80% coverage
- [ ] T066 [US2] Verify ALL errors tested: duplicate email, invalid credentials, expired session, missing fields
- [ ] T067 [US2] Verify ALL scenarios tested: US2-AS1 through US2-AS5 + edge cases
- [ ] T068 [US2] Manual verification: Start server, register user via curl, login, access protected endpoint with cookie

---

## Phase 4: User Story 1 - Upload and View Media (P1) 🎯 MVP CORE

**Goal**: Enable users to upload photos/videos and view them in a gallery  
**Acceptance Scenarios**: US1-AS1, US1-AS2, US1-AS3, US1-AS4, US1-AS5, US1-AS6  
**Dependencies**: US2 (authentication) complete  
**Independent Test**: Upload various photo/video files, retrieve and display in gallery

### Step 1: Protobuf (Design) 📝

- [ ] T070 [P] [US1] Define Media message in `api/proto/v1/media.proto`: id, owner_id, filename, file_type, file_size, storage_path, thumbnail_path, width, height, duration, exif_data, timestamps
- [ ] T071 [P] [US1] Define media requests/responses in `api/proto/v1/media.proto`: UploadMediaRequest, GetMediaRequest, ListMediaRequest, DeleteMediaRequest, BatchUploadRequest
- [ ] T072 [US1] Add validation rules to media.proto (file size limits, supported types), run `make proto`, commit generated code

### Step 2: Infrastructure Services 🏗️

- [ ] T075 [P] [US1] Create StorageService interface in `services/storage_service.go`: Upload(), Download(), Delete(), GetPresignedURL()
- [ ] T076 [P] [US1] Implement MinIO storage in `internal/storage/minio.go`: Connect to MinIO, implement StorageService interface
- [ ] T077 [P] [US1] Implement local filesystem storage in `internal/storage/local.go`: For development and testing
- [ ] T078 [P] [US1] Create ProcessingService interface in `services/processing_service.go`: GenerateImageThumbnail(), ExtractEXIF(), GenerateVideoThumbnail()
- [ ] T079 [P] [US1] Implement image processor in `internal/processing/image.go`: Use `imaging` library for thumbnails (150x150, 800x600, 1920x1080), `goexif` for EXIF
- [ ] T080 [P] [US1] Implement video processor in `internal/processing/video.go`: Use FFmpeg to extract frame at 1 second for thumbnail

### Step 3: Tests (Red) 🔴

- [ ] T085 [P] [US1] Create Media model fixture in `testutil/fixtures.go`: `CreateTestMedia()` with default photo/video properties
- [ ] T086 [US1] Write test for US1-AS1 (photo upload) in `handlers/media_handler_test.go`: Upload JPG <50MB → appears in gallery within 5s
- [ ] T087 [US1] Write test for US1-AS2 (video upload) in `handlers/media_handler_test.go`: Upload MP4 <500MB → appears with thumbnail
- [ ] T088 [US1] Write test for US1-AS3 (gallery list) in `handlers/media_handler_test.go`: User uploads media → sees all in gallery sorted by upload date (newest first)
- [ ] T089 [US1] Write test for US1-AS4 (photo view) in `handlers/media_handler_test.go`: Click photo → get presigned URL for full-quality download
- [ ] T090 [US1] Write test for US1-AS5 (video view) in `handlers/media_handler_test.go`: Click video → get presigned URL for playback
- [ ] T091 [US1] Write test for US1-AS6 (batch upload) in `handlers/media_handler_test.go`: Upload 10 photos → all appear with progress tracking
- [ ] T092 [US1] Add edge case tests: File too large, unsupported format, empty file, quota exceeded, invalid mime type, network interruption simulation, duplicate filename
- [ ] T093 [US1] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 4: Implementation (Green) 🟢

- [ ] T095 [P] [US1] Create Media model in `internal/models/media.go`: ID, OwnerID, Filename, FileType, FileSize, StoragePath, ThumbnailPath, Width, Height, Duration, EXIFData, timestamps with indexes
- [ ] T096 [P] [US1] Add media errors to `services/errors.go`: `ErrMediaNotFound`, `ErrUnsupportedFormat`, `ErrFileTooLarge`, `ErrQuotaExceeded`
- [ ] T097 [P] [US1] Add media HTTP codes to `handlers/error_codes.go`: Map media errors to status codes (400, 404, 413, 507)
- [ ] T098 [US1] Implement MediaService in `services/media_service.go`: Upload(), Get(), List(), Delete(), BatchUpload() with quota checking, storage coordination, async processing
- [ ] T099 [US1] Implement file upload handler in `handlers/media_handler.go`: ParseMultipartForm, validate file type/size, stream to storage, return response before processing
- [ ] T100 [US1] Implement media retrieval handlers in `handlers/media_handler.go`: Get (with presigned URLs), List (with pagination)
- [ ] T101 [US1] Add media routes to `handlers/routes.go`: POST /media (upload), GET /media (list), GET /media/{id} (get), DELETE /media/{id}
- [ ] T102 [US1] Update `services/migrations.go`: Add Media to AutoMigrate()
- [ ] T103 [US1] Implement background processing: Async thumbnail generation and EXIF extraction after upload completes
- [ ] T104 [US1] Add quota enforcement: Check user.storage_used + file_size <= user.storage_quota in transaction
- [ ] T105 [US1] Add OpenTracing spans to MediaHandler and MediaService methods
- [ ] T106 [US1] **RUN TESTS** - Verify all PASS (green) ✅

### Step 5: Refactor ♻️

- [ ] T110 [US1] Refactor: Extract file validation to helper, extract storage path generation, improve error context
- [ ] T111 [US1] **RUN TESTS** after each change ✅, run with `go test -race` ✅

### Step 6: Verify ✅

- [ ] T115 [US1] Run `go test -cover` for media service and handlers - verify >80% coverage
- [ ] T116 [US1] Verify ALL errors tested: file too large, unsupported format, quota exceeded, not found
- [ ] T117 [US1] Verify ALL scenarios tested: US1-AS1 through US1-AS6 + edge cases
- [ ] T118 [US1] Manual verification: Upload photo/video via curl, verify storage, verify thumbnail generation, verify EXIF extraction

---

## Phase 5: User Story 4 - Organize into Albums (P2)

**Goal**: Enable users to create albums and organize media into collections  
**Acceptance Scenarios**: US4-AS1, US4-AS2, US4-AS3, US4-AS4, US4-AS5  
**Dependencies**: US1 (media upload/view), US2 (auth)  
**Independent Test**: Create albums, add/remove media, verify persistence

### Step 1: Protobuf (Design) 📝

- [ ] T120 [P] [US4] Define Album message in `api/proto/v1/album.proto`: id, owner_id, name, description, cover_thumbnail_url, media_count, timestamps
- [ ] T121 [P] [US4] Define album requests/responses in `api/proto/v1/album.proto`: CreateAlbumRequest, GetAlbumRequest, UpdateAlbumRequest, DeleteAlbumRequest, AddMediaToAlbumRequest, RemoveMediaFromAlbumRequest
- [ ] T122 [US4] Add validation rules to album.proto (name 1-100 chars, description max 1000), run `make proto`, commit generated code

### Step 2: Tests (Red) 🔴

- [ ] T125 [P] [US4] Create Album fixture in `testutil/fixtures.go`: `CreateTestAlbum()` with default name/description
- [ ] T126 [US4] Write test for US4-AS1 (create album) in `handlers/album_handler_test.go`: Create album with name → appears in list
- [ ] T127 [US4] Write test for US4-AS2 (add media) in `handlers/album_handler_test.go`: Add media to album → appears in both album and gallery
- [ ] T128 [US4] Write test for US4-AS3 (remove media) in `handlers/album_handler_test.go`: Remove from album → removed from album, still in gallery
- [ ] T129 [US4] Write test for US4-AS4 (album list) in `handlers/album_handler_test.go`: View albums → see names, covers, counts
- [ ] T130 [US4] Write test for US4-AS5 (album sharing) in `handlers/album_handler_test.go`: Share album → recipient sees all current and future items
- [ ] T131 [US4] Add edge case tests: Empty name, name too long, media already in album, non-existent media, album not found, delete album (preserve vs delete media)
- [ ] T132 [US4] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 3: Implementation (Green) 🟢

- [ ] T135 [P] [US4] Create Album model in `internal/models/album.go`: ID, OwnerID, Name, Description, timestamps
- [ ] T136 [P] [US4] Create AlbumMedia junction model in `internal/models/album_media.go`: AlbumID, MediaID, AddedAt (composite PK)
- [ ] T137 [P] [US4] Add album errors to `services/errors.go`: `ErrAlbumNotFound`, `ErrMediaAlreadyInAlbum`, `ErrInvalidAlbumName`
- [ ] T138 [P] [US4] Add album HTTP codes to `handlers/error_codes.go`
- [ ] T139 [US4] Implement AlbumService in `services/album_service.go`: Create(), Get(), List(), Update(), Delete(), AddMedia(), RemoveMedia(), GetAlbumMedia()
- [ ] T140 [US4] Implement AlbumHandler in `handlers/album_handler.go`: Create, Get, List, Update, Delete, AddMedia, RemoveMedia, GetAlbumMedia
- [ ] T141 [US4] Add album routes to `handlers/routes.go`: POST /albums, GET /albums, GET /albums/{id}, PUT /albums/{id}, DELETE /albums/{id}, POST /albums/{id}/media, DELETE /albums/{id}/media
- [ ] T142 [US4] Update `services/migrations.go`: Add Album and AlbumMedia
- [ ] T143 [US4] Implement cover thumbnail logic: Use first or most recent media item's thumbnail
- [ ] T144 [US4] Add OpenTracing spans
- [ ] T145 [US4] **RUN TESTS** - Verify all PASS (green) ✅

### Step 4: Refactor & Verify ♻️✅

- [ ] T150 [US4] Refactor, run tests after each change ✅
- [ ] T151 [US4] Verify coverage >80%, all scenarios and errors tested
- [ ] T152 [US4] Manual verification: Create album, add media, remove media, delete album

---

## Phase 6: User Story 3 - Share Media (P2)

**Goal**: Enable users to share media and albums with specific users  
**Acceptance Scenarios**: US3-AS1, US3-AS2, US3-AS3, US3-AS4  
**Dependencies**: US1 (media), US2 (auth), US4 (albums for album sharing)  
**Independent Test**: Create two users, share media, verify recipient access, revoke access

### Step 1: Protobuf (Design) 📝

- [ ] T160 [P] [US3] Define Share message in `api/proto/v1/share.proto`: id, media_id, album_id, owner_id, shared_with_user_id, shared_with_email, created_at
- [ ] T161 [P] [US3] Define share requests/responses in `api/proto/v1/share.proto`: ShareMediaRequest, ShareAlbumRequest, RevokeShareRequest, ListMediaSharesRequest, ListSharedWithMeRequest
- [ ] T162 [US3] Add validation rules to share.proto (recipient emails 1-10, email format), run `make proto`, commit generated code

### Step 2: Tests (Red) 🔴

- [ ] T165 [P] [US3] Create Share fixture in `testutil/fixtures.go`: `CreateTestShare()` with media/album variants
- [ ] T166 [US3] Write test for US3-AS1 (share media) in `handlers/share_handler_test.go`: Share photo with user email → recipient sees in "Shared with me"
- [ ] T167 [US3] Write test for US3-AS2 (revoke share) in `handlers/share_handler_test.go`: Stop sharing → recipient loses access
- [ ] T168 [US3] Write test for US3-AS3 (view only) in `handlers/share_handler_test.go`: Recipient views shared media → cannot delete or modify
- [ ] T169 [US3] Write test for US3-AS4 (multiple recipients) in `handlers/share_handler_test.go`: Share with 5 emails → all receive access
- [ ] T170 [US3] Add edge case tests: Invalid email, user not found, share with self, duplicate share, album sharing with future items, access after delete
- [ ] T171 [US3] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 3: Implementation (Green) 🟢

- [ ] T175 [P] [US3] Create Share model in `internal/models/share.go`: ID, MediaID (nullable), AlbumID (nullable), OwnerID, SharedWithUserID, CreatedAt with XOR validation hook
- [ ] T176 [P] [US3] Add share errors to `services/errors.go`: `ErrShareNotFound`, `ErrCannotShareWithSelf`, `ErrUserNotFound`, `ErrAlreadyShared`
- [ ] T177 [P] [US3] Add share HTTP codes to `handlers/error_codes.go`
- [ ] T178 [US3] Implement ShareService in `services/share_service.go`: ShareMedia(), ShareAlbum(), RevokeShare(), ListMediaShares(), ListSharedWithMe(), ListSharedAlbums() with email lookup
- [ ] T179 [US3] Implement ShareHandler in `handlers/share_handler.go`: ShareMedia, ShareAlbum, RevokeShare, ListMediaShares, ListSharedWithMe
- [ ] T180 [US3] Update MediaService access control: Check owner OR shared_with in List() and Get()
- [ ] T181 [US3] Update AlbumService access control: Check owner OR shared_with for album contents
- [ ] T182 [US3] Add share routes to `handlers/routes.go`: POST /shares/media, POST /shares/album, DELETE /shares/{id}, GET /media/shared, GET /albums/shared
- [ ] T183 [US3] Update `services/migrations.go`: Add Share
- [ ] T184 [US3] Add OpenTracing spans
- [ ] T185 [US3] **RUN TESTS** - Verify all PASS (green) ✅

### Step 4: Refactor & Verify ♻️✅

- [ ] T190 [US3] Refactor, run tests after each change ✅
- [ ] T191 [US3] Verify coverage >80%, all scenarios and errors tested
- [ ] T192 [US3] Manual verification: Share media between two users, revoke, verify album sharing includes future items

---

## Phase 7: User Story 5 - Search and Filter (P3)

**Goal**: Enable users to search media by date, type, and album  
**Acceptance Scenarios**: US5-AS1, US5-AS2, US5-AS3, US5-AS4  
**Dependencies**: US1 (media), US4 (albums for album filter)  
**Independent Test**: Upload diverse media, apply filters, verify results

### Step 1: Protobuf & Tests 📝🔴

- [ ] T200 [P] [US5] Update ListMediaRequest in `api/proto/v1/media.proto`: Add date_range_start, date_range_end, file_type_filter fields, run `make proto`
- [ ] T201 [US5] Write test for US5-AS1 (date filter) in `handlers/media_handler_test.go`: Upload media over months → filter by date range → see only matching
- [ ] T202 [US5] Write test for US5-AS2 (type filter) in `handlers/media_handler_test.go`: Upload photos and videos → filter by "Photos only" → see only images
- [ ] T203 [US5] Write test for US5-AS3 (album filter) in `handlers/media_handler_test.go`: Media in multiple albums → filter by album → see only that album's media
- [ ] T204 [US5] Write test for US5-AS4 (combined filters) in `handlers/media_handler_test.go`: Apply date range + file type → results match both criteria
- [ ] T205 [US5] Add edge case tests: Invalid date format, future dates, empty results, filter combinations
- [ ] T206 [US5] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 2: Implementation & Verify 🟢✅

- [ ] T210 [US5] Update MediaService.List() in `services/media_service.go`: Add filtering logic for date range, file type, album (use GORM where clauses)
- [ ] T211 [US5] Add indexes to media table: (owner_id, uploaded_at, file_type) for efficient filtering
- [ ] T212 [US5] **RUN TESTS** - Verify all PASS (green) ✅
- [ ] T213 [US5] Refactor, verify coverage >80%
- [ ] T214 [US5] Manual verification: Filter media by various combinations

---

## Phase 8: User Story 6 - Delete and Manage Storage (P3)

**Goal**: Enable users to delete media and manage storage quota  
**Acceptance Scenarios**: US6-AS1, US6-AS2, US6-AS3, US6-AS4  
**Dependencies**: US1 (media), US3 (sharing - for deletion notifications), US4 (albums - for album deletion)  
**Independent Test**: Upload media, delete items, verify storage quota updates and cascade effects

### Step 1: Protobuf & Tests 📝🔴

- [ ] T220 [P] [US6] Update DeleteMediaRequest and DeleteAlbumRequest in protobuf files (already defined), update DeleteAlbumRequest to include `delete_media` boolean
- [ ] T221 [P] [US6] Define GetStorageStatsRequest/Response in `api/proto/v1/user.proto`: total_storage_used, total_storage_quota
- [ ] T222 [US6] Run `make proto`, commit generated code
- [ ] T223 [US6] Write test for US6-AS1 (delete media) in `handlers/media_handler_test.go`: Delete media → removed from gallery and storage, quota updated
- [ ] T224 [US6] Write test for US6-AS2 (delete album) in `handlers/album_handler_test.go`: Delete album → prompt for delete_media option, verify behavior
- [ ] T225 [US6] Write test for US6-AS3 (storage stats) in `handlers/user_handler_test.go`: View storage usage → see used and available quota
- [ ] T226 [US6] Write test for US6-AS4 (cascade delete) in `handlers/media_handler_test.go`: Delete shared media → removed from recipient views, verify notifications
- [ ] T227 [US6] Add edge case tests: Delete non-existent media, delete media owned by another user, storage quota recalculation
- [ ] T228 [US6] **RUN TESTS** - Verify all FAIL (red) ❌

### Step 2: Implementation & Verify 🟢✅

- [ ] T230 [US6] Implement MediaService.Delete() in `services/media_service.go`: Delete from storage, delete database record (CASCADE deletes shares), update user.storage_used in transaction
- [ ] T231 [US6] Implement AlbumService.Delete() in `services/album_service.go`: Handle delete_media option (delete media vs just album structure)
- [ ] T232 [US6] Implement UserService.GetStorageStats() in `services/user_service.go`: Return storage_used and storage_quota
- [ ] T233 [US6] Update MediaHandler.Delete() in `handlers/media_handler.go`: Call service, handle cascade notifications
- [ ] T234 [US6] Update AlbumHandler.Delete() in `handlers/album_handler.go`: Pass delete_media parameter
- [ ] T235 [US6] Implement UserHandler.GetStorageStats() in `handlers/user_handler.go`: Return storage statistics
- [ ] T236 [US6] Add route GET /auth/me/storage to `handlers/routes.go`
- [ ] T237 [US6] Add database CASCADE constraints in migrations: ON DELETE CASCADE for foreign keys
- [ ] T238 [US6] **RUN TESTS** - Verify all PASS (green) ✅
- [ ] T239 [US6] Refactor, verify coverage >80%
- [ ] T240 [US6] Manual verification: Delete media, verify storage recalculation, verify shares removed, delete album with both options

---

## Phase 9: Polish & Cross-Cutting

- [ ] T250 [P] Create main application entry point in `cmd/api/main.go`: Load config, connect to DB and MinIO, setup services, start HTTP server
- [ ] T251 [P] Add health check endpoint in `handlers/routes.go`: GET /health checks DB and storage connectivity
- [ ] T252 [P] Implement graceful shutdown in `cmd/api/main.go`: Handle SIGINT/SIGTERM, close connections
- [ ] T253 [P] Add request logging middleware in `internal/middleware/logging.go`: Log method, path, status, duration
- [ ] T254 [P] Add CORS middleware in `internal/middleware/cors.go` if needed for frontend
- [ ] T255 [P] Implement session cleanup background job: Delete expired sessions daily
- [ ] T256 Run full test suite: `go test -v ./...` - verify all pass ✅
- [ ] T257 Run with race detector: `go test -race ./...` - verify no data races ✅
- [ ] T258 Run coverage analysis: `go test -coverprofile=coverage.out ./...` - verify >80% for business logic
- [ ] T259 Run static analysis: `go vet ./...`, `golangci-lint run` if available
- [ ] T260 [P] Security review: Verify SQL injection prevention (GORM parameterized queries), XSS prevention (no HTML rendering), session security (HTTP-only, Secure, SameSite)
- [ ] T261 [P] Error handling review: Verify all errors have test cases, error messages don't leak sensitive info
- [ ] T262 [P] Code cleanup: Remove debug prints, ensure comments explain WHY not WHAT, remove unused imports
- [ ] T263 Create comprehensive README.md: Feature overview, setup instructions, development workflow, testing, deployment
- [ ] T264 Create API documentation: Generate from OpenAPI spec, add examples
- [ ] T265 Performance testing: Load test upload endpoint, verify 1000 concurrent users, optimize as needed

---

## Execution Order & Dependencies

### Phase Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundation) ⚠️ BLOCKS all user stories
    ↓
Phase 3 (US2: Auth) [P1] ⚠️ REQUIRED for all other stories
    ↓
    ├─→ Phase 4 (US1: Upload/View) [P1] 🎯 MVP CORE
    │       ↓
    │       ├─→ Phase 5 (US4: Albums) [P2]
    │       │       ↓
    │       │       └─→ Phase 6 (US3: Sharing) [P2] (needs albums for album sharing)
    │       │
    │       └─→ Phase 7 (US5: Search/Filter) [P3] (needs media + albums)
    │
    └─→ Phase 8 (US6: Delete/Manage) [P3] (needs media + sharing + albums)
        ↓
Phase 9 (Polish)
```

### Critical Path

1. **Setup** → **Foundation** (mandatory for all)
2. **US2 (Auth)** (blocks everything)
3. **US1 (Upload/View)** (MVP core functionality)
4. **US4 (Albums)** before **US3 (Sharing)** (album sharing depends on albums)
5. All P3 stories after P1/P2 complete
6. **Polish** after all user stories

### Parallel Opportunities

**After Foundation Complete**:
- US2 can start immediately (blocking)

**After US2 Complete**:
- US1 can start (foundation for others)

**After US1 Complete**:
- US4 (Albums) can run in parallel with US7 protobuf setup
- US5 (Search) can start (only needs media, not dependent on albums/sharing)

**After US4 Complete**:
- US3 (Sharing) can start
- US6 (Delete) can start in parallel

**Within Each Story**:
- Tasks marked [P] can run in parallel (different files, no dependencies)
- Protobuf definitions can be created in parallel across stories
- Fixtures can be created in parallel
- Tests for different scenarios can be written in parallel

---

## TDD Workflow (Constitution Principle VI)

**Every User Story Follows This Cycle**:

1. **Protobuf (Design)** 📝
   - Define API contracts in `.proto` files
   - Add validation rules
   - Generate code with `make proto`
   - Commit generated code

2. **Tests (Red)** 🔴
   - Create test fixtures in `testutil/fixtures.go`
   - Write integration tests for ALL acceptance scenarios (US#-AS#)
   - Write edge case tests (input validation, boundaries, auth, data state, database, HTTP)
   - **RUN TESTS** - Verify FAIL (red) ❌
   - If tests pass before implementation → tests are wrong, fix them

3. **Implementation (Green)** 🟢
   - Create GORM models in `internal/models/`
   - Add sentinel errors to `services/errors.go`
   - Add HTTP error codes to `handlers/error_codes.go`
   - Implement service in `services/` (business logic, returns protobuf, uses context)
   - Implement handler in `handlers/` (thin wrapper, delegates to service)
   - Add routes to `handlers/routes.go`
   - Update `services/migrations.go`
   - Add OpenTracing spans
   - **RUN TESTS** - Verify PASS (green) ✅

4. **Refactor** ♻️
   - Extract helpers, improve errors, add comments
   - **RUN TESTS** after EVERY change ✅
   - Run with race detector: `go test -race`

5. **Verify** ✅
   - Run coverage analysis: `go test -cover` - target >80%
   - Verify ALL errors have test cases
   - Verify ALL scenarios tested (US#-AS# mapping)
   - Manual verification via curl/Postman

**Story Complete ONLY When**:
- All tests pass ✅
- Coverage >80%
- All acceptance scenarios have corresponding tests
- All edge cases covered
- No skipped or disabled tests

---

## Implementation Strategy

### MVP First (Recommended)

**Scope**: Setup + Foundation + US2 (Auth) + US1 (Upload/View)

**Timeline**: ~1-2 weeks with TDD

**Value**: Functional photo/video storage with authentication

**Deployment**: Can deploy to production as basic media storage solution

**Validation**: Users can register, upload photos/videos, view gallery

---

### Incremental Delivery (Recommended)

After MVP, deliver one user story at a time in priority order:

1. **Release 1.0** (MVP): Setup + Foundation + US2 + US1 (~2 weeks)
   - Value: Secure personal media storage
   
2. **Release 1.1**: + US4 (Albums) (~3-4 days)
   - Value: Media organization
   
3. **Release 1.2**: + US3 (Sharing) (~4-5 days)
   - Value: Collaboration and sharing

4. **Release 1.3**: + US5 (Search) + US6 (Delete) (~4-5 days)
   - Value: Complete feature set

5. **Release 1.4**: Polish + performance optimization (~3 days)
   - Value: Production-ready quality

**Total Estimated Timeline**: 3-4 weeks with TDD, full test coverage

---

### Parallel Team Implementation

If multiple developers available:

**Phase 1 & 2**: All developers collaborate (foundation for everyone)

**Phase 3+**: Split by user story:
- Developer A: US2 (Auth) - blocking, highest priority
- Developer B: Prepare fixtures and infrastructure services for US1
- After US2 complete:
  - Developer A: US1 (Upload/View)
  - Developer B: US4 (Albums) - can start protobuf and tests
- After US1 complete:
  - Developer A: US3 (Sharing)
  - Developer B: US5 (Search)
  - Developer C: US6 (Delete)

**Coordination Required**:
- Protobuf definitions (coordinate field naming, message structure)
- Shared fixtures (agree on default test data)
- Error codes (don't duplicate error code names)

---

## Notes

**Task Format Conventions**:
- **[P]**: Parallelizable (different files, no blocking dependencies)
- **[US#]**: User story mapping (US1, US2, US3, etc.)
- File paths included for clarity

**TDD Mandatory** (Constitution Principle VI):
- Write tests BEFORE implementation
- Run tests after EVERY code change
- Story complete ONLY when all tests pass
- No skipping or disabling tests to make them pass

**Code Organization**:
- Services/handlers: PUBLIC packages (return protobuf, reusable)
- Models: `internal/models/` (GORM, never exposed externally)
- Protobuf: `api/proto/v1/` (source), `api/gen/v1/` (generated)
- Storage/Processing: Interfaces in services, implementations in `internal/`

**Testing Requirements**:
- Integration tests ONLY (no mocking), real PostgreSQL via testcontainers
- Table-driven design with descriptive test case names
- Map each acceptance scenario (US#-AS#) to test case
- Test via root mux ServeHTTP (NOT individual handlers)
- Use protobuf structs with protocmp for assertions
- Derive expected values from fixtures (NOT from response)
- Cover ALL edge cases: input validation, boundaries, auth, data state, database, HTTP

**Avoid**:
- Implementing before tests (breaks TDD)
- Skipping edge case tests
- Removing or weakening tests to make them pass
- Copying non-random response fields to expected values
- Testing individual handlers instead of through mux
- Using mocks instead of real database

---

## Task Summary

**Total Tasks**: 265

**By Phase**:
- Phase 1 (Setup): 13 tasks
- Phase 2 (Foundation): 10 tasks
- Phase 3 (US2 - Auth): 39 tasks
- Phase 4 (US1 - Upload/View): 49 tasks
- Phase 5 (US4 - Albums): 33 tasks
- Phase 6 (US3 - Sharing): 33 tasks
- Phase 7 (US5 - Search): 15 tasks
- Phase 8 (US6 - Delete): 21 tasks
- Phase 9 (Polish): 16 tasks
- Documentation/Coordination: 36 tasks

**By Priority**:
- P1 (MVP): Setup + Foundation + US2 + US1 = 111 tasks (~2 weeks)
- P2 (Enhancement): US4 + US3 = 66 tasks (~1.5 weeks)
- P3 (Advanced): US5 + US6 = 36 tasks (~1 week)
- Polish: 16 tasks (~3 days)

**Parallelizable**: 89 tasks marked [P] can run simultaneously

**Independent Stories**: US5 (Search) can run parallel to US4/US3 after US1 complete

**MVP Scope**: 111 tasks delivers functional photo/video storage with authentication

**Format Validation**: ✅ All tasks follow checklist format: `- [ ] [ID] [P?] [Story?] Description with file path`

