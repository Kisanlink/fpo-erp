#!/bin/bash
set -e

echo "🚀 Kisanlink ERP - Local Development Setup"
echo "=========================================="

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi
echo "✅ Go $(go version | awk '{print $3}')"

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi
echo "✅ Docker $(docker --version | awk '{print $3}' | tr -d ',')"

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi
echo "✅ Docker Compose $(docker-compose --version | awk '{print $3}' | tr -d ',')"

# Install development tools
echo ""
echo "Installing development tools..."
go install github.com/swaggo/swag/cmd/swag@latest
echo "✅ Swagger CLI installed"

go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
echo "✅ golangci-lint installed"

go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "✅ gosec installed"

# Setup .env file
echo ""
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << 'EOF'
# Server Configuration
SERVER_HTTP_PORT=8080
SERVER_MODE=debug
SERVER_PUBLIC_URL=http://localhost:8080
JWT_SECRET=local-dev-jwt-secret-32-characters

# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=kisanlink_erp
POSTGRES_USER=erp_admin
POSTGRES_PASSWORD=local_dev_password
POSTGRES_MAX_IDLE_CONNS=10
POSTGRES_MAX_OPEN_CONNS=100
POSTGRES_CONN_MAX_LIFETIME_SECONDS=3600

# AAA Service Configuration (Bypass Mode)
AAA_ENABLED=false
AAA_JWT_SECRET=local-aaa-jwt-secret
AAA_GRPC_ADDRESS=localhost:9090
AAA_CACHE_TTL_MINUTES=60
AAA_TIMEOUT_SECONDS=30

# AWS S3 Configuration (MinIO)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_S3_BUCKET=kisanlink-erp-attachments
AWS_ENDPOINT=http://localhost:9000
AWS_S3_FORCE_PATH_STYLE=true

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_HEADERS=Content-Type,Authorization
EOF
    echo "✅ Created .env file"
else
    echo "⚠️  .env file already exists, skipping..."
fi

# Install pre-commit hooks
echo ""
echo "Installing pre-commit hooks..."
if command -v pre-commit &> /dev/null; then
    pre-commit install
    echo "✅ Pre-commit hooks installed"
else
    echo "⚠️  pre-commit not found, skipping hooks installation"
    echo "   Install with: pip install pre-commit"
fi

# Start infrastructure
echo ""
echo "Starting infrastructure services..."
docker-compose up -d

echo "Waiting for services to be healthy..."
sleep 10

# Create MinIO bucket
echo ""
echo "Configuring MinIO..."
docker run --rm --network host --entrypoint /bin/sh minio/mc -c "
  mc alias set local http://localhost:9000 minioadmin minioadmin
  mc mb local/kisanlink-erp-attachments --ignore-existing
  echo 'Bucket created: kisanlink-erp-attachments'
"
echo "✅ MinIO configured"

# Download Go dependencies
echo ""
echo "Downloading Go dependencies..."
go mod download
echo "✅ Dependencies downloaded"

# Run database migrations
echo ""
echo "Running database migrations..."
echo "Note: Migrations will run automatically when you start the server"

echo ""
echo "=========================================="
echo "✅ Setup complete!"
echo ""
echo "Quick start:"
echo "  1. Start the server:  make run"
echo "  2. Run tests:         make test"
echo "  3. View services:"
echo "     - ERP API:         http://localhost:8080"
echo "     - MinIO Console:   http://localhost:9001 (minioadmin/minioadmin)"
echo "     - PostgreSQL:      localhost:5432 (erp_admin/local_dev_password)"
echo ""
echo "Stop services: docker-compose down"
echo "=========================================="
