# Quick Reference Card - Kisanlink ERP Infrastructure
## For DevOps Team & On-Call Engineers

**Last Updated**: January 2025
**Version**: 1.0 (Post-P0 Implementation)

---

## 🚨 Emergency Contacts

| Role | Contact | When to Call |
|------|---------|-------------|
| DevOps Lead | devops@kisanlink.com | Stack failures, security incidents |
| SRE On-Call | [PagerDuty/Slack] | Production outages, alarms |
| Database Admin | [Contact] | RDS issues, connection problems |
| Security Team | [Contact] | Secret leaks, audit log alerts |

---

## 🔍 Quick Diagnostics

### Check Infrastructure Health (2 minutes)

```bash
# Set variables
FPO_ID="fpo001"
ENV="staging"  # or "production"
STACK_NAME="kisanlink-erp-${FPO_ID}-${ENV}"

# 1. Check stack status
aws cloudformation describe-stacks --stack-name $STACK_NAME --query 'Stacks[0].StackStatus'

# 2. Get Load Balancer URL
LB_URL=$(aws cloudformation describe-stacks --stack-name $STACK_NAME \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' --output text)

# 3. Test health endpoint
curl $LB_URL/health

# 4. Check ECS service
aws ecs describe-services \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --services ${FPO_ID}-${ENV}-erp-service \
  --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount}'

# 5. Check RDS status
aws rds describe-db-instances \
  --db-instance-identifier ${FPO_ID}-${ENV}-erp-db \
  --query 'DBInstances[0].DBInstanceStatus'
```

---

## 📊 Monitoring Locations

### CloudWatch Dashboard
```
Name: {FPO}-{ENV}-ERP-Dashboard
URL: https://console.aws.amazon.com/cloudwatch/home?region=us-west-2#dashboards:name={FPO}-{ENV}-ERP-Dashboard

Example: fpo001-staging-ERP-Dashboard
```

### CloudWatch Logs
```bash
# Tail application logs
aws logs tail /ecs/${FPO_ID}-${ENV}-erp --follow

# Search for errors (last 1 hour)
aws logs filter-log-events \
  --log-group-name /ecs/${FPO_ID}-${ENV}-erp \
  --start-time $(date -u -d '1 hour ago' +%s000) \
  --filter-pattern "ERROR"
```

### CloudTrail Audit Logs
```
Bucket: {FPO}-{ENV}-erp-audit-logs
Path: s3://{FPO}-{ENV}-erp-audit-logs/AWSLogs/{AccountID}/CloudTrail/

Query via AWS Console > CloudTrail > Event history
```

---

## 🚑 Common Issues & Fixes

### Issue 1: High CPU Alarm

**Symptoms**: Email alert "High CPU utilization"

**Quick Fix**:
1. Check CloudWatch Dashboard → ECS CPU widget
2. If sustained >80% for >10 minutes:
   ```bash
   # Scale up manually (temporary)
   aws ecs update-service \
     --cluster ${FPO_ID}-${ENV}-erp-cluster \
     --service ${FPO_ID}-${ENV}-erp-service \
     --desired-count 3  # or higher
   ```
3. Investigate root cause in logs
4. Implement P1.1 (autoscaling) if not done

**Prevention**: Implement ECS autoscaling (P1.1)

---

### Issue 2: Health Check Failing

**Symptoms**: ALB unhealthy targets, 503 errors

**Quick Fix**:
1. Check ECS task logs:
   ```bash
   aws logs tail /ecs/${FPO_ID}-${ENV}-erp --follow --filter-pattern ERROR
   ```
2. Check database connectivity:
   ```bash
   # From ECS task (via ECS Exec or logs)
   # Look for: "could not connect to database"
   ```
3. Restart ECS service:
   ```bash
   aws ecs update-service \
     --cluster ${FPO_ID}-${ENV}-erp-cluster \
     --service ${FPO_ID}-${ENV}-erp-service \
     --force-new-deployment
   ```

**Prevention**: Monitor database connections alarm

---

### Issue 3: Database Connections Exhausted

**Symptoms**: Email alert "High database connections", app errors

**Quick Fix**:
1. Check current connections:
   ```bash
   # Via RDS console or:
   aws rds describe-db-instances \
     --db-instance-identifier ${FPO_ID}-${ENV}-erp-db \
     --query 'DBInstances[0].DBInstanceStatus'
   ```
2. Identify connection leak in logs:
   ```bash
   aws logs filter-log-events \
     --log-group-name /ecs/${FPO_ID}-${ENV}-erp \
     --filter-pattern "connection pool"
   ```
3. Temporary fix - restart tasks (forces connection pool reset)
4. Long-term: Review connection pool settings in app config

**Prevention**: Implement P1.6 (read replica) for production

---

### Issue 4: Deployment Failed

**Symptoms**: CloudFormation stack in ROLLBACK_IN_PROGRESS or UPDATE_FAILED

**Quick Fix**:
1. Check stack events:
   ```bash
   aws cloudformation describe-stack-events \
     --stack-name $STACK_NAME \
     --max-items 20 \
     --query 'StackEvents[?ResourceStatus==`CREATE_FAILED` || ResourceStatus==`UPDATE_FAILED`]'
   ```
2. Common causes:
   - Secret ARN not found → Create secrets first (see README.md Step 0)
   - Subnet ID invalid → Verify VPC/subnet IDs
   - ECR image not found → Check image was pushed
3. Fix parameter file and redeploy:
   ```bash
   cd infrastructure
   bash scripts/deploy.sh ${FPO_ID} ${ENV}
   ```

**Prevention**: Validate template before deployment (`make validate`)

---

### Issue 5: Secrets Access Denied

**Symptoms**: ECS tasks failing to start, "Unable to retrieve secret"

**Quick Fix**:
1. Verify secrets exist:
   ```bash
   aws secretsmanager describe-secret \
     --secret-id ${FPO_ID}/${ENV}/erp/db-password

   aws secretsmanager describe-secret \
     --secret-id ${FPO_ID}/${ENV}/erp/jwt-secret

   aws secretsmanager describe-secret \
     --secret-id ${FPO_ID}/${ENV}/erp/aaa-jwt-secret
   ```
2. Check IAM role permissions:
   ```bash
   aws iam get-role-policy \
     --role-name ${FPO_ID}-${ENV}-erp-exec-role \
     --policy-name SecretsManagerAccess
   ```
3. If secrets missing, create them (see README.md Step 0)

**Prevention**: Always create secrets before stack deployment

---

## 🔐 Secret Management

### View Secret ARNs
```bash
aws secretsmanager list-secrets \
  --filters Key=name,Values=${FPO_ID}/${ENV}/erp \
  --query 'SecretList[*].{Name:Name,ARN:ARN}'
```

### Rotate Secret (Manual)
```bash
# Generate new secret
NEW_SECRET=$(openssl rand -base64 32)

# Update secret value
aws secretsmanager update-secret \
  --secret-id ${FPO_ID}/${ENV}/erp/jwt-secret \
  --secret-string "$NEW_SECRET"

# Force ECS task restart (picks up new secret)
aws ecs update-service \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --service ${FPO_ID}-${ENV}-erp-service \
  --force-new-deployment
```

### View Secret Value (Use sparingly!)
```bash
aws secretsmanager get-secret-value \
  --secret-id ${FPO_ID}/${ENV}/erp/db-password \
  --query 'SecretString' --output text
```

---

## 🔄 Deployment Workflows

### Normal Deployment (After Image Push)
```bash
cd infrastructure
make deploy-fpo FPO=${FPO_ID} ENV=${ENV}

# Or manual:
bash scripts/deploy.sh ${FPO_ID} ${ENV}
```

### Emergency Rollback
```bash
# Option 1: Revert to previous stack version
aws cloudformation update-stack \
  --stack-name $STACK_NAME \
  --use-previous-template \
  --parameters file://parameters/${FPO_ID}-${ENV}.json

# Option 2: Rollback ECS service to previous task definition
PREVIOUS_TASK_DEF=$(aws ecs describe-services \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --services ${FPO_ID}-${ENV}-erp-service \
  --query 'services[0].taskDefinition' --output text)

aws ecs update-service \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --service ${FPO_ID}-${ENV}-erp-service \
  --force-new-deployment \
  --task-definition $PREVIOUS_TASK_DEF
```

### Force Task Restart (No Deployment)
```bash
aws ecs update-service \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --service ${FPO_ID}-${ENV}-erp-service \
  --force-new-deployment
```

---

## 📈 Scaling Operations

### Manual ECS Scaling (Until P1.1 autoscaling implemented)
```bash
# Scale up
aws ecs update-service \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --service ${FPO_ID}-${ENV}-erp-service \
  --desired-count 4

# Scale down
aws ecs update-service \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --service ${FPO_ID}-${ENV}-erp-service \
  --desired-count 2
```

### RDS Scaling (Requires downtime)
```bash
# Scale up instance class
aws rds modify-db-instance \
  --db-instance-identifier ${FPO_ID}-${ENV}-erp-db \
  --db-instance-class db.t3.small \
  --apply-immediately

# Scale storage (online, no downtime)
aws rds modify-db-instance \
  --db-instance-identifier ${FPO_ID}-${ENV}-erp-db \
  --allocated-storage 50
```

---

## 🔍 Audit & Security

### Search CloudTrail Logs (Recent 7 days)
```bash
# Find who accessed a secret
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceName,AttributeValue=${FPO_ID}/${ENV}/erp/db-password \
  --max-results 50

# Find who modified RDS
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceType,AttributeValue=AWS::RDS::DBInstance \
  --max-results 50
```

### Generate Compliance Report
```bash
# List all resources with tags
aws resourcegroupstaggingapi get-resources \
  --tag-filters Key=FPO,Values=${FPO_ID} \
  --query 'ResourceTagMappingList[*].{Resource:ResourceARN,Tags:Tags}'

# Check encryption status
aws rds describe-db-instances \
  --query 'DBInstances[*].{Name:DBInstanceIdentifier,Encrypted:StorageEncrypted}'

aws s3api get-bucket-encryption \
  --bucket ${FPO_ID}-${ENV}-erp-attachments
```

---

## 💰 Cost Monitoring

### Check Monthly Costs
```bash
# Get cost for specific FPO (requires AWS Cost Explorer)
aws ce get-cost-and-usage \
  --time-period Start=2025-01-01,End=2025-01-31 \
  --granularity MONTHLY \
  --metrics UnblendedCost \
  --filter file://cost-filter.json

# cost-filter.json:
# {
#   "Tags": {
#     "Key": "FPO",
#     "Values": ["fpo001"]
#   }
# }
```

### Cost Optimization Checklist
- [ ] Right-size RDS instance (review CPU usage)
- [ ] Enable S3 lifecycle policies (delete old versions)
- [ ] Review CloudWatch log retention
- [ ] Stop non-production environments overnight (optional)
- [ ] Use Fargate Spot for staging (50% cost savings)

---

## 🎯 Performance Tuning

### Database Query Performance
```bash
# Enable slow query log
aws rds modify-db-instance \
  --db-instance-identifier ${FPO_ID}-${ENV}-erp-db \
  --db-parameter-group-name custom-postgres14-params \
  --apply-immediately

# View slow queries in CloudWatch Logs
aws logs tail /aws/rds/instance/${FPO_ID}-${ENV}-erp-db/postgresql --follow
```

### ECS Task Performance
```bash
# Check task metrics
aws ecs describe-services \
  --cluster ${FPO_ID}-${ENV}-erp-cluster \
  --services ${FPO_ID}-${ENV}-erp-service \
  --query 'services[0].{CPU:deployments[0].desiredCount,Memory:deployments[0].desiredCount}'
```

---

## 📋 Pre-Deployment Checklist

Before deploying **any** stack update:

- [ ] Run `make validate` (CloudFormation syntax check)
- [ ] Verify secrets exist in Secrets Manager
- [ ] Check parameter file has correct ARNs
- [ ] Review stack events for previous errors
- [ ] Notify team of maintenance window (if production)
- [ ] Backup RDS manually (optional, automated backups exist)
- [ ] Test in staging environment first
- [ ] Have rollback plan ready

---

## 📞 Escalation Matrix

| Severity | Response Time | Actions |
|----------|--------------|---------|
| P0 - Critical Outage | 15 minutes | Page on-call SRE, notify DevOps Lead |
| P1 - Major Degradation | 30 minutes | Notify on-call SRE, investigate |
| P2 - Minor Issue | 2 hours | Create ticket, investigate next business day |
| P3 - Cosmetic | Next sprint | Add to backlog |

### Severity Definitions

**P0 - Critical**:
- Production down (health check failing)
- Database unavailable
- Security breach detected
- Data loss occurring

**P1 - Major**:
- High error rate (>5% 5XX errors)
- Severe performance degradation (>5s response time)
- Single availability zone failure
- Alarms firing consistently

**P2 - Minor**:
- Occasional errors (<1% 5XX)
- Moderate performance degradation (2-5s response time)
- Non-critical alarm firing
- Staging environment issue

**P3 - Cosmetic**:
- Dashboard widget not displaying
- Documentation typo
- Log format issue
- Feature request

---

## 🛠️ Useful Commands Reference

### AWS CLI Shortcuts
```bash
# Set default region (add to ~/.bashrc)
export AWS_DEFAULT_REGION=us-west-2

# Alias for common commands
alias ecs-describe='aws ecs describe-services --cluster'
alias rds-status='aws rds describe-db-instances --query'
alias cf-events='aws cloudformation describe-stack-events --stack-name'
```

### jq Filters (JSON Parsing)
```bash
# Parse health endpoint response
curl -s $LB_URL/health | jq '.status'

# Parse ECS service status
aws ecs describe-services ... | jq '.services[0].status'

# Parse CloudFormation outputs
aws cloudformation describe-stacks ... | jq '.Stacks[0].Outputs[]'
```

---

## 📚 Reference Documentation

| Document | Location | Purpose |
|----------|----------|---------|
| Architecture Overview | `/CLAUDE.md` | Complete system architecture |
| Implementation Roadmap | `/infrastructure/IMPLEMENTATION_ROADMAP.md` | P1/P2/P3 implementation guide |
| Implementation Summary | `/infrastructure/IMPLEMENTATION_SUMMARY.md` | What was implemented in P0 |
| Quick Start Guide | `/infrastructure/QUICK_START.md` | First-time deployment |
| Full README | `/infrastructure/README.md` | Complete infrastructure docs |
| This Quick Reference | `/infrastructure/QUICK_REFERENCE.md` | You are here |

---

## 🔄 Change Log

| Date | Version | Changes |
|------|---------|---------|
| Jan 2025 | 1.0 | Initial version (Post-P0 implementation) |

---

**Keep this document updated** as infrastructure evolves!
**Print/bookmark** for quick reference during incidents.

---

**Last Reviewed**: January 2025
**Next Review**: After P1 completion
**Owner**: DevOps Team
