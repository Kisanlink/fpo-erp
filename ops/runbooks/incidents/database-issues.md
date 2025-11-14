# Database Issues - Incident Response

## Overview
This runbook covers response procedures for RDS database connectivity, performance, and availability issues.

## Prerequisites
- AWS CLI configured
- RDS access permissions
- Database connection credentials (from Secrets Manager)
- CloudWatch access

## Symptoms
- CloudWatch alarm: `{fpo-id}-{env}-erp-rds-high-cpu` in ALARM
- CloudWatch alarm: `{fpo-id}-{env}-erp-rds-high-connections` in ALARM
- Application errors: "database connection failed"
- Slow query performance
- Health check failures

## Diagnosis

### Step 1: Check RDS Status
```bash
aws rds describe-db-instances \
  --db-instance-identifier <db-instance-id> \
  --query 'DBInstances[0].[DBInstanceStatus,DBInstanceClass,Endpoint.Address]' \
  --output table
```

### Step 2: Check CloudWatch Metrics
```bash
# CPU Utilization
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value=<db-instance-id> \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average,Maximum \
  --output table

# Database Connections
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name DatabaseConnections \
  --dimensions Name=DBInstanceIdentifier,Value=<db-instance-id> \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average,Maximum \
  --output table
```

### Step 3: Check Application Logs
```bash
# Look for database errors
aws logs tail /ecs/<fpo-id>-<env>-erp \
  --since 30m \
  --filter-pattern "database\|connection\|postgres\|sql" \
  --format short
```

### Step 4: Test Database Connectivity
```bash
# Get database endpoint from CloudFormation
aws cloudformation describe-stacks \
  --stack-name kisanlink-erp-<fpo-id>-<env> \
  --query 'Stacks[0].Outputs[?OutputKey==`DatabaseEndpoint`].OutputValue' \
  --output text

# Test connection (requires psql)
psql -h <endpoint> -U <username> -d <database> -c "SELECT 1;"
```

## Resolution

### Issue: High CPU Utilization

**Option 1: Check for Long-Running Queries**
```sql
-- Connect to database and run:
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
  AND state = 'active';
```

**Option 2: Scale Up Database Instance**
1. Update CloudFormation parameter:
   ```json
   {
     "ParameterKey": "DBInstanceClass",
     "ParameterValue": "db.t3.small"  // Upgrade from db.t3.micro
   }
   ```
2. Redeploy stack (will cause brief downtime)

**Option 3: Enable Read Replica (Production Only)**
- Read replica should already exist in production
- Verify application is using read replica for read queries

### Issue: High Connection Count

**Option 1: Check Connection Pool Settings**
- Review application connection pool configuration
- Ensure `max_open_conns` is reasonable (100-200)
- Check for connection leaks

**Option 2: Increase Max Connections**
```sql
-- Check current setting
SHOW max_connections;

-- Increase (requires parameter group modification)
-- This should be done via CloudFormation parameter group
```

### Issue: Database Unavailable

**Step 1: Check RDS Events**
```bash
aws rds describe-events \
  --source-identifier <db-instance-id> \
  --source-type db-instance \
  --max-items 10 \
  --output table
```

**Step 2: Check Multi-AZ Status**
```bash
aws rds describe-db-instances \
  --db-instance-identifier <db-instance-id> \
  --query 'DBInstances[0].[MultiAZ,AvailabilityZone,SecondaryAvailabilityZone]' \
  --output table
```

**Step 3: Failover to Standby (If Multi-AZ)**
```bash
# Force failover (causes brief downtime)
aws rds reboot-db-instance \
  --db-instance-identifier <db-instance-id> \
  --force-failover
```

### Issue: Slow Queries

**Step 1: Enable Performance Insights**
```bash
# Check if enabled
aws rds describe-db-instances \
  --db-instance-identifier <db-instance-id> \
  --query 'DBInstances[0].PerformanceInsightsEnabled' \
  --output text
```

**Step 2: Review Slow Query Log**
- Enable slow query log in parameter group
- Review CloudWatch Logs for slow queries
- Optimize identified queries

## Verification

### Step 1: Confirm Metrics Normalized
```bash
# Wait 10-15 minutes, then check
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value=<db-instance-id> \
  --start-time $(date -u -d '15 minutes ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average \
  --query 'Datapoints[*].Average' \
  --output text
```

Should be < 70%

### Step 2: Test Application Connectivity
```bash
# Health check
curl -f https://<alb-url>/health

# Should return database status: "healthy"
```

### Step 3: Verify Read Replica (Production)
```bash
# Check read replica status
aws rds describe-db-instances \
  --db-instance-identifier <read-replica-id> \
  --query 'DBInstances[0].DBInstanceStatus' \
  --output text
```

## Prevention

1. **Regular Monitoring**: Review RDS metrics weekly
2. **Connection Pooling**: Ensure proper connection pool configuration
3. **Query Optimization**: Regular query performance reviews
4. **Capacity Planning**: Monitor storage and connection trends
5. **Backup Verification**: Test restore procedures quarterly

## Escalation

If issue persists after:
- Scaling database
- Connection pool adjustments
- Query optimization

**Escalate to:**
- Database Administrator
- Backend Engineering Team Lead
- DevOps Principal

## Related Runbooks
- [Backup and Restore](../maintenance/backup-restore.md)
- [Application Errors](./application-errors.md)
- [Service Recovery](../emergency/service-recovery.md)

---

**Last Updated:** January 2025

