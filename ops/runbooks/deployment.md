# Deployment Procedures

## Overview
Step-by-step guide for deploying the Kisanlink ERP application to AWS using CloudFormation.

## Prerequisites
- AWS CLI configured with appropriate permissions
- GitHub Actions access or manual deployment capability
- CloudFormation stack exists or can be created
- ECR image built and pushed
- Parameter file exists for target FPO and environment

## Pre-Deployment Checklist

- [ ] Code reviewed and approved
- [ ] Tests passing (unit, integration)
- [ ] Docker image built and scanned (Trivy)
- [ ] CloudFormation template validated
- [ ] Parameter file updated with correct image URI
- [ ] Staging deployment tested (for production)
- [ ] Rollback plan documented

## Deployment Methods

### Method 1: GitHub Actions (Recommended)

1. **Navigate to GitHub Actions**
   - Go to repository → Actions tab
   - Select "Deploy ERP to AWS" workflow

2. **Trigger Deployment**
   - Click "Run workflow"
   - Enter:
     - **FPO ID**: e.g., `fpo001`
     - **Environment**: `staging` or `production`
     - **Image Tag**: (optional) defaults to `{env}-latest`

3. **Monitor Deployment**
   - Watch workflow execution
   - Check for validation steps
   - Monitor CloudFormation stack events
   - Verify smoke tests pass

### Method 2: Manual Deployment

```bash
# 1. Validate template
cd infrastructure
make validate

# 2. Update parameter file with image URI
# Edit: infrastructure/parameters/<fpo-id>-<env>.json
# Update ECRImageUri parameter

# 3. Deploy stack
bash scripts/deploy.sh <fpo-id> <env>

# 4. Monitor deployment
aws cloudformation describe-stack-events \
  --stack-name kisanlink-erp-<fpo-id>-<env> \
  --max-items 20 \
  --output table
```

## Deployment Steps

### Step 1: Build and Push Docker Image
```bash
# This is typically done via GitHub Actions build workflow
# Manual build:
docker build -t kisanlink-erp:<tag> .
docker tag kisanlink-erp:<tag> <ecr-registry>/kisanlink-erp:<tag>
docker push <ecr-registry>/kisanlink-erp:<tag>
```

### Step 2: Validate CloudFormation Template
```bash
aws cloudformation validate-template \
  --template-body file://infrastructure/templates/erp-application.yaml
```

### Step 3: Update Parameter File
```bash
# Update ECRImageUri in parameter file
jq --arg image "<ecr-uri>:<tag>" \
  '(.[] | select(.ParameterKey == "ECRImageUri") | .ParameterValue) |= $image' \
  infrastructure/parameters/<fpo-id>-<env>.json > tmp.json
mv tmp.json infrastructure/parameters/<fpo-id>-<env>.json
```

### Step 4: Deploy Stack
```bash
cd infrastructure
bash scripts/deploy.sh <fpo-id> <env>
```

### Step 5: Wait for ECS Service Stabilization
```bash
# Get cluster and service names
CLUSTER=$(aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-<fpo-id>-<env> \
  --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
  --output text)

SERVICE=$(aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-<fpo-id>-<env> \
  --query 'Stacks[0].Outputs[?OutputKey==`ECSServiceName`].OutputValue' \
  --output text)

# Wait for service to stabilize
aws ecs wait services-stable \
  --cluster $CLUSTER \
  --services $SERVICE
```

### Step 6: Verify Health Check
```bash
# Get load balancer URL
LB_URL=$(aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-<fpo-id>-<env> \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text)

# Health check
curl -f $LB_URL/health
```

### Step 7: Run Smoke Tests
```bash
# Products endpoint
curl -s -o /dev/null -w "%{http_code}" $LB_URL/api/v1/products

# Warehouses endpoint
curl -s -o /dev/null -w "%{http_code}" $LB_URL/api/v1/warehouses

# Response time
curl -w "\nTime: %{time_total}s\n" $LB_URL/health
```

## Post-Deployment Verification

### 1. Check CloudWatch Alarms
```bash
aws cloudwatch describe-alarms \
  --alarm-name-prefix <fpo-id>-<env>-erp \
  --query 'MetricAlarms[*].[AlarmName,StateValue]' \
  --output table
```

All alarms should be in `OK` state.

### 2. Verify ECS Tasks
```bash
aws ecs list-tasks \
  --cluster <cluster-name> \
  --service-name <service-name> \
  --desired-status RUNNING
```

### 3. Check Application Logs
```bash
aws logs tail /ecs/<fpo-id>-<env>-erp \
  --since 10m \
  --filter-pattern "ERROR" \
  --format short
```

### 4. Monitor Dashboard
- Open CloudWatch Dashboard
- Verify metrics are normal
- Check for any anomalies

## Rollback Procedure

If deployment fails or issues are detected:

See [Rollback Procedures](./rollback.md) runbook.

## Troubleshooting

### Deployment Fails at CloudFormation
- Check CloudFormation events for specific error
- Verify parameter file format
- Check AWS service quotas
- Review IAM permissions

### ECS Service Won't Start
- Check task definition
- Verify image URI is correct
- Check ECS task logs
- Verify security groups allow traffic

### Health Checks Failing
- Check application logs
- Verify database connectivity
- Check ALB target group health
- Review security group rules

## Production Deployment Notes

1. **Always deploy to staging first**
2. **Wait 24 hours** after staging deployment
3. **Perform smoke tests** in staging
4. **Review metrics** before production
5. **Schedule maintenance window** if needed
6. **Have rollback plan ready**
7. **Notify stakeholders** before deployment

## Related Runbooks
- [Rollback Procedures](./rollback.md)
- [Application Errors](./incidents/application-errors.md)
- [Service Recovery](./emergency/service-recovery.md)

---

**Last Updated:** January 2025

