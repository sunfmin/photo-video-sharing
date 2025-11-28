# API Contracts

**Feature**: Photo Video Sharing System  
**Version**: 1.0.0  
**Protocol**: REST over HTTP/HTTPS with JSON payloads (except file uploads)

## Overview

This directory contains the API contract definitions for the photo video sharing system:

- **Protobuf definitions** (`.proto` files): Source of truth for API contracts
- **OpenAPI specification** (`api-spec.yaml`): REST API documentation

All services and handlers must implement these contracts exactly as specified.

## Files

- `user.proto` - Authentication and user management messages
- `media.proto` - Media upload, retrieval, and management messages
- `album.proto` - Album organization messages
- `share.proto` - Sharing and permissions messages
- `api-spec.yaml` - Complete OpenAPI 3.0 specification for REST endpoints

## Authentication

**Method**: Session-based authentication with HTTP-only cookies

**Flow**:
1. User registers (`POST /api/v1/auth/register`) or logs in (`POST /api/v1/auth/login`)
2. Server creates session, returns `Set-Cookie` header with `session_id`
3. Client includes cookie in all subsequent requests automatically
4. Middleware validates session and attaches user context to requests
5. User logs out (`POST /api/v1/auth/logout`) to invalidate session

**Cookie Attributes**:
- `HttpOnly`: Prevents JavaScript access (XSS protection)
- `Secure`: Requires HTTPS (production only)
- `SameSite=Strict`: Prevents CSRF attacks
- Max-Age: 7 days (604800 seconds)

## API Conventions

### Endpoints

All endpoints are prefixed with `/api/v1/`:

- **Authentication**: `/api/v1/auth/*`
- **Media**: `/api/v1/media/*`
- **Albums**: `/api/v1/albums/*`
- **Sharing**: `/api/v1/shares/*`

### HTTP Methods

- `GET` - Retrieve resource(s)
- `POST` - Create new resource or perform action
- `PUT` - Update existing resource (full replacement)
- `PATCH` - Partial update (not used initially)
- `DELETE` - Remove resource

### Request/Response Format

**Content-Type**: `application/json` (except multipart uploads)

**Request Body**: JSON-encoded protobuf message
**Response Body**: JSON-encoded protobuf message

**Example**:
```json
POST /api/v1/albums
Content-Type: application/json

{
  "name": "Vacation 2025",
  "description": "Photos from our summer trip"
}
```

### File Uploads

Media uploads use `multipart/form-data`:

```
POST /api/v1/media
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="photo.jpg"
Content-Type: image/jpeg

[binary data]
------WebKitFormBoundary--
```

**Supported Formats**:
- Photos: JPG, PNG, HEIC (max 50MB)
- Videos: MP4, MOV, AVI (max 500MB)

### Pagination

List endpoints support cursor-based pagination:

**Request**:
- `page_size` (query param): Number of items per page (1-100, default 50)
- `page_token` (query param): Opaque token from previous response

**Response**:
```json
{
  "media": [...],
  "next_page_token": "eyJpZCI6IjEyMyIsInRpbWUiOiIyMDI1LTExLTI4In0=",
  "total_count": 1234
}
```

**Usage**:
```
GET /api/v1/media?page_size=50
GET /api/v1/media?page_size=50&page_token=eyJpZC...
```

### Error Responses

All errors return consistent structure:

```json
{
  "code": "QUOTA_EXCEEDED",
  "message": "Storage quota exceeded. Please delete files or upgrade your account.",
  "http_status": 507
}
```

**Common Error Codes**:
- `MISSING_REQUIRED` (400) - Required field missing
- `INVALID_REQUEST` (400) - Malformed request
- `UNAUTHORIZED` (401) - Not authenticated
- `FORBIDDEN` (403) - Not authorized for this resource
- `NOT_FOUND` (404) - Resource not found
- `DUPLICATE_EMAIL` (409) - Email already registered
- `QUOTA_EXCEEDED` (507) - Storage quota exceeded
- `INTERNAL_ERROR` (500) - Server error

## Protobuf Generation

To generate Go code from proto files:

```bash
# Install protoc compiler and plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/envoyproxy/protoc-gen-validate@latest

# Generate code (from repository root)
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --validate_out="lang=go:." \
  --validate_opt=paths=source_relative \
  api/proto/v1/*.proto
```

Generated files will be placed in `api/gen/v1/`:
- `*.pb.go` - Protobuf structs and serialization
- `*.pb.validate.go` - Validation functions

## Validation Rules

All protobuf messages include validation rules via `protoc-gen-validate`:

**User Registration**:
- Email: Must be valid email format
- Password: Min 8 chars, must contain letter and number

**Media Upload**:
- File type: Must be in allowed list (image/jpeg, video/mp4, etc.)
- File size: Photos ≤ 50MB, Videos ≤ 500MB

**Album Creation**:
- Name: 1-100 characters
- Description: Max 1000 characters

**Sharing**:
- Recipient emails: Valid email format, max 10 recipients per request

Validation errors return `400 Bad Request` with specific field errors.

## Testing

### Manual Testing

Use curl or Postman to test endpoints:

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "Password123"}' \
  -c cookies.txt

# Upload media (using saved session cookie)
curl -X POST http://localhost:8080/api/v1/media \
  -b cookies.txt \
  -F "file=@photo.jpg"

# List media
curl http://localhost:8080/api/v1/media \
  -b cookies.txt
```

### Integration Tests

Tests MUST validate against these contracts:

```go
// Test using protobuf structs
reqData := &pb.CreateAlbumRequest{
    Name: "Test Album",
    Description: "Test description",
}

// Send to API
body, _ := json.Marshal(reqData)
req := httptest.NewRequest("POST", "/api/v1/albums", bytes.NewReader(body))
rec := httptest.NewRecorder()
mux.ServeHTTP(rec, req)

// Validate response structure
var response pb.Album
json.NewDecoder(rec.Body).Decode(&response)

// Assert using protocmp
if diff := cmp.Diff(expected, &response, protocmp.Transform()); diff != "" {
    t.Errorf("Mismatch (-want +got):\n%s", diff)
}
```

## Versioning

API version is included in URL path (`/api/v1/`).

**Breaking Changes** (require new version):
- Removing fields from messages
- Changing field types
- Removing endpoints
- Changing endpoint URLs
- Changing authentication method

**Non-Breaking Changes** (same version):
- Adding new optional fields
- Adding new endpoints
- Relaxing validation rules

## Next Steps

1. Generate Go code: Run `protoc` commands above
2. Implement services: Create service layer matching protobuf interfaces
3. Implement handlers: Create HTTP handlers that call services
4. Write tests: Map each acceptance scenario to integration test
5. Document: Add inline comments to proto files for auto-generated docs

