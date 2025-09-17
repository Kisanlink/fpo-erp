# Kisanlink ERP Environment Configuration

This guide provides all the environment variables needed to run the Kisanlink ERP application.

## Quick Start

1. Copy the example below to create your `.env` file
2. Update the values according to your environment
3. Run the application

## Complete .env Configuration

```bash
# ========================================
# Kisanlink ERP Environment Configuration
# ========================================

# ========================================
# SERVER CONFIGURATION
# ========================================
# HTTP server port
SERVER_HTTP_PORT=8080



# Application mode (debug, release, test)
SERVER_MODE=release

# ========================================
# DATABASE CONFIGURATION
# ========================================
# PostgreSQL database host
DATABASE_HOST=localhost

# PostgreSQL database port
DATABASE_PORT=5432

# Database username
DATABASE_USER=kisanlink_user

# Database password
DATABASE_PASSWORD=secure_password

# Database name
DATABASE_NAME=kisanlink_erp

# SSL mode (disable, require, verify-ca, verify-full)
DATABASE_SSL_MODE=disable

# ========================================
# JWT CONFIGURATION
# ========================================
# JWT secret key (change this in production!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# JWT token expiry in hours
JWT_EXPIRY_HOURS=24

# ========================================
# AWS CONFIGURATION
# ========================================
# AWS region
AWS_REGION=us-east-1

# AWS Access Key ID
AWS_ACCESS_KEY_ID=your-aws-access-key-id

# AWS Secret Access Key
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key

# S3 bucket name for file attachments
AWS_S3_BUCKET=kisanlink-attachments

# ========================================
# CORS CONFIGURATION
# ========================================
# Allowed origins (comma-separated)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080,https://yourdomain.com

# Allowed headers (comma-separated)
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Request-ID

# ========================================
# AAA SERVICE CONFIGURATION
# ========================================
# AAA service URL for address management (gRPC)
AAA_SERVICE_URL=localhost:9091

# AAA JWT secret for authentication
AAA_JWT_SECRET=your-aaa-jwt-secret

# AAA cache TTL in seconds
AAA_CACHE_TTL=300

# ========================================
# LOGGING CONFIGURATION
# ========================================
# Log level (debug, info, warn, error, fatal, panic)
LOG_LEVEL=info

# ========================================
# DEVELOPMENT OVERRIDES
# ========================================
# Uncomment and modify these for development

# For local development with Docker
# DATABASE_HOST=postgres
# DATABASE_USER=postgres
# DATABASE_PASSWORD=postgres
# DATABASE_NAME=kisanlink_erp_dev

# For local AAA service
# AAA_SERVICE_URL=localhost:50051

# For testing
# SERVER_MODE=debug
# LOG_LEVEL=debug
# DATABASE_NAME=kisanlink_erp_test

# ========================================
# PRODUCTION NOTES
# ========================================
# 1. Change JWT_SECRET to a strong, unique secret
# 2. Use proper AWS credentials with minimal required permissions
# 3. Set DATABASE_SSL_MODE to 'require' or 'verify-full' in production
# 4. Configure proper CORS origins for your frontend domain
# 5. Set SERVER_MODE to 'release' in production
# 6. Use environment-specific database credentials

## Environment-Specific Configurations

### Local Development
```bash
# Copy this section for local development
SERVER_MODE=debug
LOG_LEVEL=debug
DATABASE_HOST=localhost
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=kisanlink_erp_dev
DATABASE_SSL_MODE=disable
```

### Docker Development
```bash
# Copy this section for Docker development
SERVER_MODE=debug
LOG_LEVEL=debug
DATABASE_HOST=postgres
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=kisanlink_erp_dev
DATABASE_SSL_MODE=disable
```

### Production
```bash
# Copy this section for production
SERVER_MODE=release
LOG_LEVEL=info
DATABASE_SSL_MODE=require
JWT_SECRET=your-very-long-and-secure-jwt-secret-key
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### Testing
```bash
# Copy this section for testing
SERVER_MODE=test
LOG_LEVEL=debug
DATABASE_NAME=kisanlink_erp_test
JWT_SECRET=test-secret-key
```

## Database Setup

### PostgreSQL Installation

#### Windows
1. Download PostgreSQL from https://www.postgresql.org/download/windows/
2. Install with default settings
3. Create a database user:
```sql
CREATE USER kisanlink_user WITH PASSWORD 'secure_password';
CREATE DATABASE kisanlink_erp OWNER kisanlink_user;
GRANT ALL PRIVILEGES ON DATABASE kisanlink_erp TO kisanlink_user;
```

#### macOS
```bash
# Using Homebrew
brew install postgresql
brew services start postgresql

# Create database and user
createdb kisanlink_erp
psql -d kisanlink_erp -c "CREATE USER kisanlink_user WITH PASSWORD 'secure_password';"
psql -d kisanlink_erp -c "GRANT ALL PRIVILEGES ON DATABASE kisanlink_erp TO kisanlink_user;"
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql -c "CREATE USER kisanlink_user WITH PASSWORD 'secure_password';"
sudo -u postgres psql -c "CREATE DATABASE kisanlink_erp OWNER kisanlink_user;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE kisanlink_erp TO kisanlink_user;"
```

### Docker Setup
```bash
# Create a docker-compose.yml file
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: kisanlink_erp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## AWS Setup

### S3 Bucket Creation
1. Go to AWS S3 Console
2. Create a new bucket named `kisanlink-attachments`
3. Configure bucket permissions:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::kisanlink-attachments/*"
        }
    ]
}
```

### IAM User Setup
1. Create an IAM user for the application
2. Attach the following policy:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::kisanlink-attachments",
                "arn:aws:s3:::kisanlink-attachments/*"
            ]
        }
    ]
}
```

## Security Best Practices

### JWT Secret
- Use a strong, random secret (at least 32 characters)
- Generate using: `openssl rand -base64 32`
- Never commit secrets to version control

### Database Security
- Use strong passwords
- Enable SSL in production
- Restrict database access to application servers only
- Regular backups

### AWS Security
- Use IAM roles instead of access keys when possible
- Rotate access keys regularly
- Use least privilege principle
- Enable CloudTrail for audit logging

## Troubleshooting

### Common Issues

#### Database Connection Failed
- Check if PostgreSQL is running
- Verify database credentials
- Ensure database exists
- Check firewall settings

#### AWS S3 Access Denied
- Verify AWS credentials
- Check IAM permissions
- Ensure S3 bucket exists
- Verify bucket permissions

#### JWT Token Issues
- Check JWT secret configuration
- Verify token expiry settings
- Ensure consistent secret across instances

### Logs
- Check application logs for detailed error messages
- Use `LOG_LEVEL=debug` for development
- Monitor database connection logs
- Review AWS CloudWatch logs

## Next Steps

1. Create your `.env` file with the appropriate configuration
2. Set up PostgreSQL database
3. Configure AWS S3 bucket and credentials
4. Run database migrations
5. Start the application: `go run cmd/server/main.go`

## Support

For issues related to:
- Environment configuration: Check this guide
- Database setup: Refer to PostgreSQL documentation
- AWS setup: Check AWS documentation
- Application issues: Review application logs