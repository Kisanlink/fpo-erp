# Kisanlink ERP - AWS Infrastructure

This directory contains CloudFormation templates, scripts, and GitHub Actions workflows for deploying the Kisanlink ERP system to AWS in a multi-tenant architecture.

## 📁 Directory Structure

```
infrastructure/
├── templates/
│   ├── shared-vpc.yaml                # Shared VPC and Flow Logs template (one-time per environment)
│   └── erp-application.yaml           # Per-FPO CloudFormation template
├── parameters/
│   ├── template.json                  # Parameter template for new FPOs
│   ├── fpo001-staging.json            # Example: FPO001 staging parameters
│   └── fpo001-production.json        # Example: FPO001 production parameters
├── scripts/
│   ├── deploy.sh                      # Deploy FPO stack
│   ├── deploy-shared-vpc.sh           # Deploy shared VPC and Flow Logs (one-time)
│   ├── validate.sh                    # Template validation script
│   └── create-fpo.sh                   # Create new FPO configuration
├── Makefile                            # Automation commands
├── README.md                           # This file
└── DEPLOYMENT_ORDER.md                 # Deployment order guide
```

## 🏗️ Architecture Overview

### Multi-Tenant Strategy

Each FPO (Farmer Producer Organization) gets its own isolated infrastructure stack:

- **Separate RDS Database**: Dedicated PostgreSQL instance per FPO
- **Separate S3 Bucket**: Isolated file storage per FPO
- **Separate ECS Service**: Dedicated Fargate tasks per FPO
- **Separate ECS Cluster**: Each FPO gets its own ECS cluster for better isolation (clusters are free)
- **Shared VPC**: All FPO deployments share the same VPC for cost optimization
- **Shared VPC Flow Logs**: Single flow log per VPC/environment (deployed once, used by all FPOs)

### Naming Convention

All resources follow a consistent naming pattern:

```
{FPO_ID}-{Environment}-erp-{resource}

Examples:
- Stack: kisanlink-erp-fpo001-staging
- Database: fpo001-staging-erp-db
- S3 Bucket: fpo001-staging-erp-attachments
- ECS Cluster: fpo001-staging-erp-cluster
- ECS Service: fpo001-staging-erp-service
- Load Balancer: fpo001-staging-erp-alb
```

### AWS Resources Created

For each FPO deployment:

1. **Compute**
   - ECS Fargate tasks (CPU: 512-4096, Memory: 1GB-8GB)
   - Application Load Balancer (ALB)
   - CloudWatch Log Group

2. **Database**
   - RDS PostgreSQL (14.10)
   - DB Subnet Group
   - DB Security Group
   - Automated backups (7 days in production, 1 day in staging)
   - Multi-AZ deployment (production only)

3. **Storage**
   - S3 bucket with encryption
   - Versioning (production only)
   - Lifecycle policies

4. **Security**
   - IAM roles (ECS Task Execution Role, ECS Task Role)
   - Security groups (ALB, ECS Tasks, RDS)
   - Secrets Manager for database credentials

5. **Monitoring**
   - CloudWatch alarms (CPU, Memory, DB Connections)
   - CloudWatch logs

## 🚀 Getting Started

### Prerequisites

1. **AWS CLI** installed and configured
   ```bash
   aws --version
   aws configure
   ```

2. **Shared Infrastructure** (one-time setup per environment):
   - **Shared VPC Stack**: Deploy `shared-vpc.yaml` to create VPC, subnets, NAT gateway, and VPC Flow Logs
   - **ECR Repository**: Create manually or via separate stack
   - Note: ECS Clusters are created automatically per FPO deployment (free resource)

3. **Required IAM Permissions**:
   - CloudFormation: Full access
   - ECS: Full access
   - RDS: Full access
   - S3: Full access
   - IAM: Role creation and management
   - Secrets Manager: Full access

### Quick Start

#### 0. Deploy Shared VPC (REQUIRED FIRST - One-time per Environment)

**IMPORTANT**: Deploy shared VPC before any FPO stacks. This creates the VPC and subnets that all FPOs will use.

```bash
# Deploy shared VPC for staging
cd infrastructure/scripts
bash deploy-shared-vpc.sh staging

# Save the outputs (VPC ID and Subnet IDs) - you'll need them for FPO deployments
```

**Outputs to save:**
- `VPCId` - VPC ID
- `PublicSubnet1Id` - Public Subnet 1 ID
- `PublicSubnet2Id` - Public Subnet 2 ID
- `PrivateSubnet1Id` - Private Subnet 1 ID
- `PrivateSubnet2Id` - Private Subnet 2 ID

**Note**: VPC Flow Logs are automatically created as part of this stack. No separate deployment needed!

#### 1. Create Configuration for New FPO

```bash
# Navigate to infrastructure directory
cd infrastructure

# Create parameter files for new FPO
make create-fpo FPO=fpo001

# This creates:
# - parameters/fpo001-staging.json
# - parameters/fpo001-production.json
```

**Note**: Secrets are now created automatically by the CloudFormation template from parameters you provide. No need to create secrets manually!

#### 3. Update Parameter Values

Edit the generated parameter files and update:

```bash
vi parameters/fpo001-staging.json
```

**Required updates** (from Step 0 outputs):
- `VpcId`: VPC ID from shared VPC stack output
- `PublicSubnet1Id`, `PublicSubnet2Id`: Public subnet IDs from shared VPC stack
- `PrivateSubnet1Id`, `PrivateSubnet2Id`: Private subnet IDs from shared VPC stack
- `ECRImageUri`: Your ECR image URI (e.g., `123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest`)

**Required secrets** (enter as plain values - will be stored in Secrets Manager):
- `DBPassword`: Strong password (min 8 characters)
- `JWTSecret`: Random 32+ character string
- `AAAJWTSecret`: Random 32+ character string

**Optional updates**:
- `CertificateArn`: ACM certificate ARN (production only, for HTTPS)

**Optional updates**:
- `DesiredCount`: Number of ECS tasks (default: 1 for staging, 2 for production)
- `TaskCPU`: CPU units (default: 512 for staging, 1024 for production)
- `TaskMemory`: Memory in MB (default: 1024 for staging, 2048 for production)
- `DBInstanceClass`: RDS instance size (default: db.t3.micro for staging, db.t3.small for production)
- `AAAGrpcAddress`: AAA service gRPC address

#### 4. Validate CloudFormation Template

```bash
make validate
```

#### 4. Deploy FPO Stack

```bash
# Deploy to staging
make deploy-fpo FPO=fpo001 ENV=staging

# Or use the script directly
cd infrastructure/scripts
bash deploy.sh fpo001 staging
```

#### 5. Verify Deployment

```bash
# Get stack outputs (including Load Balancer URL)
make outputs FPO=fpo001 ENV=staging

# Test health endpoint
curl http://<load-balancer-url>/health
```

#### 6. Deploy to Production

After testing staging:

```bash
# Update production parameter file
vi parameters/fpo001-production.json

# Deploy to production
make deploy-fpo FPO=fpo001 ENV=production
```

## 📝 Makefile Commands

### Available Commands

```bash
make help                             # Show all available commands
make validate                         # Validate CloudFormation template
make create-fpo FPO=xxx               # Create parameter files for new FPO
make deploy-fpo FPO=xxx ENV=yyy       # Deploy ERP for specific FPO and environment
make list-stacks                      # List all ERP CloudFormation stacks
make describe-stack STACK=xxx         # Describe specific stack
make outputs FPO=xxx ENV=yyy          # Get stack outputs
make events FPO=xxx ENV=yyy           # Get recent stack events (troubleshooting)
make logs FPO=xxx ENV=yyy             # Tail CloudWatch logs
make delete-stack FPO=xxx ENV=yyy     # Delete specific stack (with confirmation)
```

### Examples

```bash
# Create configuration for FPO002
make create-fpo FPO=fpo002

# Deploy FPO001 to staging
make deploy-fpo FPO=fpo001 ENV=staging

# Deploy FPO002 to production
make deploy-fpo FPO=fpo002 ENV=production

# List all stacks
make list-stacks

# Get outputs for FPO001 staging
make outputs FPO=fpo001 ENV=staging

# Tail logs for FPO001 production
make logs FPO=fpo001 ENV=production

# Delete FPO001 staging stack
make delete-stack FPO=fpo001 ENV=staging
```

## 🔄 CI/CD with GitHub Actions

### Workflows

#### 1. Build and Push Docker Image

**Workflow**: `.github/workflows/build-and-push.yml`

**Triggers**:
- Push to `main` branch → builds production image
- Push to `development` branch → builds staging image
- Manual workflow dispatch → specify environment

**Process**:
1. Run tests
2. Run linter
3. Build Docker image
4. Push to ECR with tags:
   - `{environment}-{git-sha}`
   - `{environment}-latest`

#### 2. Deploy ERP to AWS

**Workflow**: `.github/workflows/deploy-erp.yml`

**Trigger**: Manual workflow dispatch

**Inputs**:
- `fpo_id`: FPO identifier (e.g., fpo001)
- `environment`: staging or production
- `image_tag`: Docker image tag (optional, defaults to `{environment}-latest`)

**Process**:
1. Validate inputs
2. Update parameter file with image URI
3. Validate CloudFormation template
4. Deploy CloudFormation stack
5. Wait for ECS service to stabilize
6. Perform health check
7. Automatic rollback on failure

### Required GitHub Secrets

```
AWS_ROLE_ARN: arn:aws:iam::123456789012:role/GitHubActionsRole
```

### Manual Deployment via GitHub Actions

1. Go to **Actions** tab in GitHub
2. Select **Deploy ERP to AWS** workflow
3. Click **Run workflow**
4. Enter:
   - FPO ID (e.g., `fpo001`)
   - Environment (`staging` or `production`)
   - Image tag (optional, leave empty for latest)
5. Click **Run workflow**

## 🔧 Manual Deployment (Without GitHub Actions)

### 1. Build Docker Image Locally

```bash
# Build image
docker build -t kisanlink-erp:latest .

# Tag for ECR
docker tag kisanlink-erp:latest \
  123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest
```

### 2. Push to ECR

```bash
# Login to ECR
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-west-2.amazonaws.com

# Push image
docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest
```

### 3. Deploy CloudFormation Stack

```bash
cd infrastructure
bash scripts/deploy.sh fpo001 staging
```

## 🔍 Monitoring and Troubleshooting

### CloudWatch Logs

```bash
# Tail logs for specific FPO
make logs FPO=fpo001 ENV=staging

# Or use AWS CLI directly
aws logs tail /ecs/fpo001-staging-erp --follow
```

### Stack Events

```bash
# View recent stack events
make events FPO=fpo001 ENV=staging
```

### ECS Service Status

```bash
# Describe ECS service
aws ecs describe-services \
  --cluster fpo001-staging-erp-cluster \
  --services fpo001-staging-erp-service
```

### RDS Database Connection

```bash
# Get database endpoint from outputs
make outputs FPO=fpo001 ENV=staging

# Connect to database (requires bastion host or VPN)
psql -h <db-endpoint> -U erp_admin -d fpo001_erp
```

### Health Check

```bash
# Get Load Balancer URL
LB_URL=$(aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-fpo001-staging \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text)

# Check health
curl $LB_URL/health
```

## 💰 Cost Estimates

### Per FPO Deployment (Monthly)

**Staging Environment**:
- ECS Fargate (1 task, 0.5 vCPU, 1GB): ~$15/month
- RDS db.t3.micro: ~$15/month
- Application Load Balancer: ~$23/month
- S3 storage (minimal): ~$1/month
- Data transfer & CloudWatch: ~$5/month
- **Total: ~$59/month**

**Production Environment**:
- ECS Fargate (2 tasks, 1 vCPU, 2GB): ~$60/month
- RDS db.t3.small (Multi-AZ): ~$60/month
- Application Load Balancer: ~$23/month
- S3 storage: ~$2/month
- Data transfer & CloudWatch: ~$8/month
- **Total: ~$153/month**

### Shared Resources (One-time)

- VPC: Free
- ECS Cluster: Free
- ECR Repository: ~$1/month per GB
- Secrets Manager: ~$0.40/secret/month

### Cost Optimization Tips

1. **Use Shared VPC**: All FPOs share the same VPC ($0 cost vs $32/VPC/month)
2. **Right-size instances**: Start with smaller instances, scale up as needed
3. **Use spot instances**: For non-critical staging environments (50-70% savings)
4. **Enable S3 lifecycle policies**: Automatically delete old file versions
5. **Use CloudWatch log retention**: Set to 7 days for staging, 30 days for production

## 🔐 Security Best Practices

1. **Secrets Management**:
   - ✅ All secrets stored in AWS Secrets Manager (NOT in parameter files)
   - ✅ Secrets created externally before stack deployment
   - ✅ CloudFormation uses secret ARNs, never plaintext values
   - ✅ ECS tasks retrieve secrets at runtime via IAM permissions
   - ✅ Database passwords can be auto-rotated (optional)
   - ✅ Audit logging of all secret access via CloudTrail
   - ❌ Never commit secrets to Git
   - ❌ Never store secrets in parameter files

2. **Network Security**:
   - ECS tasks in private subnets only
   - RDS in private subnets only
   - ALB in public subnets
   - Security groups restrict traffic

3. **IAM Permissions**:
   - Least privilege principle
   - Separate task execution role and task role
   - Service-specific policies

4. **Data Encryption**:
   - S3 buckets encrypted at rest (AES256)
   - RDS encrypted at rest
   - TLS/SSL for data in transit

5. **Backups**:
   - RDS automated backups (7 days production, 1 day staging)
   - S3 versioning enabled for production
   - Manual snapshots before major changes

## 🆘 Common Issues and Solutions

### Issue 1: Stack Creation Fails - Invalid VPC/Subnet IDs

**Error**: `Invalid id: "vpc-xxxxxxxxx" (expecting "vpc-...")`

**Solution**: Update parameter file with correct VPC and subnet IDs from your AWS account

```bash
# List VPCs
aws ec2 describe-vpcs --query 'Vpcs[*].{ID:VpcId,Name:Tags[?Key==`Name`].Value|[0]}'

# List subnets
aws ec2 describe-subnets --query 'Subnets[*].{ID:SubnetId,VPC:VpcId,AZ:AvailabilityZone,CIDR:CidrBlock}'
```

### Issue 2: ECS Tasks Not Starting

**Error**: `ResourceInitializationError: unable to pull secrets or registry auth`

**Solution**: Verify IAM role permissions for Secrets Manager and ECR

```bash
# Check ECS task logs
make logs FPO=fpo001 ENV=staging
```

### Issue 3: Health Check Failing

**Error**: `Target.FailedHealthChecks`

**Solution**:
1. Check application logs
2. Verify container is listening on correct port
3. Ensure `/health` endpoint returns HTTP 200

```bash
# Check ECS task status
aws ecs describe-tasks \
  --cluster fpo001-staging-erp-cluster \
  --tasks $(aws ecs list-tasks --cluster fpo001-staging-erp-cluster --service-name fpo001-staging-erp-service --query 'taskArns[0]' --output text)
```

### Issue 4: Database Connection Timeout

**Error**: `could not connect to server: Connection timed out`

**Solution**: Verify security group rules allow traffic from ECS tasks to RDS

```bash
# Check RDS security group
aws rds describe-db-instances --db-instance-identifier fpo001-staging-erp-db \
  --query 'DBInstances[0].VpcSecurityGroups'
```

### Issue 5: S3 Access Denied

**Error**: `AccessDenied: Access Denied`

**Solution**: Verify ECS task role has S3 bucket permissions

```bash
# Check task role policy
aws iam get-role-policy \
  --role-name fpo001-staging-erp-task-role \
  --policy-name S3Access
```

## 📊 Rollback Procedures

### Automatic Rollback

CloudFormation automatically rolls back on:
- Stack creation failure
- Stack update failure (with `DeploymentCircuitBreaker` enabled)

GitHub Actions workflow also includes automatic rollback on deployment failure.

### Manual Rollback

#### Option 1: Revert to Previous Stack

```bash
# Update stack with previous parameter values
make deploy-fpo FPO=fpo001 ENV=production
```

#### Option 2: Update ECS Service to Previous Image

```bash
# Update task definition with previous image
aws ecs update-service \
  --cluster fpo001-staging-erp-cluster \
  --service fpo001-production-erp-service \
  --force-new-deployment \
  --task-definition fpo001-production-erp:PREVIOUS_REVISION
```

#### Option 3: Delete Stack and Recreate

```bash
# Delete stack (creates final snapshot for RDS)
make delete-stack FPO=fpo001 ENV=staging

# Recreate from scratch
make deploy-fpo FPO=fpo001 ENV=staging
```

## 📚 Additional Resources

- [AWS CloudFormation Documentation](https://docs.aws.amazon.com/cloudformation/)
- [AWS ECS Fargate Documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html)
- [AWS RDS PostgreSQL Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

## 🤝 Support

For issues or questions:
1. Check CloudWatch logs: `make logs FPO=xxx ENV=yyy`
2. Check stack events: `make events FPO=xxx ENV=yyy`
3. Review this README and troubleshooting section
4. Contact DevOps team

---

**Last Updated**: January 2025
**Version**: 1.0
**Maintained By**: Kisanlink DevOps Team
