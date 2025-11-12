# High CPU/Memory Usage - Incident Response

## Overview
This runbook covers response procedures when CloudWatch alarms trigger for high CPU or memory utilization in ECS tasks.

## Prerequisites
- AWS CLI configured with appropriate permissions
- Access to CloudWatch console
- Access to ECS console
- SNS alert notifications configured

## Symptoms
- CloudWatch alarm: `{fpo-id}-{env}-erp-high-cpu` in ALARM state
- CloudWatch alarm: `{fpo-id}-{env}-erp-high-memory` in ALARM state
- SNS notification received
- Application performance degradation
- Increased response times

## Diagnosis

### Step 1: Verify Alarm Status
```bash
aws cloudwatch describe-alarms \
  --alarm-name-prefix <fpo-id>-<env>-erp-high-cpu \
  --query 'MetricAlarms[*].[AlarmName,StateValue,StateReason]' \
  --output table
```

### Step 2: Check Current Resource Utilization
```bash
# Get current CPU/Memory metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=<service-name> Name=ClusterName,Value=<cluster-name> \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average,Maximum \
  --output table
```

### Step 3: Check ECS Service Status
```bash
aws ecs describe-services \
  --cluster <cluster-name> \
  --services <service-name> \
  --query 'services[0].[runningCount,desiredCount,deployments[0].status]' \
  --output table
```

### Step 4: Review Application Logs
```bash
# Check for errors or unusual patterns
aws logs tail /ecs/<fpo-id>-<env>-erp \
  --since 30m \
  --filter-pattern "ERROR" \
  --format short
```

## Resolution

### Option 1: Autoscaling Should Handle It (Recommended)
If autoscaling is enabled, it should automatically scale up tasks. Verify:
```bash
# Check autoscaling policies
aws application-autoscaling describe-scaling-policies \
  --service-namespace ecs \
  --resource-id service/<cluster-name>/<service-name>

# Check scaling activities
aws application-autoscaling describe-scaling-activities \
  --service-namespace ecs \
  --resource-id service/<cluster-name>/<service-name> \
  --max-results 10
```

**Wait 5-10 minutes** for autoscaling to respond. If it doesn't, proceed to Option 2.

### Option 2: Manual Scaling
If autoscaling isn't working or needs immediate response:

```bash
# Scale up tasks (max 10 for production, 4 for staging)
aws ecs update-service \
  --cluster <cluster-name> \
  --service <service-name> \
  --desired-count <new-count> \
  --force-new-deployment
```

**Production:** Scale to 4-6 tasks  
**Staging:** Scale to 2-3 tasks

### Option 3: Increase Task Resources (If Persistent)
If high utilization persists, consider increasing task CPU/memory:

1. Update CloudFormation parameter file:
   ```json
   {
     "ParameterKey": "TaskCPU",
     "ParameterValue": "1024"  // Increase from 512
   },
   {
     "ParameterKey": "TaskMemory",
     "ParameterValue": "2048"  // Increase from 1024
   }
   ```

2. Redeploy stack:
   ```bash
   cd infrastructure
   bash scripts/deploy.sh <fpo-id> <env>
   ```

### Option 4: Application-Level Investigation
If scaling doesn't help, investigate application code:

1. Check for memory leaks
2. Review recent code changes
3. Check for inefficient queries
4. Review database connection pooling

## Verification

### Step 1: Confirm Alarm Cleared
```bash
# Wait 5-10 minutes, then check
aws cloudwatch describe-alarms \
  --alarm-name-prefix <fpo-id>-<env>-erp-high-cpu \
  --query 'MetricAlarms[0].StateValue' \
  --output text
```

Should return: `OK`

### Step 2: Verify Service Health
```bash
# Check service is stable
aws ecs describe-services \
  --cluster <cluster-name> \
  --services <service-name> \
  --query 'services[0].[runningCount,desiredCount,deployments[0].runningCount]' \
  --output table
```

### Step 3: Test Application
```bash
# Health check
curl -f https://<alb-url>/health

# Response time check
curl -w "\nTime: %{time_total}s\n" https://<alb-url>/health
```

## Prevention

1. **Monitor Trends**: Review CloudWatch dashboards weekly
2. **Capacity Planning**: Plan for expected traffic increases
3. **Load Testing**: Perform load tests before major releases
4. **Code Optimization**: Regular code reviews for performance
5. **Autoscaling Tuning**: Adjust autoscaling thresholds based on historical data

## Escalation

If issue persists after:
- Scaling up tasks
- Increasing task resources
- Application investigation

**Escalate to:**
- Backend Engineering Team Lead
- DevOps Principal
- On-call Manager

## Related Runbooks
- [Scaling Operations](../scaling.md)
- [Application Errors](./application-errors.md)
- [Log Analysis](../maintenance/log-analysis.md)

---

**Last Updated:** January 2025

