# Deployment Order Guide

This document explains the correct order for deploying Kisanlink ERP infrastructure.

## 📋 Overview

The infrastructure is split into **2 types of stacks**:

1. **Shared VPC** - One-time deployment (includes VPC, subnets, and Flow Logs - shared by all FPOs)
2. **Per-FPO Stacks** - One deployment per FPO

---

## 🚀 Deployment Order

### Step 1: Deploy Shared VPC (One-Time)

**When**: Before deploying any FPO stacks  
**Frequency**: Once per environment (staging/production)

```bash
cd infrastructure/scripts
bash deploy-shared-vpc.sh staging
```

**What it creates:**
- VPC
- 2 Public Subnets
- 2 Private Subnets
- Internet Gateway
- NAT Gateway
- Route Tables
- VPC Flow Logs (automatically included)
- CloudWatch Log Group for Flow Logs
- IAM Role for Flow Logs

**Outputs to save:**
- `VPCId` - VPC ID
- `PublicSubnet1Id` - Public Subnet 1 ID
- `PublicSubnet2Id` - Public Subnet 2 ID
- `PrivateSubnet1Id` - Private Subnet 1 ID
- `PrivateSubnet2Id` - Private Subnet 2 ID

---

### Step 2: Deploy FPO Stacks (Per FPO)

**When**: After shared VPC stack is deployed (includes Flow Logs)  
**Frequency**: Once per FPO

#### Option A: Using AWS Console

1. Go to **CloudFormation** console
2. Click **"Create stack"** → **"With new resources"**
3. **"Upload a template file"** → Select `infrastructure/templates/erp-application.yaml`
4. Fill in parameters:

   **Basic:**
   - `FPOIdentifier`: `fpo001`
   - `Environment`: `staging`

   **Network (from Step 1 outputs):**
   - `VpcId`: VPC ID from shared VPC stack
   - `PublicSubnet1Id`: Public Subnet 1 ID
   - `PublicSubnet2Id`: Public Subnet 2 ID
   - `PrivateSubnet1Id`: Private Subnet 1 ID
   - `PrivateSubnet2Id`: Private Subnet 2 ID

   **Container:**
   - `ECRImageUri`: Your ECR image URI
   - `ContainerPort`: `8080`
   - `DesiredCount`: `1` (staging) or `2` (production)
   - `TaskCPU`: `512`
   - `TaskMemory`: `1024`

   **Database:**
   - `DBInstanceClass`: `db.t3.micro`
   - `DBAllocatedStorage`: `20`
   - `DBUsername`: `erp_admin`
   - `DBPassword`: **Enter a strong password** (min 8 characters)

   **AAA Service:**
   - `AAAServiceURL`: `https://aaa.kisanlink.com`
   - `AAAGrpcAddress`: `aaa-grpc.kisanlink.com:9090`
   - `AAAJWTSecret`: **Enter a random 32+ character string**
   - `JWTSecret`: **Enter a random 32+ character string**

   **Other:**
   - `AWSRegion`: `us-west-2`
   - `CertificateArn`: (leave empty for staging)

5. Click **"Next"** → **"Next"**
6. Check **"I acknowledge that AWS CloudFormation might create IAM resources"**
7. Click **"Submit"**

#### Option B: Using Script

```bash
cd infrastructure/scripts

# Update parameter file first
# Edit: infrastructure/parameters/fpo001-staging.json
# Update: VpcId, Subnet IDs, DBPassword, JWTSecret, AAAJWTSecret

bash deploy.sh fpo001 staging
```

---

## 📝 Parameter File Template

For each FPO, create a parameter file: `infrastructure/parameters/{fpo-id}-{env}.json`

```json
[
  {
    "ParameterKey": "FPOIdentifier",
    "ParameterValue": "fpo001"
  },
  {
    "ParameterKey": "Environment",
    "ParameterValue": "staging"
  },
  {
    "ParameterKey": "VpcId",
    "ParameterValue": "vpc-xxxxxxxxxxxxxxxxx"
  },
  {
    "ParameterKey": "PublicSubnet1Id",
    "ParameterValue": "subnet-xxxxxxxxxxxxxxxxx"
  },
  {
    "ParameterKey": "PublicSubnet2Id",
    "ParameterValue": "subnet-yyyyyyyyyyyyyyyyy"
  },
  {
    "ParameterKey": "PrivateSubnet1Id",
    "ParameterValue": "subnet-zzzzzzzzzzzzzzzzz"
  },
  {
    "ParameterKey": "PrivateSubnet2Id",
    "ParameterValue": "subnet-aaaaaaaaaaaaaaaaa"
  },
  {
    "ParameterKey": "ECRImageUri",
    "ParameterValue": "123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest"
  },
  {
    "ParameterKey": "ContainerPort",
    "ParameterValue": "8080"
  },
  {
    "ParameterKey": "DesiredCount",
    "ParameterValue": "1"
  },
  {
    "ParameterKey": "TaskCPU",
    "ParameterValue": "512"
  },
  {
    "ParameterKey": "TaskMemory",
    "ParameterValue": "1024"
  },
  {
    "ParameterKey": "DBInstanceClass",
    "ParameterValue": "db.t3.micro"
  },
  {
    "ParameterKey": "DBAllocatedStorage",
    "ParameterValue": "20"
  },
  {
    "ParameterKey": "DBUsername",
    "ParameterValue": "erp_admin"
  },
  {
    "ParameterKey": "DBPassword",
    "ParameterValue": "YourStrongPassword123!"
  },
  {
    "ParameterKey": "AAAServiceURL",
    "ParameterValue": "https://aaa.kisanlink.com"
  },
  {
    "ParameterKey": "AAAGrpcAddress",
    "ParameterValue": "aaa-grpc.kisanlink.com:9090"
  },
  {
    "ParameterKey": "AAAJWTSecret",
    "ParameterValue": "Your32CharacterAAAJWTSecretKeyHere123456"
  },
  {
    "ParameterKey": "JWTSecret",
    "ParameterValue": "Your32CharacterJWTSecretKeyHere123456"
  },
  {
    "ParameterKey": "AWSRegion",
    "ParameterValue": "us-west-2"
  },
  {
    "ParameterKey": "CertificateArn",
    "ParameterValue": ""
  }
]
```

---

## ✅ Checklist

### Before First FPO Deployment

- [ ] Shared VPC stack deployed (includes Flow Logs)
- [ ] VPC ID and Subnet IDs saved
- [ ] ECR repository created
- [ ] Docker image pushed to ECR
- [ ] Parameter file created for FPO

### For Each FPO Deployment

- [ ] Parameter file updated with:
  - [ ] VPC ID and Subnet IDs (from shared VPC)
  - [ ] ECR image URI
  - [ ] Strong database password (8+ characters)
  - [ ] JWT secret (32+ characters)
  - [ ] AAA JWT secret (32+ characters)
- [ ] Stack deployed successfully
- [ ] Application URL retrieved from outputs
- [ ] Health check passes

---

## 🔄 Adding More FPOs

To add additional FPOs:

1. **No need to redeploy shared VPC** - Reuse existing VPC (Flow Logs already included)
2. **Create new parameter file**: `fpo002-staging.json`
3. **Use same VPC/subnet IDs** from shared VPC stack
4. **Deploy FPO stack** with new FPO ID

---

## 📊 Stack Dependencies

```
Shared VPC Stack (includes VPC Flow Logs)
    ↓
FPO001 Stack (uses VPC/subnet IDs)
FPO002 Stack (uses VPC/subnet IDs)
FPO003 Stack (uses VPC/subnet IDs)
...
```

---

## 🆘 Troubleshooting

### Error: "VPC not found"
- **Cause**: VPC ID is incorrect or shared VPC not deployed
- **Fix**: Deploy shared VPC stack first, verify VPC ID

### Error: "Subnet not found"
- **Cause**: Subnet ID is incorrect or in wrong VPC
- **Fix**: Use subnet IDs from shared VPC stack outputs

### Error: "Secret already exists"
- **Cause**: Secret with same name already exists
- **Fix**: Secrets are created per-FPO, ensure FPO ID is unique

### Error: "Flow log already exists"
- **Cause**: VPC Flow Log already exists for this VPC
- **Fix**: Flow Logs are automatically created in the shared VPC stack. Don't create them separately.

---

**Last Updated:** January 2025

