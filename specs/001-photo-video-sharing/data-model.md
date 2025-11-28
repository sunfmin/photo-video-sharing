# Data Model: Photo Video Sharing System

**Feature**: 001-photo-video-sharing  
**Date**: Friday Nov 28, 2025  
**Source**: Derived from spec.md functional requirements and research.md decisions

## Entity Relationship Diagram

```
┌─────────────────┐
│     User        │
│─────────────────│
│ id (PK)         │
│ email (unique)  │
│ password_hash   │
│ storage_used    │
│ storage_quota   │
│ created_at      │
│ updated_at      │
└────────┬────────┘
         │
         │ 1:N (owner)
         │
         ▼
┌─────────────────┐         ┌──────────────────┐
│     Media       │ N:M     │   AlbumMedia     │
│─────────────────│◄────────┤──────────────────│
│ id (PK)         │         │ album_id (FK)    │
│ owner_id (FK)   │         │ media_id (FK)    │
│ filename        │         │ added_at         │
│ file_type       │         └─────────┬────────┘
│ file_size       │                   │
│ storage_path    │                   │
│ thumbnail_path  │                   │ N:1
│ width           │                   │
│ height          │                   ▼
│ duration        │         ┌──────────────────┐
│ exif_data       │         │     Album        │
│ uploaded_at     │         │──────────────────│
│ created_at      │         │ id (PK)          │
│ updated_at      │         │ owner_id (FK)    │
└────────┬────────┘         │ name             │
         │                  │ description      │
         │                  │ created_at       │
         │ 1:N              │ updated_at       │
         │                  └─────────┬────────┘
         ▼                            │
┌─────────────────┐                  │ 1:N
│     Share       │                  │
│─────────────────│                  ▼
│ id (PK)         │         ┌──────────────────┐
│ media_id (FK)   │         │   AlbumShare     │
│ album_id (FK)   │         │──────────────────│
│ owner_id (FK)   │         │ id (PK)          │
│ shared_with (FK)│         │ album_id (FK)    │
│ created_at      │         │ owner_id (FK)    │
└─────────────────┘         │ shared_with (FK) │
                            │ created_at       │
         ┌──────────────────┴──────────────────┘
         │
         │ N:1 (user)
         │
         ▼
┌─────────────────┐
│    Session      │
│─────────────────│
│ id (PK)         │
│ user_id (FK)    │
│ expires_at      │
│ last_used_at    │
│ created_at      │
└─────────────────┘
```

**Cardinality Legend**:
- `1:N` = One-to-many
- `N:M` = Many-to-many (through junction table)
- `N:1` = Many-to-one

---

## Entity Definitions

### User

Represents a registered account with authentication credentials and storage allocation.

**Attributes**:
- `id` (UUID, PK): Unique identifier
- `email` (string, unique, not null): Login email address (indexed)
- `password_hash` (string, not null): Bcrypt hashed password
- `storage_used` (bigint, default 0): Current storage usage in bytes
- `storage_quota` (bigint, default 524288000): Maximum allowed storage (500MB)
- `created_at` (timestamp): Account creation timestamp
- `updated_at` (timestamp): Last modification timestamp

**Validation Rules** (from FR-008, FR-009):
- Email must be valid format
- Password minimum 8 characters, at least one letter and one number
- Storage quota must be positive
- Storage used cannot exceed quota

**Relationships**:
- Has many `Media` (as owner)
- Has many `Album` (as owner)
- Has many `Share` (as owner)
- Has many `Share` (as recipient via shared_with)
- Has many `AlbumShare` (as owner and recipient)
- Has many `Session`

**Indexes**:
- Primary: `id`
- Unique: `email`

---

### Media

Represents an uploaded photo or video file with metadata and storage references.

**Attributes**:
- `id` (UUID, PK): Unique identifier
- `owner_id` (UUID, FK → User, not null): Owner of the media
- `filename` (string, not null): Original filename (user-friendly display)
- `file_type` (string, not null): MIME type (image/jpeg, video/mp4, etc.)
- `file_size` (bigint, not null): File size in bytes
- `storage_path` (string, not null): Object storage key for original file
- `thumbnail_path` (string, nullable): Object storage key for thumbnail
- `width` (int, nullable): Image/video width in pixels
- `height` (int, nullable): Image/video height in pixels
- `duration` (int, nullable): Video duration in seconds (null for images)
- `exif_data` (jsonb, nullable): EXIF metadata for photos (date taken, camera, GPS, etc.)
- `uploaded_at` (timestamp, not null): Upload completion timestamp
- `created_at` (timestamp): Record creation timestamp
- `updated_at` (timestamp): Last modification timestamp

**Validation Rules** (from FR-001, FR-002, FR-004):
- File type must be in allowed list: image/jpeg, image/png, image/heic, video/mp4, video/quicktime, video/x-msvideo
- Photo file size ≤ 52428800 bytes (50MB)
- Video file size ≤ 524288000 bytes (500MB)
- Storage path must be non-empty
- Owner ID must reference valid user

**Relationships**:
- Belongs to `User` (owner)
- Has many `Share`
- Belongs to many `Album` through `AlbumMedia`

**Indexes**:
- Primary: `id`
- Foreign key: `owner_id` (for user's media queries)
- Composite: `(owner_id, uploaded_at DESC)` (for gallery pagination)
- Single: `file_type` (for filtering photos vs videos)

---

### Album

Represents a collection of media items organized by the user.

**Attributes**:
- `id` (UUID, PK): Unique identifier
- `owner_id` (UUID, FK → User, not null): Owner of the album
- `name` (string, not null): Album name (max 100 characters)
- `description` (text, nullable): Optional album description
- `created_at` (timestamp): Album creation timestamp
- `updated_at` (timestamp): Last modification timestamp

**Validation Rules** (from FR-025):
- Name must be 1-100 characters
- Owner ID must reference valid user
- Name allows alphanumeric, spaces, hyphens, apostrophes

**Relationships**:
- Belongs to `User` (owner)
- Has many `Media` through `AlbumMedia`
- Has many `AlbumShare`

**Indexes**:
- Primary: `id`
- Foreign key: `owner_id` (for user's albums)
- Composite: `(owner_id, created_at DESC)` (for album list)

---

### AlbumMedia (Junction Table)

Links media items to albums (many-to-many relationship).

**Attributes**:
- `album_id` (UUID, FK → Album, not null): Album reference
- `media_id` (UUID, FK → Media, not null): Media reference
- `added_at` (timestamp): When media was added to album

**Validation Rules** (from FR-026, FR-027):
- Album and media must exist
- Same media can be in multiple albums
- Media can only be added once per album (unique constraint)

**Relationships**:
- Belongs to `Album`
- Belongs to `Media`

**Indexes**:
- Primary: `(album_id, media_id)` (composite)
- Single: `media_id` (for reverse lookup: which albums contain this media)

---

### Share

Represents sharing of individual media items with specific users.

**Attributes**:
- `id` (UUID, PK): Unique identifier
- `media_id` (UUID, FK → Media, nullable): Shared media item (null if album share)
- `album_id` (UUID, FK → Album, nullable): Shared album (null if media share)
- `owner_id` (UUID, FK → User, not null): User who created the share
- `shared_with_user_id` (UUID, FK → User, not null): User receiving access
- `created_at` (timestamp): Share creation timestamp

**Validation Rules** (from FR-019, FR-020, FR-024):
- Exactly one of media_id or album_id must be non-null (XOR constraint)
- Owner must be the actual owner of the media/album
- Cannot share with yourself (shared_with ≠ owner_id)
- Same media cannot be shared with same user twice (unique constraint)

**Relationships**:
- Belongs to `Media` (nullable)
- Belongs to `Album` (nullable)
- Belongs to `User` (owner)
- Belongs to `User` (shared_with)

**Indexes**:
- Primary: `id`
- Composite: `(media_id, shared_with_user_id)` (unique, for media shares)
- Composite: `(album_id, shared_with_user_id)` (unique, for album shares)
- Single: `shared_with_user_id` (for "Shared with me" queries)

**Note**: Initially combined media and album shares in single table. If complexity grows, split into `MediaShare` and `AlbumShare` tables.

---

### AlbumShare

Represents sharing of entire albums with specific users (separate from media shares for clarity).

**Attributes**:
- `id` (UUID, PK): Unique identifier
- `album_id` (UUID, FK → Album, not null): Shared album
- `owner_id` (UUID, FK → User, not null): Album owner who created the share
- `shared_with_user_id` (UUID, FK → User, not null): User receiving access
- `created_at` (timestamp): Share creation timestamp

**Validation Rules** (from FR-024):
- Owner must be the album owner
- Cannot share with yourself
- Same album cannot be shared with same user twice

**Relationships**:
- Belongs to `Album`
- Belongs to `User` (owner)
- Belongs to `User` (shared_with)

**Indexes**:
- Primary: `id`
- Composite: `(album_id, shared_with_user_id)` (unique)
- Single: `shared_with_user_id` (for "Shared albums" queries)

**Access Logic**:
When album is shared, recipient can view ALL media in album, including future additions (dynamic access).

---

### Session

Represents an active user session for authentication.

**Attributes**:
- `id` (UUID, PK): Session identifier (stored in HTTP-only cookie)
- `user_id` (UUID, FK → User, not null): Authenticated user
- `expires_at` (timestamp, not null): Session expiration time (7 days from creation)
- `last_used_at` (timestamp, not null): Last activity timestamp (updated on requests)
- `created_at` (timestamp): Session creation timestamp

**Validation Rules** (from research.md authentication decision):
- Session ID must be cryptographically random UUID
- Expires at must be in future
- Session is valid if current time < expires_at

**Relationships**:
- Belongs to `User`

**Indexes**:
- Primary: `id`
- Foreign key: `user_id` (for user session lookup)
- Single: `expires_at` (for cleanup job to delete expired sessions)

**Cleanup**:
Daily background job deletes sessions where `expires_at < NOW()`.

---

## GORM Model Examples

### User Model

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email         string    `gorm:"uniqueIndex;not null"`
    PasswordHash  string    `gorm:"not null"`
    StorageUsed   int64     `gorm:"default:0"`
    StorageQuota  int64     `gorm:"default:524288000"` // 500MB
    CreatedAt     time.Time
    UpdatedAt     time.Time
    
    // Relationships (not stored in DB)
    Media         []Media         `gorm:"foreignKey:OwnerID"`
    Albums        []Album         `gorm:"foreignKey:OwnerID"`
    Sessions      []Session       `gorm:"foreignKey:UserID"`
}
```

### Media Model

```go
package models

import (
    "time"
    "gorm.io/datatypes"
)

type Media struct {
    ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OwnerID       string         `gorm:"type:uuid;not null;index:idx_owner_uploaded,priority:1"`
    Filename      string         `gorm:"not null"`
    FileType      string         `gorm:"not null;index"`
    FileSize      int64          `gorm:"not null"`
    StoragePath   string         `gorm:"not null"`
    ThumbnailPath *string        // Nullable
    Width         *int           // Nullable
    Height        *int           // Nullable
    Duration      *int           // Nullable (video only)
    EXIFData      datatypes.JSON `gorm:"type:jsonb"` // Nullable
    UploadedAt    time.Time      `gorm:"not null;index:idx_owner_uploaded,priority:2,sort:desc"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    
    // Relationships
    Owner         User           `gorm:"foreignKey:OwnerID"`
    Albums        []Album        `gorm:"many2many:album_media"`
    Shares        []Share        `gorm:"foreignKey:MediaID"`
}
```

### Album Model

```go
package models

import "time"

type Album struct {
    ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OwnerID     string    `gorm:"type:uuid;not null;index:idx_owner_created,priority:1"`
    Name        string    `gorm:"not null;size:100"`
    Description *string   // Nullable
    CreatedAt   time.Time `gorm:"index:idx_owner_created,priority:2,sort:desc"`
    UpdatedAt   time.Time
    
    // Relationships
    Owner       User         `gorm:"foreignKey:OwnerID"`
    Media       []Media      `gorm:"many2many:album_media"`
    Shares      []AlbumShare `gorm:"foreignKey:AlbumID"`
}
```

### AlbumMedia Model

```go
package models

import "time"

type AlbumMedia struct {
    AlbumID  string    `gorm:"primaryKey;type:uuid"`
    MediaID  string    `gorm:"primaryKey;type:uuid;index"`
    AddedAt  time.Time `gorm:"autoCreateTime"`
    
    // Relationships
    Album    Album     `gorm:"foreignKey:AlbumID"`
    Media    Media     `gorm:"foreignKey:MediaID"`
}
```

### Share Model

```go
package models

import "time"

type Share struct {
    ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    MediaID          *string   `gorm:"type:uuid;index:idx_media_shared,priority:1"` // Nullable
    AlbumID          *string   `gorm:"type:uuid;index:idx_album_shared,priority:1"` // Nullable
    OwnerID          string    `gorm:"type:uuid;not null"`
    SharedWithUserID string    `gorm:"type:uuid;not null;index:idx_media_shared,priority:2;index:idx_album_shared,priority:2;index:idx_shared_with"`
    CreatedAt        time.Time
    
    // Relationships
    Media            *Media    `gorm:"foreignKey:MediaID"`
    Album            *Album    `gorm:"foreignKey:AlbumID"`
    Owner            User      `gorm:"foreignKey:OwnerID"`
    SharedWith       User      `gorm:"foreignKey:SharedWithUserID"`
}

// BeforeCreate hook ensures exactly one of MediaID or AlbumID is set
func (s *Share) BeforeCreate(tx *gorm.DB) error {
    if (s.MediaID == nil && s.AlbumID == nil) || (s.MediaID != nil && s.AlbumID != nil) {
        return fmt.Errorf("exactly one of MediaID or AlbumID must be set")
    }
    return nil
}
```

### Session Model

```go
package models

import "time"

type Session struct {
    ID         string    `gorm:"primaryKey;type:uuid"`
    UserID     string    `gorm:"type:uuid;not null;index"`
    ExpiresAt  time.Time `gorm:"not null;index"`
    LastUsedAt time.Time `gorm:"not null"`
    CreatedAt  time.Time
    
    // Relationships
    User       User      `gorm:"foreignKey:UserID"`
}
```

---

## Database Constraints

### Check Constraints

```sql
-- User storage constraints
ALTER TABLE users ADD CONSTRAINT chk_storage_quota_positive 
    CHECK (storage_quota > 0);
ALTER TABLE users ADD CONSTRAINT chk_storage_within_quota 
    CHECK (storage_used >= 0 AND storage_used <= storage_quota);

-- Media file size constraints
ALTER TABLE media ADD CONSTRAINT chk_file_size_positive 
    CHECK (file_size > 0);

-- Album name length
ALTER TABLE albums ADD CONSTRAINT chk_name_length 
    CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100);

-- Share XOR constraint (media or album, not both)
ALTER TABLE shares ADD CONSTRAINT chk_share_target 
    CHECK ((media_id IS NOT NULL AND album_id IS NULL) OR 
           (media_id IS NULL AND album_id IS NOT NULL));

-- Session expiration in future
ALTER TABLE sessions ADD CONSTRAINT chk_expires_future 
    CHECK (expires_at > created_at);
```

### Unique Constraints

```sql
-- Prevent duplicate shares
CREATE UNIQUE INDEX idx_unique_media_share 
    ON shares (media_id, shared_with_user_id) 
    WHERE media_id IS NOT NULL;

CREATE UNIQUE INDEX idx_unique_album_share 
    ON shares (album_id, shared_with_user_id) 
    WHERE album_id IS NOT NULL;

-- Prevent duplicate album media
ALTER TABLE album_media ADD CONSTRAINT unique_album_media 
    UNIQUE (album_id, media_id);
```

### Foreign Key Constraints

All foreign keys use `ON DELETE CASCADE` to ensure referential integrity:

```sql
-- Media ownership
ALTER TABLE media ADD CONSTRAINT fk_media_owner 
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

-- Album ownership
ALTER TABLE albums ADD CONSTRAINT fk_album_owner 
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

-- Share relationships
ALTER TABLE shares ADD CONSTRAINT fk_share_media 
    FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE;
ALTER TABLE shares ADD CONSTRAINT fk_share_owner 
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE shares ADD CONSTRAINT fk_share_recipient 
    FOREIGN KEY (shared_with_user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Album-Media junction
ALTER TABLE album_media ADD CONSTRAINT fk_album_media_album 
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE;
ALTER TABLE album_media ADD CONSTRAINT fk_album_media_media 
    FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE;

-- Sessions
ALTER TABLE sessions ADD CONSTRAINT fk_session_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

---

## State Transitions

### Media Lifecycle

```
[Upload Initiated]
       ↓
[Temporary Storage] → (validation fails) → [Rejected]
       ↓
[Processing] → (thumbnail generation, EXIF extraction)
       ↓
[Object Storage Upload]
       ↓
[Database Record Created] → [Active]
       ↓
[User Views/Shares] → [Active]
       ↓
[User Deletes] → [Deleted from Storage] → [Database Record Removed]
```

**States**:
- **Processing**: File uploaded, thumbnails being generated (not visible in gallery yet)
- **Active**: Fully processed, available for viewing
- **Deleted**: Removed from storage and database (cascades to shares)

### Share Lifecycle

```
[Created by Owner]
       ↓
[Active] → (recipient can view)
       ↓
       ├─→ [Revoked by Owner] → [Deleted]
       ├─→ [Media/Album Deleted] → [Cascade Deleted]
       └─→ [Recipient Account Deleted] → [Cascade Deleted]
```

### Session Lifecycle

```
[Created on Login]
       ↓
[Active] → (updated on each request)
       ↓
       ├─→ [Explicit Logout] → [Deleted]
       ├─→ [Expiration Time Reached] → [Cleanup Job Deletes]
       └─→ [User Account Deleted] → [Cascade Deleted]
```

---

## Query Patterns

### Common Queries

**User's Media Gallery** (paginated):
```sql
SELECT * FROM media 
WHERE owner_id = ? 
ORDER BY uploaded_at DESC 
LIMIT 50 OFFSET ?;
```

**Shared With Me**:
```sql
SELECT m.* FROM media m
INNER JOIN shares s ON s.media_id = m.id
WHERE s.shared_with_user_id = ?
ORDER BY s.created_at DESC;
```

**Album Contents**:
```sql
SELECT m.* FROM media m
INNER JOIN album_media am ON am.media_id = m.id
WHERE am.album_id = ?
ORDER BY am.added_at DESC;
```

**Check Media Access** (owner or shared):
```sql
SELECT EXISTS(
    SELECT 1 FROM media m
    LEFT JOIN shares s ON s.media_id = m.id
    WHERE m.id = ? 
    AND (m.owner_id = ? OR s.shared_with_user_id = ?)
) AS has_access;
```

**Storage Quota Check**:
```sql
SELECT storage_used, storage_quota 
FROM users 
WHERE id = ? 
FOR UPDATE; -- Lock row for quota transaction
```

---

## Migrations Strategy

**GORM AutoMigrate** (development and testing):
```go
// services/migrations.go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Media{},
        &models.Album{},
        &models.AlbumMedia{},
        &models.Share{},
        &models.Session{},
    )
}
```

**Production Migrations** (future):
Use migration tool (e.g., golang-migrate, goose) for versioned, reversible migrations with explicit SQL control.

---

## Data Retention

**Session Cleanup** (daily cron job):
```sql
DELETE FROM sessions WHERE expires_at < NOW();
```

**Soft Deletes** (not implemented initially):
Future enhancement: Add `deleted_at` column for soft deletes, allowing recovery period before permanent deletion.

**Audit Trail** (not implemented initially):
Future enhancement: Add audit table logging all media/album/share operations for compliance.

---

## Next Steps

Data model complete. Proceed to:
1. Generate API contracts (OpenAPI spec from protobuf definitions)
2. Generate quickstart.md (developer setup guide)
3. Update agent context with technology stack

