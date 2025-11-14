.PHONY: build run test test-all test-services test-handlers test-integration test-verbose test-clean clean proto docs install-hooks uninstall-hooks test-hooks update-hooks dev-setup

# Build the application
build:
	go build -o bin/kisanlink-erp cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Run all tests (no cache)
test-all:
	go test ./tests/... -count=1

# Run service layer tests only
test-services:
	go test ./tests/services/... -count=1

# Run handler tests only
test-handlers:
	go test ./tests/handlers/... -count=1

# Run webhook integration tests only
test-integration:
	go test ./tests/integration/... -count=1

# Run all tests with verbose output
test-verbose:
	go test ./tests/... -v -count=1

# Run tests with clean output (using test-clean.sh script)
test-clean:
	bash scripts/test-clean.sh ./tests/... -count=1

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate protobuf code (requires protoc and go-grpc plugins)
proto:
	protoc --go_out=. --go_opt=module=kisanlink-erp \
		--go-grpc_out=. --go-grpc_opt=module=kisanlink-erp \
		proto/authz.proto proto/address.proto

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

# Generate API documentation (requires swag)
docs:
	swag init --parseDependency --parseInternal -g cmd/server/main.go

# Git hooks management
install-hooks:
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { \
		echo "Error: pre-commit not found. Please install it first:"; \
		echo "  pip install pre-commit"; \
		echo "  or: brew install pre-commit (macOS)"; \
		echo "  or: conda install -c conda-forge pre-commit"; \
		exit 1; \
	}
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "✅ Git hooks installed successfully"
	@echo "Run 'make test-hooks' to verify installation"

uninstall-hooks:
	@echo "Uninstalling pre-commit hooks..."
	pre-commit uninstall
	pre-commit uninstall --hook-type commit-msg
	@echo "✅ Git hooks uninstalled"

test-hooks:
	@echo "Testing pre-commit hooks..."
	pre-commit run --all-files
	@echo "✅ Hook tests complete"

update-hooks:
	@echo "Updating pre-commit hooks..."
	pre-commit autoupdate
	@echo "✅ Hooks updated"

# Complete development setup
dev-setup: deps install-hooks
	@echo "✅ Development environment ready"
