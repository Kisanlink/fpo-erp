# Operations Runbooks

This directory contains operational runbooks for common tasks and incident response procedures for the Kisanlink ERP system.

## Runbook Index

### Infrastructure Operations
- [Deployment Procedures](./deployment.md) - Step-by-step deployment guide
- [Rollback Procedures](./rollback.md) - How to rollback a failed deployment
- [Scaling Operations](./scaling.md) - Manual scaling procedures

### Incident Response
- [High CPU/Memory](./incidents/high-resource-usage.md) - Response to high resource utilization
- [Database Issues](./incidents/database-issues.md) - RDS connectivity and performance issues
- [Application Errors](./incidents/application-errors.md) - 5xx errors and application failures
- [ALB Issues](./incidents/alb-issues.md) - Load balancer and routing issues

### Maintenance
- [Database Backup and Restore](./maintenance/backup-restore.md) - Backup verification and restore procedures
- [Log Analysis](./maintenance/log-analysis.md) - How to analyze CloudWatch logs
- [Cost Optimization](./maintenance/cost-optimization.md) - Cost monitoring and optimization

### Emergency Procedures
- [Service Recovery](./emergency/service-recovery.md) - Complete service recovery procedures
- [Data Recovery](./emergency/data-recovery.md) - Data loss recovery procedures

---

## Quick Reference

### Common Commands

```bash
# Check ECS service status
aws ecs describe-services \
  --cluster <cluster-name> \
  --services <service-name>

# View recent CloudWatch logs
aws logs tail /ecs/<fpo-id>-<env>-erp --follow

# Check alarm status
aws cloudwatch describe-alarms \
  --alarm-name-prefix <fpo-id>-<env>-erp

# Scale ECS service manually
aws ecs update-service \
  --cluster <cluster-name> \
  --service <service-name> \
  --desired-count <count>
```

---

## Runbook Format

Each runbook follows this structure:

1. **Overview** - What this runbook covers
2. **Prerequisites** - Required access and tools
3. **Symptoms** - How to identify the issue
4. **Diagnosis** - Steps to diagnose the problem
5. **Resolution** - Step-by-step resolution
6. **Verification** - How to verify the fix
7. **Prevention** - How to prevent recurrence

---

**Last Updated:** January 2025

