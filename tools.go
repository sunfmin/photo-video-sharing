//go:build tools
// +build tools

// This file declares dependencies for tooling
package tools

import (
	_ "github.com/disintegration/imaging"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/google/go-cmp/cmp"
	_ "github.com/minio/minio-go/v7"
	_ "github.com/opentracing/opentracing-go"
	_ "github.com/rwcarlsen/goexif/exif"
	_ "github.com/testcontainers/testcontainers-go"
	_ "github.com/u2takey/ffmpeg-go"
	_ "golang.org/x/crypto/bcrypt"
	_ "google.golang.org/protobuf/proto"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)

