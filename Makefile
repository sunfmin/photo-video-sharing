.PHONY: proto test run docker-up docker-down clean build help

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@VALIDATE_PATH=$$(go list -f '{{.Dir}}' -m github.com/envoyproxy/protoc-gen-validate 2>/dev/null || echo "."); \
	protoc --go_out=. --go_opt=paths=source_relative \
	       --validate_out="lang=go:." --validate_opt=paths=source_relative \
	       -I. -I$$VALIDATE_PATH \
	       api/proto/v1/*.proto
	@echo "Moving generated files to api/gen/v1/..."
	@mv -f api/proto/v1/*.pb.go api/gen/v1/ 2>/dev/null || true
	@mv -f api/proto/v1/*.pb.validate.go api/gen/v1/ 2>/dev/null || true
	@echo "✅ Protobuf generation complete! Generated files are in api/gen/v1/"

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detector
test-race:
	go test -v -race ./...

# Run the application
run:
	go run cmd/api/main.go

# Build the application
build:
	go build -o bin/photovideo cmd/api/main.go

# Start infrastructure (PostgreSQL + MinIO)
docker-up:
	docker-compose up -d

# Stop infrastructure
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Help
help:
	@echo "Available targets:"
	@echo "  proto          - Generate protobuf code"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-race      - Run tests with race detector"
	@echo "  run            - Run the application"
	@echo "  build          - Build the application binary"
	@echo "  docker-up      - Start PostgreSQL and MinIO"
	@echo "  docker-down    - Stop PostgreSQL and MinIO"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
