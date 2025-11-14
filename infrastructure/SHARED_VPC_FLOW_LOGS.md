# Shared VPC Flow Logs Setup

## Overview

VPC Flow Logs capture IP traffic information for network interfaces in a VPC. Since AWS only allows **one flow log per VPC**, we deploy VPC Flow Logs in a separate shared infrastructure stack that all FPOs use.

## Why Separate Stack?

- ✅ **AWS Limitation**: Only one flow log can exist per VPC
- ✅ **Cost Optimization**: Single log group for all FPOs
- ✅ **Centralized Monitoring**: All network traffic in one place
- ✅ **Compliance**: Network audit trail for all tenants

## Architecture

```
┌─────────────────────────────────────────┐
│  Shared VPC (vpc-xxxxx)                │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  VPC Flow Log (ONE per VPC)      │  │
│  │  → CloudWatch Logs               │  │
│  │  → /aws/vpc/flowlogs/...         │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │ FPO001 Stack │  │ FPO002 Stack │   │
│  │ (uses shared │  │ (uses shared │   │
│  │  flow log)   │  │  flow log)   │   │
│  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────┘
```

## Deployment

### Prerequisites

- AWS CLI configured
- VPC ID where FPOs will be deployed
- IAM permissions for CloudFormation and VPC Flow Logs

### Deploy Shared VPC Flow Logs

```bash
# Navigate to infrastructure directory
cd infrastructure/scripts

# Deploy for staging
bash deploy-shared-vpc-flow-logs.sh \
  vpc-1234567890abcdef0 \
  staging \
  7  # Retention: 7 days for staging

# Deploy for production
bash deploy-shared-vpc-flow-logs.sh \
  vpc-1234567890abcdef0 \
  production \
  90  # Retention: 90 days for production
```

### Parameters

- **VPC ID**: The shared VPC where all FPOs will be deployed
- **Environment**: `staging` or `production`
- **Retention Days**: How long to keep logs (optional, default: 90)

### Stack Name

The stack will be named: `kisanlink-erp-shared-vpc-flow-logs-{environment}`

Example:
- Staging: `kisanlink-erp-shared-vpc-flow-logs-staging`
- Production: `kisanlink-erp-shared-vpc-flow-logs-production`

## What Gets Created

1. **IAM Role** (`VPCFlowLogsRole`)
   - Allows VPC Flow Logs service to write to CloudWatch Logs
   - Named: `kisanlink-erp-shared-vpc-flow-logs-role-{environment}`

2. **CloudWatch Log Group** (`VPCFlowLogsLogGroup`)
   - Stores all VPC flow log data
   - Named: `/aws/vpc/flowlogs/kisanlink-erp-shared-{environment}`
   - Retention: Configurable (default: 90 days production, 7 days staging)

3. **VPC Flow Log** (`VPCFlowLog`)
   - Captures ALL traffic (ACCEPT + REJECT) for the VPC
   - Writes to the CloudWatch Log Group above

## Verification

### Check Stack Status

```bash
aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-shared-vpc-flow-logs-staging \
  --query 'Stacks[0].StackStatus' \
  --output text
```

Should return: `CREATE_COMPLETE` or `UPDATE_COMPLETE`

### Check Flow Log Status

```bash
# Get VPC ID
VPC_ID="vpc-1234567890abcdef0"

# Check if flow log exists
aws ec2 describe-flow-logs \
  --filter Name=resource-id,Values=$VPC_ID \
  --query 'FlowLogs[*].[FlowLogId,FlowLogStatus]' \
  --output table
```

### View Logs

```bash
# View recent flow logs
aws logs tail /aws/vpc/flowlogs/kisanlink-erp-shared-staging \
  --since 1h \
  --format short
```

## Important Notes

### ⚠️ One-Time Deployment

- **Deploy ONCE per VPC/environment** before deploying any FPO stacks
- If you try to deploy again, it will update the existing stack (safe)
- All FPOs automatically use this shared flow log

### ⚠️ Per-FPO Stacks

- **Do NOT** include VPC Flow Logs in per-FPO stacks
- The `erp-application.yaml` template has VPC Flow Logs removed
- Each FPO stack will use the shared flow log automatically

### ⚠️ Cost Considerations

- CloudWatch Logs charges for ingestion and storage
- Retention period affects storage costs
- Production: 90 days retention (recommended for compliance)
- Staging: 7 days retention (cost optimization)

## Troubleshooting

### Error: "Flow log already exists"

**Cause**: VPC Flow Log already exists for this VPC

**Solution**: 
1. Check if shared stack already exists:
   ```bash
   aws cloudformation describe-stacks \
     --stack-name kisanlink-erp-shared-vpc-flow-logs-staging
   ```
2. If it exists, update it instead of creating new
3. If it doesn't exist, check for manually created flow logs:
   ```bash
   aws ec2 describe-flow-logs --filter Name=resource-id,Values=$VPC_ID
   ```
4. Delete existing flow log if needed (be careful!)

### Error: "Insufficient permissions"

**Cause**: IAM role doesn't have required permissions

**Solution**: Ensure you have:
- `ec2:CreateFlowLogs`
- `ec2:DescribeFlowLogs`
- `logs:CreateLogGroup`
- `logs:PutRetentionPolicy`
- `iam:CreateRole`
- `iam:PutRolePolicy`

## Related Documentation

- [Main Infrastructure README](./README.md)
- [AWS VPC Flow Logs Documentation](https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs.html)
- [CloudWatch Logs Pricing](https://aws.amazon.com/cloudwatch/pricing/)

---

**Last Updated:** January 2025

