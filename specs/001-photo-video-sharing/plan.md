# Implementation Plan: Photo Video Sharing System

**Branch**: `001-photo-video-sharing` | **Date**: Friday Nov 28, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-photo-video-sharing/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

A photo and video sharing platform that allows users to upload, organize, view, and share media files with others. The system provides secure user authentication, personal media galleries, album organization, selective sharing capabilities, and search/filter functionality. Core technical approach uses Go backend with PostgreSQL for metadata storage, object storage for media files, image/video processing pipeline for thumbnails and transcoding, and RESTful API following constitutional architecture patterns.

## Technical Context

**Stack**: Go 1.25+, PostgreSQL 15+, GORM, Protobuf, OpenTracing, testcontainers-go  
**Project Type**: Web API (single service)  
**Target**: Linux server (containerized)  
**Performance**: Support 1000 concurrent uploads, 5000 reads/sec, <5s upload processing for 10MB photos, <10s thumbnail generation for videos
**Scale**: 10,000 users, 500MB storage quota per user (5TB total), 100K media items initially

**Additional Technology Requirements**:
- **Object Storage**: NEEDS CLARIFICATION (S3-compatible for media files - AWS S3, MinIO, or local filesystem for development)
- **Image Processing**: NEEDS CLARIFICATION (thumbnail generation and EXIF extraction - ImageMagick, libvips, or Go native libraries)
- **Video Processing**: NEEDS CLARIFICATION (thumbnail extraction and optional transcoding - FFmpeg or cloud service)
- **File Upload**: Multipart form handling, chunked uploads for large files, progress tracking
- **Authentication**: NEEDS CLARIFICATION (session-based, JWT, or OAuth2 - default to secure session cookies)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Architecture Compliance**:
- ✅ **Single Go API**: Follows Option 1 structure from constitution (services/, handlers/, internal/, api/proto/)
- ✅ **Service Layer**: Business logic in public `services/` package returning protobuf structs
- ✅ **Handler Layer**: Thin HTTP handlers in public `handlers/` package delegating to services
- ✅ **Models**: GORM models in `internal/models/` (never exposed externally)
- ✅ **Protobuf**: API contracts defined in `api/proto/v1/`, generated code in `api/gen/v1/`
- ✅ **Builder Pattern**: Services use `NewMediaService(db).WithStorage(store).Build()` pattern
- ✅ **Reusability**: Public packages allow external apps to import as library

**Testing Compliance**:
- ✅ **Integration Tests Only**: Real PostgreSQL via testcontainers, no mocking
- ✅ **Table-Driven**: All tests use table-driven design with descriptive `name` fields
- ✅ **Edge Cases**: Cover input validation, boundaries, auth, data state, database, HTTP errors
- ✅ **ServeHTTP**: Test via root mux (not individual handlers), shared routing configuration
- ✅ **Protobuf Structs**: Use protobuf with protocmp for assertions
- ✅ **Fixture-Derived**: Expected values from fixtures (testutil/fixtures.go), not response copies
- ✅ **Scenario Mapping**: Each acceptance scenario (US#-AS#) maps to test case
- ✅ **TDD Workflow**: Write tests first, verify fail, implement, verify pass

**Error Handling Compliance**:
- ✅ **Sentinel Errors**: Service layer defines `var ErrMediaNotFound`, `ErrQuotaExceeded`, etc.
- ✅ **Error Wrapping**: Use `fmt.Errorf("%w")` for context breadcrumbs
- ✅ **HTTP Mapping**: Singleton error codes with automatic ServiceErr mapping
- ✅ **Test Coverage**: ALL errors must have test cases

**Context & Tracing Compliance**:
- ✅ **Context-Aware**: Services accept `context.Context`, handlers use `r.Context()`, database uses `db.WithContext(ctx)`
- ✅ **Distributed Tracing**: HTTP endpoints and services create OpenTracing spans
- ✅ **Cancellation**: Long operations (uploads, processing) check context cancellation

**Gate Status**: ✅ **PASSED** - No constitutional violations. Standard single-service API architecture.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., cmd/api, cmd/worker). The delivered plan must not include
  Option labels.
-->

```text
api/
├── proto/v1/               # Protobuf definitions (.proto files)
│   ├── media.proto        # Media upload/retrieval messages
│   ├── user.proto         # User authentication messages
│   ├── album.proto        # Album organization messages
│   └── share.proto        # Sharing messages
└── gen/v1/                 # Protobuf generated code (PUBLIC - importable)
    ├── media.pb.go
    ├── media.pb.validate.go
    ├── user.pb.go
    ├── user.pb.validate.go
    ├── album.pb.go
    ├── album.pb.validate.go
    ├── share.pb.go
    └── share.pb.validate.go

services/                   # PUBLIC package - business logic (reusable)
├── media_service.go       # Upload, view, delete media operations
├── user_service.go        # Authentication, account management
├── album_service.go       # Album creation, organization
├── share_service.go       # Sharing and permissions
├── storage_service.go     # Object storage abstraction
├── processing_service.go  # Image/video processing (thumbnails, EXIF)
├── errors.go              # Sentinel errors (ErrMediaNotFound, ErrQuotaExceeded, etc.)
└── migrations.go          # AutoMigrate() function for external apps

handlers/                   # PUBLIC package - HTTP handlers (reusable)
├── media_handler.go       # Media upload/download endpoints
├── media_handler_test.go
├── user_handler.go        # Authentication endpoints
├── user_handler_test.go
├── album_handler.go       # Album management endpoints
├── album_handler_test.go
├── share_handler.go       # Sharing endpoints
├── share_handler_test.go
├── error_codes.go         # HTTP error code singleton with ServiceErr mapping
└── routes.go              # Shared routing configuration

internal/                   # INTERNAL - implementation details only
├── models/                # GORM models (internal - services return protobuf)
│   ├── user.go
│   ├── media.go
│   ├── album.go
│   └── share.go
├── storage/               # Storage backend implementations
│   ├── s3.go             # S3-compatible storage
│   ├── local.go          # Local filesystem (development)
│   └── interface.go      # Storage interface
├── processing/            # Image/video processing implementations
│   ├── image.go          # Image thumbnail & EXIF
│   └── video.go          # Video thumbnail extraction
├── middleware/            # App-specific middleware
│   ├── auth.go           # Authentication middleware
│   ├── logging.go
│   └── tracing.go
└── config/                # Configuration loading
    └── config.go

cmd/
└── api/                   # Main application entry point
    └── main.go

testutil/                   # Test helpers and fixtures
├── fixtures.go            # CreateTestUser(), CreateTestMedia(), etc.
├── db.go                  # setupTestDB() with testcontainers
└── storage.go             # Mock storage for testing
```

**Structure Decision**: Selected **Option 1: Single Go API** - this is a monolithic web service with multiple related domains (media, users, albums, sharing). All functionality is cohesive and shares the same database schema, making a single service the appropriate choice. Microservices would add unnecessary complexity for this feature scope.

Key architectural decisions:
- **Services Layer**: Public package with domain-specific services (media, user, album, share) plus infrastructure services (storage, processing)
- **Storage Abstraction**: `internal/storage/` provides interface with multiple backends (S3, local filesystem) for flexibility
- **Processing Abstraction**: `internal/processing/` handles image/video operations, keeping media-specific logic isolated
- **Models**: Internal GORM models map to database schema, never exposed outside service layer
- **Protobuf API**: Clean contract-first design with separate proto files per domain

**Architecture** (Constitution Principle X):
- Services/handlers: PUBLIC packages (return protobuf, reusable by external apps)
- Models: `internal/models/` (GORM only, never exposed)
- Protobuf: PUBLIC `api/gen/v1/` (external apps import these types)
- `AutoMigrate()`: Exported in `services/migrations.go` for external schema setup
- Storage/Processing: Interfaces in services, implementations in `internal/` (pluggable backends)

**Data Flow**:
```
HTTP Request → Handler (parse/validate) → Service (business logic, protobuf) → GORM Model → PostgreSQL
                                       ↓
                                  Storage Service → Object Storage (S3/local)
                                       ↓
                                  Processing Service → Image/Video processing
```

## Testing Strategy

### Test-First Development (TDD)

TDD workflow (Constitution Development Workflow):

1. **Design**: Define API in `.proto` files → generate code
2. **Red**: Write integration tests → verify FAIL
3. **Green**: Implement → run tests → verify PASS
4. **Refactor**: Improve code → run tests after each change
5. **Complete**: Done only when ALL tests pass

### Integration Testing Requirements

Constitution Testing Principles I-IX:

- **Integration tests ONLY** (NO mocking), real PostgreSQL via testcontainers
- **Table-driven** with `name` fields
- **Edge cases MANDATORY**: Input validation, boundaries, auth, data state, database, HTTP
- **ServeHTTP testing** via root mux (NOT individual handlers)
- **Protobuf** structs with `protocmp` assertions
- **Derive from fixtures** (NOT response, except UUIDs/timestamps)
- **Run tests** after EVERY change (Principle VI)
- **Map scenarios** to tests (US#-AS#, Principle VIII)
- **Coverage >80%** (Principle IX)

### Error Handling Strategy

Constitution Principle XI:

- **Service**: Sentinel errors (`var ErrXxx`), wrap with `fmt.Errorf("%w")`
- **HTTP**: Error singleton with `ServiceErr` mapping, automatic via `HandleServiceError()`
- **Testing**: ALL errors must have test cases

### Test Database Isolation

- **testcontainers-go** with PostgreSQL (Docker required)
- **Truncation**: `defer truncateTables(db, "tables...")` with CASCADE
- **Parallel**: `t.Parallel()` safe

### Context-Aware Operations

Constitution Principle XII: Services accept `context.Context`, handlers use `r.Context()`, database uses `db.WithContext(ctx)`, tests verify cancellation.

### Distributed Tracing

Constitution Principle XI: HTTP endpoints create OpenTracing spans, services create child spans, database as ONE span per transaction (NOT per query). Dev uses `NoopTracer{}`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitutional violations - this section is not applicable.

---

## Phase 1 Completion Summary

**Status**: ✅ **COMPLETE** (Friday Nov 28, 2025)

### Deliverables Created

1. **research.md** ✅
   - Object storage solution: MinIO (production), local filesystem (dev/test)
   - Image processing: Go native with `imaging` and `goexif` libraries
   - Video processing: FFmpeg via `ffmpeg-go` wrapper
   - Authentication: Session-based with bcrypt password hashing
   - All technical unknowns resolved

2. **data-model.md** ✅
   - 7 entities defined: User, Media, Album, AlbumMedia, Share, AlbumShare, Session
   - GORM model examples with proper relationships
   - Database constraints and indexes specified
   - Query patterns and state transitions documented

3. **contracts/** ✅
   - `user.proto` - Authentication and user management (6 messages)
   - `media.proto` - Media operations (8 messages)
   - `album.proto` - Album management (10 messages)
   - `share.proto` - Sharing operations (8 messages)
   - `api-spec.yaml` - OpenAPI 3.0 specification (complete REST API)
   - `README.md` - Contract documentation and usage guide

4. **quickstart.md** ✅
   - Complete developer setup guide
   - Prerequisites and installation instructions
   - Development workflow (TDD)
   - Common tasks and troubleshooting
   - Docker Compose configuration for local environment

5. **Agent Context** ✅
   - Updated `.cursor/rules/specify-rules.mdc` with technology stack
   - Project type: Web API (single service)
   - Technology decisions integrated

### Post-Design Constitution Check

**Architecture Verification**:
- ✅ **Package Structure**: Follows Option 1 (single Go API) exactly
  - `services/` - PUBLIC (media, user, album, share, storage, processing services)
  - `handlers/` - PUBLIC (HTTP handlers with shared routing)
  - `internal/models/` - INTERNAL (GORM models never exposed)
  - `api/proto/v1/` - PUBLIC (protobuf contracts)
  - `api/gen/v1/` - PUBLIC (generated protobuf code)
  
- ✅ **Service Interfaces**: All services return protobuf structs, accept context
- ✅ **Builder Pattern**: Services use `NewXxxService(db).WithYyy().Build()`
- ✅ **Storage Abstraction**: Interface in services, implementations in internal/
- ✅ **Processing Abstraction**: Interface pattern for pluggable image/video processing

**Testing Design**:
- ✅ **Integration Tests**: Real PostgreSQL via testcontainers (no mocking)
- ✅ **Fixtures**: `testutil/fixtures.go` with `CreateTestXxx()` functions
- ✅ **Scenario Mapping**: All 25 acceptance scenarios (US#-AS#) will map to tests
- ✅ **Edge Cases**: Comprehensive coverage planned across all domains

**Error Handling Design**:
- ✅ **Sentinel Errors**: Defined in `services/errors.go` (ErrMediaNotFound, ErrQuotaExceeded, etc.)
- ✅ **HTTP Mapping**: Singleton in `handlers/error_codes.go` with ServiceErr field
- ✅ **Automatic Mapping**: `HandleServiceError()` function for DRY error responses

**Technology Compliance**:
- ✅ **Go 1.25+**: Standard library HTTP, no frameworks
- ✅ **PostgreSQL 15+**: GORM for ORM, testcontainers for testing
- ✅ **Protobuf**: Contract-first design, protoc-gen-validate for validation
- ✅ **OpenTracing**: Distributed tracing at appropriate granularity

**Final Gate Status**: ✅ **PASSED** - Design fully complies with constitution. No violations. Ready for implementation.

### Metrics

- **Entities**: 7 (User, Media, Album, AlbumMedia, Share, AlbumShare, Session)
- **Protobuf Messages**: 32 (across 4 proto files)
- **API Endpoints**: 24 (authentication, media, albums, sharing)
- **Functional Requirements**: 43 (38 core + 5 error handling)
- **Acceptance Scenarios**: 25 (mapped from 6 user stories)
- **Test Coverage Target**: >80% for business logic

### Technology Stack (Final)

**Core**:
- Go 1.25+, PostgreSQL 15+, GORM, Protobuf, OpenTracing

**Infrastructure**:
- MinIO (object storage), FFmpeg (video processing), Docker, Docker Compose

**Libraries**:
- `github.com/minio/minio-go/v7` - Object storage client
- `github.com/disintegration/imaging` - Image resize/thumbnails
- `github.com/rwcarlsen/goexif` - EXIF extraction
- `github.com/u2takey/ffmpeg-go` - Video processing wrapper
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/testcontainers/testcontainers-go` - Integration testing

**Development**:
- `protoc` - Protobuf compiler
- `protoc-gen-go`, `protoc-gen-validate` - Go code generation
- `github.com/google/go-cmp` with `protocmp` - Test assertions

### Next Phase: Implementation (Phase 2)

**Ready for**: `/speckit.tasks` to break down implementation into concrete tasks

**Implementation Order** (by priority):
1. **P1**: Core upload/view + authentication (US1, US2)
2. **P2**: Sharing + albums (US3, US4)
3. **P3**: Search/filtering + deletion (US5, US6)

**Expected Timeline**:
- Phase 2 (Tasks): ~30 minutes (task breakdown)
- Phase 3 (Implementation): ~2-3 weeks (TDD with all tests)

---

## References

- **Specification**: [spec.md](./spec.md) - Feature requirements and user scenarios
- **Research**: [research.md](./research.md) - Technical decisions and rationale
- **Data Model**: [data-model.md](./data-model.md) - Database schema and entities
- **Contracts**: [contracts/](./contracts/) - Protobuf and OpenAPI specifications
- **Quickstart**: [quickstart.md](./quickstart.md) - Developer setup guide
- **Constitution**: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md) - Architectural principles
