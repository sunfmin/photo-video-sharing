# Developer Quickstart Guide

**Feature**: Photo Video Sharing System  
**Branch**: `001-photo-video-sharing`  
**Last Updated**: Friday Nov 28, 2025

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Development Environment](#development-environment)
4. [Running Tests](#running-tests)
5. [Running the Application](#running-the-application)
6. [Development Workflow](#development-workflow)
7. [Common Tasks](#common-tasks)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

- **Go**: Version 1.25 or later ([download](https://golang.org/dl/))
- **Docker**: For PostgreSQL and MinIO containers ([download](https://www.docker.com/get-started))
- **Docker Compose**: Usually included with Docker Desktop
- **Protocol Buffers Compiler**: For generating Go code from .proto files
- **Git**: Version control

### System Requirements

- **OS**: macOS, Linux, or Windows (with WSL2)
- **RAM**: Minimum 4GB (8GB recommended for running all containers)
- **Disk**: Minimum 10GB free space (for Go modules, Docker images, media storage)

### Installing Protobuf Compiler

**macOS**:
```bash
brew install protobuf
```

**Linux** (Ubuntu/Debian):
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

**Windows** (via Chocolatey):
```bash
choco install protoc
```

**Verify installation**:
```bash
protoc --version  # Should show libprotoc 3.20.0 or later
```

### Installing Go Protobuf Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/envoyproxy/protoc-gen-validate@latest

# Verify plugins are in PATH
which protoc-gen-go
which protoc-gen-validate
```

---

## Initial Setup

### 1. Clone Repository

```bash
git clone https://github.com/yourorg/photo-video-sharing.git
cd photo-video-sharing
git checkout 001-photo-video-sharing
```

### 2. Install Go Dependencies

```bash
go mod download
go mod tidy
```

**Expected modules**:
- `gorm.io/gorm` - ORM for database operations
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/minio/minio-go/v7` - Object storage client
- `github.com/disintegration/imaging` - Image processing
- `github.com/rwcarlsen/goexif` - EXIF extraction
- `github.com/u2takey/ffmpeg-go` - Video processing
- `golang.org/x/crypto` - Password hashing
- `google.golang.org/protobuf` - Protobuf runtime
- `github.com/testcontainers/testcontainers-go` - Integration testing
- `github.com/google/go-cmp` - Test assertions

### 3. Generate Protobuf Code

```bash
# From repository root
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --validate_out="lang=go:." \
  --validate_opt=paths=source_relative \
  api/proto/v1/*.proto
```

This generates:
- `api/gen/v1/*.pb.go` - Protobuf structs
- `api/gen/v1/*.pb.validate.go` - Validation functions

**Add to Makefile for convenience**:
```makefile
.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --validate_out="lang=go:." --validate_opt=paths=source_relative \
	       api/proto/v1/*.proto
```

Then run: `make proto`

### 4. Set Up Environment Variables

Create `.env` file in repository root:

```bash
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/photovideo?sslmode=disable

# MinIO Object Storage
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=media
MINIO_USE_SSL=false

# Application
PORT=8080
SESSION_SECRET=your-secret-key-change-in-production-min-32-chars
STORAGE_QUOTA_DEFAULT=524288000  # 500MB in bytes

# Development
ENV=development
LOG_LEVEL=debug
```

**Important**: Never commit `.env` to version control. Add to `.gitignore`:
```bash
echo ".env" >> .gitignore
```

---

## Development Environment

### Start Infrastructure with Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: photovideo-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: photovideo
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    container_name: photovideo-minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"  # API
      - "9001:9001"  # Web console
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

volumes:
  postgres_data:
  minio_data:
```

**Start services**:
```bash
docker-compose up -d

# Verify services are running
docker-compose ps

# View logs
docker-compose logs -f
```

**Access MinIO Console**: http://localhost:9001 (minioadmin / minioadmin)

### Create MinIO Bucket

```bash
# Install MinIO client (mc)
brew install minio/stable/mc  # macOS
# or download from https://min.io/docs/minio/linux/reference/minio-mc.html

# Configure MinIO client
mc alias set local http://localhost:9000 minioadmin minioadmin

# Create bucket
mc mb local/media

# Verify bucket exists
mc ls local
```

### Initialize Database Schema

**Option 1**: Run migrations manually (using GORM AutoMigrate):

```bash
go run cmd/api/main.go migrate
```

**Option 2**: Use `psql` to connect and verify:

```bash
# Connect to database
docker exec -it photovideo-postgres psql -U postgres -d photovideo

# List tables (after running migrations)
\dt

# Exit
\q
```

---

## Running Tests

### Prerequisites for Tests

- Docker running (for testcontainers)
- No need to start docker-compose services (tests create isolated containers)

### Run All Tests

```bash
# Full test suite
go test -v ./...

# With race detection
go test -v -race ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # View coverage report
```

### Run Specific Tests

```bash
# Single package
go test -v ./services

# Single test function
go test -v ./handlers -run TestMediaHandler_Upload

# Specific test case (table-driven)
go test -v ./handlers -run TestMediaHandler_Upload/US1-AS1
```

### Test Execution Time

Expected timing (on modern laptop):
- **Setup**: ~5-10 seconds (PostgreSQL container start)
- **Per test**: ~10-100ms (database operations)
- **Full suite**: ~1-2 minutes (depends on number of tests)

### Test Best Practices

1. **Always use real database** (no mocking)
2. **Run tests after every code change** (TDD workflow)
3. **Use table-driven design** with descriptive test case names
4. **Map each acceptance scenario** (US#-AS#) to test case
5. **Test through ServeHTTP** (root mux, not individual handlers)
6. **Derive expected values from fixtures** (not response)

---

## Running the Application

### Development Mode

```bash
# Ensure infrastructure is running
docker-compose up -d

# Run API server
go run cmd/api/main.go

# Or build and run
go build -o bin/photovideo cmd/api/main.go
./bin/photovideo
```

**Server starts on**: http://localhost:8080

**Health check**: http://localhost:8080/health

### Watch Mode (Auto-Reload)

Install `air` for hot reload during development:

```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:
```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "tmp/main"
  include_ext = ["go", "proto"]
  exclude_dir = ["tmp", "vendor", "specs"]
  delay = 1000
```

Run with auto-reload:
```bash
air
```

### Production Build

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o bin/photovideo-linux-amd64 \
  cmd/api/main.go

# Build Docker image
docker build -t photovideo:latest .

# Run container
docker run -p 8080:8080 \
  --env-file .env \
  --network host \
  photovideo:latest
```

---

## Development Workflow

### 1. Test-First Development (TDD)

Follow the constitutional TDD workflow:

```bash
# 1. Define API contract (edit .proto files)
vim api/proto/v1/media.proto

# 2. Generate protobuf code
make proto

# 3. Write integration test (verify it FAILS)
vim handlers/media_handler_test.go
go test -v ./handlers -run TestMediaHandler_Upload
# Expected: FAIL (not implemented yet)

# 4. Implement minimal code to pass test
vim handlers/media_handler.go
vim services/media_service.go

# 5. Run test again (verify it PASSES)
go test -v ./handlers -run TestMediaHandler_Upload
# Expected: PASS

# 6. Refactor while keeping tests green
# Run tests after EVERY change
go test -v ./handlers -run TestMediaHandler_Upload

# 7. Run full test suite before commit
go test -v ./...
```

### 2. Adding New Endpoint

**Example**: Add endpoint to get media statistics

```bash
# Step 1: Define protobuf message
cat >> api/proto/v1/media.proto << 'EOF'
message GetMediaStatsRequest {}

message GetMediaStatsResponse {
  int64 total_media = 1;
  int64 total_photos = 2;
  int64 total_videos = 3;
  int64 total_storage_used = 4;
}
EOF

# Step 2: Generate code
make proto

# Step 3: Write test
cat >> handlers/media_handler_test.go << 'EOF'
func TestMediaHandler_GetStats(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    defer truncateTables(db, "media", "users")
    
    // Create test user
    user := createTestUser(db, nil)
    
    // Create test media
    createTestMedia(db, map[string]interface{}{
        "owner_id": user.ID,
        "file_type": "image/jpeg",
        "file_size": 1024000,
    })
    
    service := services.NewMediaService(db).Build()
    mux := handlers.SetupRoutes(service)
    
    req := httptest.NewRequest("GET", "/api/v1/media/stats", nil)
    // Add session cookie...
    rec := httptest.NewRecorder()
    
    mux.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Fatalf("Expected 200, got %d", rec.Code)
    }
    
    var response pb.GetMediaStatsResponse
    json.NewDecoder(rec.Body).Decode(&response)
    
    expected := &pb.GetMediaStatsResponse{
        TotalMedia: 1,
        TotalPhotos: 1,
        TotalVideos: 0,
        TotalStorageUsed: 1024000,
    }
    
    if diff := cmp.Diff(expected, &response, protocmp.Transform()); diff != "" {
        t.Errorf("Mismatch (-want +got):\n%s", diff)
    }
}
EOF

# Step 4: Verify test FAILS
go test -v ./handlers -run TestMediaHandler_GetStats
# Expected: FAIL (handler not implemented)

# Step 5: Implement service method
vim services/media_service.go
# Add GetStats(ctx context.Context, req *pb.GetMediaStatsRequest) (*pb.GetMediaStatsResponse, error)

# Step 6: Implement handler
vim handlers/media_handler.go
# Add GetStats(w http.ResponseWriter, r *http.Request)

# Step 7: Register route
vim handlers/routes.go
# Add: mux.HandleFunc("GET /api/v1/media/stats", handler.GetStats)

# Step 8: Verify test PASSES
go test -v ./handlers -run TestMediaHandler_GetStats
# Expected: PASS
```

---

## Common Tasks

### Reset Local Database

```bash
# Drop and recreate database
docker exec -it photovideo-postgres psql -U postgres -c "DROP DATABASE photovideo;"
docker exec -it photovideo-postgres psql -U postgres -c "CREATE DATABASE photovideo;"

# Re-run migrations
go run cmd/api/main.go migrate
```

### Clear MinIO Storage

```bash
# Remove all objects from bucket
mc rm --recursive --force local/media

# Recreate bucket
mc mb local/media
```

### View Database Records

```bash
# Connect to database
docker exec -it photovideo-postgres psql -U postgres -d photovideo

# Query users
SELECT id, email, storage_used, storage_quota FROM users;

# Query media
SELECT id, filename, file_type, file_size FROM media LIMIT 10;

# Query shares
SELECT m.filename, u.email AS shared_with 
FROM shares s 
JOIN media m ON s.media_id = m.id 
JOIN users u ON s.shared_with_user_id = u.id;
```

### Generate Test Fixtures

```bash
# Run fixture generation script (create if needed)
go run scripts/generate-fixtures.go

# Or use testutil functions in tests
createTestUser(db, map[string]interface{}{
    "email": "test@example.com",
    "password_hash": hashPassword("Password123"),
})
```

### Update Dependencies

```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get -u github.com/minio/minio-go/v7@latest
go mod tidy

# Verify tests still pass
go test -v ./...
```

---

## Troubleshooting

### Docker Containers Won't Start

**Problem**: `docker-compose up` fails with port already in use

**Solution**:
```bash
# Find process using port 5432
lsof -i :5432

# Kill the process
kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "5433:5432"  # Use 5433 instead
```

### Protobuf Generation Fails

**Problem**: `protoc: command not found` or plugin errors

**Solution**:
```bash
# Verify protoc installed
which protoc

# Verify plugins in PATH
echo $GOPATH/bin
ls $GOPATH/bin | grep protoc

# Add GOPATH/bin to PATH if missing
export PATH=$PATH:$(go env GOPATH)/bin

# Reinstall plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
```

### Tests Fail with "Docker not available"

**Problem**: testcontainers cannot connect to Docker

**Solution**:
```bash
# Verify Docker is running
docker ps

# Check Docker socket permissions (Linux)
sudo chmod 666 /var/run/docker.sock

# Set Docker host environment variable
export DOCKER_HOST=unix:///var/run/docker.sock
```

### Database Connection Refused

**Problem**: Application cannot connect to PostgreSQL

**Solution**:
```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection manually
docker exec -it photovideo-postgres psql -U postgres -d photovideo

# Verify DATABASE_URL in .env
cat .env | grep DATABASE_URL
```

### MinIO Connection Fails

**Problem**: `minio.New()` returns connection error

**Solution**:
```bash
# Verify MinIO is running
docker-compose ps minio

# Check MinIO logs
docker-compose logs minio

# Test connection manually
mc ls local

# Verify credentials match .env
cat .env | grep MINIO
```

### File Upload Fails with "Quota Exceeded"

**Problem**: User cannot upload even though quota not reached

**Solution**:
```sql
# Check user's actual quota usage
docker exec -it photovideo-postgres psql -U postgres -d photovideo
SELECT id, email, storage_used, storage_quota FROM users WHERE email = 'test@example.com';

# Reset quota if inconsistent
UPDATE users SET storage_used = 0 WHERE email = 'test@example.com';
```

### Tests Are Slow

**Problem**: Test suite takes >5 minutes to run

**Solutions**:
```bash
# Run tests in parallel (be careful with database state)
go test -v -parallel 4 ./...

# Run only fast tests (tag slow tests with build tag)
go test -v -short ./...

# Optimize table truncation (use TRUNCATE CASCADE)
# Check testutil/db.go truncateTables() function
```

---

## Next Steps

1. **Read the specification**: Review `specs/001-photo-video-sharing/spec.md`
2. **Study the data model**: Review `specs/001-photo-video-sharing/data-model.md`
3. **Review API contracts**: Check `specs/001-photo-video-sharing/contracts/`
4. **Start implementing**: Follow TDD workflow above
5. **Map scenarios to tests**: Ensure every US#-AS# has a test case
6. **Run tests continuously**: After every code change

**Questions?** Check the constitution at `.specify/memory/constitution.md` for architectural guidelines.

**Ready to build!** 🚀

