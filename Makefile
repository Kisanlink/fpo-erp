.PHONY: build run test clean proto

# Build the application
build:
	go build -o bin/kisanlink-erp cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate protobuf code (requires protoc and go-grpc plugins)
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/stock_import.proto

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run with hot reload (requires air)
dev:
	air

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/kisanlink-erp cmd/server/main.go

# Run with environment file
run-env:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create one based on ENV_SETUP.MD"; \
		exit 1; \
	fi
	go run cmd/server/main.go

# Setup environment (create .env from template)
setup-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from template..."; \
		cp ENV_SETUP.MD .env.template; \
		echo "Please edit .env file with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi

# Database migrations (if using separate migration tool)
migrate:
	# Add your migration commands here
	echo "Migrations handled by GORM auto-migration"

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...



