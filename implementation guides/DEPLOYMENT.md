# Kisanlink ERP Deployment Guide

This guide provides comprehensive instructions for deploying the Kisanlink ERP system in various environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development](#local-development)
3. [Docker Deployment](#docker-deployment)
4. [Production Deployment](#production-deployment)
5. [Environment Configuration](#environment-configuration)
6. [Monitoring & Health Checks](#monitoring--health-checks)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Go**: 1.21 or higher
- **PostgreSQL**: 12 or higher
- **Redis**: 6 or higher (optional, for caching)
- **Docker**: 20.10 or higher (for containerized deployment)
- **Docker Compose**: 2.0 or higher

### Network Requirements

- **HTTP Port**: 8080 (API)
- **PostgreSQL Port**: 5432
- **Redis Port**: 6379 (optional)

## Local Development

### 1. Clone and Setup

```bash
git clone <repository-url>
cd kisanlink-erp
```

### 2. Install Dependencies

```bash
go mod tidy
go mod download
```

### 3. Database Setup

#### Option A: Local PostgreSQL

```bash
# Install PostgreSQL (Ubuntu/Debian)
sudo apt-get install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql
CREATE DATABASE kisanlink_erp;
CREATE USER kisanlink_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE kisanlink_erp TO kisanlink_user;
\q
```

#### Option B: Docker PostgreSQL

```bash
docker run --name kisanlink-postgres \
  -e POSTGRES_DB=kisanlink_erp \
  -e POSTGRES_USER=kisanlink_user \
  -e POSTGRES_PASSWORD=secure_password \
  -p 5432:5432 \
  -d postgres:15-alpine
```

### 4. Environment Configuration

Create a `.env` file in the project root:

```bash
# Server Configuration
SERVER_HTTP_PORT=8080
SERVER_MODE=debug

# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=kisanlink_user
DATABASE_PASSWORD=secure_password
DATABASE_NAME=kisanlink_erp
DATABASE_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24

# AWS Configuration (for file uploads)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=kisanlink-attachments

# Logging
LOG_LEVEL=debug
```

### 5. Run the Application

```bash
# Development mode
go run cmd/server/main.go

# Or build and run
go build -o bin/kisanlink-erp cmd/server/main.go
./bin/kisanlink-erp
```

### 6. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# List warehouses
curl http://localhost:8080/api/v1/warehouses
```

## Docker Deployment

### 1. Quick Start with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### 2. Individual Docker Commands

```bash
# Build the application
docker build -t kisanlink-erp .

# Run the application
docker run -d \
  --name kisanlink-app \
  -p 8080:8080 \
  -e DATABASE_HOST=your-db-host \
  -e DATABASE_USER=your-db-user \
  -e DATABASE_PASSWORD=your-db-password \
  kisanlink-erp
```

### 3. Docker Compose Services

The `docker-compose.yml` includes:

- **PostgreSQL**: Database server
- **Redis**: Caching and session storage
- **App**: Kisanlink ERP application
- **Nginx**: Reverse proxy (optional)
- **Swagger UI**: API documentation

### 4. Environment-Specific Compose Files

```bash
# Development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

## Production Deployment

### 1. Kubernetes Deployment

Create `k8s/` directory with the following files:

#### `k8s/deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kisanlink-erp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kisanlink-erp
  template:
    metadata:
      labels:
        app: kisanlink-erp
    spec:
      containers:
      - name: kisanlink-erp
        image: kisanlink-erp:latest
                 ports:
         - containerPort: 8080
        env:
        - name: DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: kisanlink-secrets
              key: database-host
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: kisanlink-secrets
              key: database-password
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### `k8s/service.yaml`
```yaml
apiVersion: v1
kind: Service
metadata:
  name: kisanlink-erp-service
spec:
  selector:
    app: kisanlink-erp
  ports:
     - name: http
     port: 8080
     targetPort: 8080
  type: LoadBalancer
```

### 2. Helm Chart

Create a Helm chart for easier deployment:

```bash
helm create kisanlink-erp
```

### 3. CI/CD Pipeline

#### GitHub Actions Example

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Build and test
      run: |
        go mod tidy
        go test ./...
        go build -o bin/kisanlink-erp cmd/server/main.go
    
    - name: Build Docker image
      run: |
        docker build -t kisanlink-erp:${{ github.sha }} .
        docker tag kisanlink-erp:${{ github.sha }} kisanlink-erp:latest
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/kisanlink-erp kisanlink-erp=kisanlink-erp:${{ github.sha }}
```

## Environment Configuration

### 1. Development Environment

```bash
# .env.development
SERVER_MODE=debug
LOG_LEVEL=debug
DATABASE_SSL_MODE=disable
```

### 2. Staging Environment

```bash
# .env.staging
SERVER_MODE=release
LOG_LEVEL=info
DATABASE_SSL_MODE=require
```

### 3. Production Environment

```bash
# .env.production
SERVER_MODE=release
LOG_LEVEL=warn
DATABASE_SSL_MODE=require
JWT_SECRET=<strong-secret-key>
```

### 4. Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HTTP_PORT` | HTTP API port | 8080 | No |
| `AAA_SERVICE_URL` | AAA service URL | localhost:9091 | No |
| `SERVER_MODE` | Gin mode (debug/release) | release | No |
| `DATABASE_HOST` | PostgreSQL host | localhost | Yes |
| `DATABASE_PORT` | PostgreSQL port | 5432 | No |
| `DATABASE_USER` | Database username | - | Yes |
| `DATABASE_PASSWORD` | Database password | - | Yes |
| `DATABASE_NAME` | Database name | kisanlink_erp | No |
| `DATABASE_SSL_MODE` | SSL mode | disable | No |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `JWT_EXPIRY_HOURS` | JWT expiry hours | 24 | No |
| `AWS_REGION` | AWS region | us-east-1 | No |
| `AWS_ACCESS_KEY_ID` | AWS access key | - | No |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | - | No |
| `AWS_S3_BUCKET` | S3 bucket name | - | No |
| `LOG_LEVEL` | Log level | info | No |

## Monitoring & Health Checks

### 1. Health Check Endpoint

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "kisanlink-erp"
}
```

### 2. Prometheus Metrics

Add Prometheus metrics endpoint:

```go
// Add to your server setup
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Add metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 3. Logging

The application uses structured logging with Logrus:

```bash
# View application logs
docker-compose logs -f app

# Filter logs by level
docker-compose logs -f app | grep "level=error"
```

### 4. Database Monitoring

```bash
# Check database connection
docker-compose exec postgres psql -U kisanlink_user -d kisanlink_erp -c "SELECT version();"

# Monitor database performance
docker-compose exec postgres psql -U kisanlink_user -d kisanlink_erp -c "SELECT * FROM pg_stat_activity;"
```

## Troubleshooting

### 1. Common Issues

#### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test database connection
docker-compose exec app wget -qO- http://localhost:8080/health
```

#### Application Startup Issues

```bash
# Check application logs
docker-compose logs app

# Check if ports are available
netstat -tulpn | grep :8080

# Verify environment variables
docker-compose exec app env | grep DATABASE
```

#### Memory Issues

```bash
# Check memory usage
docker stats

# Increase memory limits in docker-compose.yml
services:
  app:
    deploy:
      resources:
        limits:
          memory: 1G
```

### 2. Performance Optimization

#### Database Optimization

```sql
-- Add indexes for better performance
CREATE INDEX idx_warehouses_name ON warehouses(name);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_inventory_batches_warehouse ON inventory_batches(warehouse_id);
```

#### Application Optimization

```bash
# Enable Go profiling
export GODEBUG=gctrace=1

# Monitor CPU and memory
go tool pprof http://localhost:8080/debug/pprof/profile
```

### 3. Security Considerations

#### SSL/TLS Configuration

```bash
# Generate SSL certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Configure nginx with SSL
server {
    listen 443 ssl;
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    # ... rest of configuration
}
```

#### JWT Security

```bash
# Generate strong JWT secret
openssl rand -base64 32

# Rotate JWT secrets regularly
# Use environment-specific secrets
```

### 4. Backup and Recovery

#### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U kisanlink_user kisanlink_erp > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U kisanlink_user kisanlink_erp < backup.sql
```

#### Application Backup

```bash
# Backup configuration
tar -czf kisanlink-config-$(date +%Y%m%d).tar.gz .env docker-compose.yml

# Backup data volumes
docker run --rm -v kisanlink_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-data-$(date +%Y%m%d).tar.gz -C /data .
```

## Support

For deployment issues:

1. Check the logs: `docker-compose logs -f`
2. Verify environment variables
3. Test database connectivity
4. Check network connectivity
5. Review security group/firewall settings

For additional support, please refer to the project documentation or create an issue in the repository.



