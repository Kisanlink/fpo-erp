# Quick Start Guide - Kisanlink ERP Infrastructure

## 🎯 TL;DR

Deploy ERP for a new FPO in 5 commands:

```bash
cd infrastructure
make create-fpo FPO=fpo001
vi parameters/fpo001-staging.json  # Update VPC/subnet IDs and secrets
make validate
make deploy-fpo FPO=fpo001 ENV=staging
```

## 📋 Pre-requisites Checklist

- [ ] AWS CLI installed and configured (`aws configure`)
- [ ] Shared VPC created with public and private subnets
- [ ] ECS Clusters are auto-created per FPO deployment (free resource)
- [ ] ECR Repository created (`kisanlink-erp`)
- [ ] Docker image pushed to ECR
- [ ] **Secrets created in AWS Secrets Manager** (see Step 0 below)
- [ ] IAM role for GitHub Actions (optional, for CI/CD)

## 🚀 First-Time Setup

### 0. Create Secrets in AWS Secrets Manager (REQUIRED)

Before deploying the CloudFormation stack, create secrets for sensitive data:

```bash
# Set your FPO ID and environment
FPO_ID="fpo001"
ENV="staging"
REGION="us-west-2"

# Generate secure secrets
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
AAA_JWT_SECRET=$(openssl rand -base64 32)

# Create database password secret
aws secretsmanager create-secret \
  --name "${FPO_ID}/${ENV}/erp/db-password" \
  --description "Database password for ${FPO_ID} ${ENV} ERP" \
  --secret-string "$DB_PASSWORD" \
  --region $REGION

# Create JWT secret
aws secretsmanager create-secret \
  --name "${FPO_ID}/${ENV}/erp/jwt-secret" \
  --description "JWT secret for ${FPO_ID} ${ENV} ERP" \
  --secret-string "$JWT_SECRET" \
  --region $REGION

# Create AAA JWT secret
aws secretsmanager create-secret \
  --name "${FPO_ID}/${ENV}/erp/aaa-jwt-secret" \
  --description "AAA JWT secret for ${FPO_ID} ${ENV} ERP" \
  --secret-string "$AAA_JWT_SECRET" \
  --region $REGION

# Get the secret ARNs (needed for parameter files)
aws secretsmanager describe-secret --secret-id "${FPO_ID}/${ENV}/erp/db-password" --region $REGION --query 'ARN' --output text
aws secretsmanager describe-secret --secret-id "${FPO_ID}/${ENV}/erp/jwt-secret" --region $REGION --query 'ARN' --output text
aws secretsmanager describe-secret --secret-id "${FPO_ID}/${ENV}/erp/aaa-jwt-secret" --region $REGION --query 'ARN' --output text
```

**Important**: Copy the ARN outputs - you'll need them in Step 3.

### 1. Get AWS Resource IDs

```bash
# Get VPC ID
aws ec2 describe-vpcs --query 'Vpcs[*].{ID:VpcId,Name:Tags[?Key==`Name`].Value|[0]}' --output table

# Get Subnet IDs
aws ec2 describe-subnets --query 'Subnets[*].{ID:SubnetId,Type:Tags[?Key==`Type`].Value|[0],AZ:AvailabilityZone}' --output table

# Get ECR Repository URI
aws ecr describe-repositories --repository-names kisanlink-erp --query 'repositories[0].repositoryUri' --output text
```

### 2. Create FPO Configuration

```bash
cd infrastructure
make create-fpo FPO=fpo001
```

This creates:
- `parameters/fpo001-staging.json`
- `parameters/fpo001-production.json`

### 3. Update Parameter Files

**Required changes** in `parameters/fpo001-staging.json`:

```json
{
  "ParameterKey": "VpcId",
  "ParameterValue": "vpc-0123456789abcdef0"  // ← Your VPC ID
},
{
  "ParameterKey": "PublicSubnet1Id",
  "ParameterValue": "subnet-0123456789abcdef0"  // ← Your public subnet 1
},
{
  "ParameterKey": "PublicSubnet2Id",
  "ParameterValue": "subnet-0123456789abcdef1"  // ← Your public subnet 2
},
{
  "ParameterKey": "PrivateSubnet1Id",
  "ParameterValue": "subnet-0123456789abcdef2"  // ← Your private subnet 1
},
{
  "ParameterKey": "PrivateSubnet2Id",
  "ParameterValue": "subnet-0123456789abcdef3"  // ← Your private subnet 2
},
{
  "ParameterKey": "ECRImageUri",
  "ParameterValue": "123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest"  // ← Your ECR URI
},
{
  "ParameterKey": "DBPasswordSecretArn",
  "ParameterValue": "arn:aws:secretsmanager:us-west-2:123456789012:secret:fpo001/staging/erp/db-password-XXXXXX"  // ← ARN from Step 0
},
{
  "ParameterKey": "JWTSecretArn",
  "ParameterValue": "arn:aws:secretsmanager:us-west-2:123456789012:secret:fpo001/staging/erp/jwt-secret-XXXXXX"  // ← ARN from Step 0
},
{
  "ParameterKey": "AAAJWTSecretArn",
  "ParameterValue": "arn:aws:secretsmanager:us-west-2:123456789012:secret:fpo001/staging/erp/aaa-jwt-secret-XXXXXX"  // ← ARN from Step 0
}
```

**Security Note**: Secrets are now stored in AWS Secrets Manager, not in parameter files. This prevents accidental exposure of sensitive data in version control.

### 4. Validate Template

```bash
make validate
```

### 5. Deploy to Staging

```bash
make deploy-fpo FPO=fpo001 ENV=staging
```

Wait for deployment to complete (~10-15 minutes).

### 6. Verify Deployment

```bash
# Get outputs (includes Load Balancer URL)
make outputs FPO=fpo001 ENV=staging

# Test health endpoint
LB_URL=$(aws cloudformation describe-stacks --stack-name kisanlink-erp-fpo001-staging \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' --output text)

curl $LB_URL/health
# Expected: {"status":"ok"}
```

### 7. Access Application

```bash
# Get Load Balancer URL
make outputs FPO=fpo001 ENV=staging

# Application is now available at:
# http://<load-balancer-dns>/api/v1/...
```

## 📝 Deploy Additional FPOs

For each new FPO, repeat:

```bash
cd infrastructure
make create-fpo FPO=fpo002
vi parameters/fpo002-staging.json  # Update VPC/subnet IDs and secrets
make deploy-fpo FPO=fpo002 ENV=staging
```

Each FPO gets:
- Separate ECS cluster (`fpo002-staging-erp-cluster`)
- Separate database (`fpo002_erp`)
- Separate S3 bucket (`fpo002-staging-erp-attachments`)
- Separate ECS service
- Separate Load Balancer URL

## 🔄 Update Existing Deployment

### Update Docker Image

```bash
# Build and push new image
docker build -t kisanlink-erp:latest .
docker tag kisanlink-erp:latest 123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest
docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest

# Force new deployment
make deploy-fpo FPO=fpo001 ENV=staging
```

### Update Configuration

```bash
# Edit parameter file
vi parameters/fpo001-staging.json

# Update stack
make deploy-fpo FPO=fpo001 ENV=staging
```

## 🗑️ Delete Deployment

```bash
# Delete stack (creates RDS snapshot before deletion)
make delete-stack FPO=fpo001 ENV=staging

# Confirm deletion when prompted
```

## 📊 Common Commands

```bash
# List all stacks
make list-stacks

# View logs
make logs FPO=fpo001 ENV=staging

# View stack outputs
make outputs FPO=fpo001 ENV=staging

# View recent events (troubleshooting)
make events FPO=fpo001 ENV=staging

# Help
make help
```

## 🐛 Troubleshooting

### Deployment Failed

```bash
# Check stack events
make events FPO=fpo001 ENV=staging

# Check ECS task logs
make logs FPO=fpo001 ENV=staging
```

### Health Check Failing

```bash
# View container logs
make logs FPO=fpo001 ENV=staging

# Check ECS service status
aws ecs describe-services \
  --cluster fpo001-staging-erp-cluster \
  --services fpo001-staging-erp-service
```

### Can't Connect to Database

- Verify RDS is in private subnets
- Check security group allows traffic from ECS tasks
- Verify database credentials in Secrets Manager

## 🎓 Next Steps

1. ✅ Deploy to staging and test
2. ✅ Update production parameter file
3. ✅ Deploy to production: `make deploy-fpo FPO=fpo001 ENV=production`
4. ✅ Setup custom domain (Route53 + ACM certificate)
5. ✅ Configure CloudWatch alarms and notifications
6. ✅ Setup automated backups and disaster recovery

## 📚 Full Documentation

See [README.md](README.md) for complete documentation including:
- Architecture overview
- CI/CD with GitHub Actions
- Cost estimates
- Security best practices
- Detailed troubleshooting

---

**Questions?** Check [README.md](README.md) or contact DevOps team.
