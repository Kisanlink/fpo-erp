# Infrastructure Implementation Roadmap
## Kisanlink ERP - Remaining P1, P2, P3 Improvements

**Generated**: January 2025
**Status**: P0 Complete (5/6 items), P1-P3 Pending
**Current Compliance Score**: 33/50 (66%)
**Target Score**: 50/50 (100%)

---

## ✅ Phase 1: P0 - Critical Security & Monitoring (COMPLETE)

### Completed Items

1. **✅ P0.1: Remove Secrets from Parameter Files**
   - Status: COMPLETE
   - Files: 6 modified
   - Impact: Critical security enhancement
   - Implementation: Secrets Manager ARNs instead of plaintext

2. **✅ P0.2: Add CloudWatch SNS Alerts**
   - Status: COMPLETE
   - Files: `erp-application.yaml`
   - Impact: Email notifications for CPU, Memory, DB alarms
   - Implementation: SNS topic with email subscription

3. **✅ P0.3: Enable CloudTrail Audit Logging**
   - Status: COMPLETE
   - Files: `erp-application.yaml`
   - Impact: 7-year audit trail, S3/Secrets monitoring
   - Implementation: CloudTrail + S3 bucket + policy

4. **✅ P0.4: Add Integration Tests to CI/CD**
   - Status: COMPLETE
   - Files: `build-and-push.yml`
   - Impact: PostgreSQL integration testing
   - Implementation: Docker PostgreSQL 14.10 in GitHub Actions

5. **✅ P0.5: Implement CloudWatch Dashboard**
   - Status: COMPLETE
   - Files: `erp-application.yaml`
   - Impact: 11-widget comprehensive monitoring
   - Implementation: CloudWatch Dashboard with ECS/ALB/RDS metrics

6. **⏭️ P0.6: Add JWT Secret Rotation**
   - Status: SKIPPED (lower priority than P1 items)
   - Reason: Requires Lambda function + rotation schedule (complex)
   - Can be added later without blocking P1-P3

---

## 🔄 Phase 2: P1 - High Availability & Security (NEXT PRIORITY)

### P1.1: Add ECS Autoscaling Policies ⚡ HIGH PRIORITY

**File**: `infrastructure/templates/erp-application.yaml`

**Add after line 730 (after ECSService)**:

```yaml
  # ========================================
  # ECS Autoscaling
  # ========================================
  ECSScalableTarget:
    Type: AWS::ApplicationAutoScaling::ScalableTarget
    Properties:
      MaxCapacity: !If [IsProduction, 10, 4]
      MinCapacity: !If [IsProduction, 2, 1]
      ResourceId: !Sub 'service/${ECSCluster}/${ECSService.Name}'
      RoleARN: !Sub 'arn:aws:iam::${AWS::AccountId}:role/aws-service-role/ecs.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_ECSService'
      ScalableDimension: ecs:service:DesiredCount
      ServiceNamespace: ecs

  ECSScalingPolicyCPU:
    Type: AWS::ApplicationAutoScaling::ScalingPolicy
    Properties:
      PolicyName: !Sub '${FPOIdentifier}-${Environment}-erp-cpu-scaling'
      PolicyType: TargetTrackingScaling
      ScalingTargetId: !Ref ECSScalableTarget
      TargetTrackingScalingPolicyConfiguration:
        TargetValue: 70.0
        PredefinedMetricSpecification:
          PredefinedMetricType: ECSServiceAverageCPUUtilization
        ScaleInCooldown: 300
        ScaleOutCooldown: 60

  ECSScalingPolicyMemory:
    Type: AWS::ApplicationAutoScaling::ScalingPolicy
    Properties:
      PolicyName: !Sub '${FPOIdentifier}-${Environment}-erp-memory-scaling'
      PolicyType: TargetTrackingScaling
      ScalingTargetId: !Ref ECSScalableTarget
      TargetTrackingScalingPolicyConfiguration:
        TargetValue: 70.0
        PredefinedMetricSpecification:
          PredefinedMetricType: ECSServiceAverageMemoryUtilization
        ScaleInCooldown: 300
        ScaleOutCooldown: 60
```

**Impact**:
- Auto-scales based on CPU (70% target) and Memory (70% target)
- Production: 2-10 tasks, Staging: 1-4 tasks
- Scale-out: 60s cooldown, Scale-in: 300s cooldown
- Handles traffic spikes automatically

---

### P1.2: Implement Docker Image Scanning ⚡ HIGH PRIORITY

**File**: `.github/workflows/build-and-push.yml`

**Add after line 97 (after "Build Docker image" step)**:

```yaml
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ steps.login-ecr.outputs.registry }}/${{ env.ECR_REPOSITORY }}:${{ env.IMAGE_TAG }}
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Fail build on HIGH/CRITICAL vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ steps.login-ecr.outputs.registry }}/${{ env.ECR_REPOSITORY }}:${{ env.IMAGE_TAG }}
          format: 'table'
          exit-code: '1'
          severity: 'CRITICAL,HIGH'
```

**Impact**:
- Scans Docker images for CVEs before push
- Uploads results to GitHub Security tab
- Fails build if CRITICAL/HIGH vulnerabilities found
- Prevents vulnerable images from reaching production

---

### P1.3: Add Smoke Tests After Deployment ⚡ HIGH PRIORITY

**File**: `.github/workflows/deploy-erp.yml`

**Add after health check step** (create comprehensive smoke test suite):

```yaml
      - name: Run smoke tests
        run: |
          echo "Running smoke tests against deployed ERP..."

          LB_URL="${{ steps.get-lb-url.outputs.url }}"

          # Test 1: Health endpoint
          echo "Test 1: Health endpoint..."
          HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $LB_URL/health)
          if [ "$HEALTH_RESPONSE" != "200" ]; then
            echo "❌ Health check failed (HTTP $HEALTH_RESPONSE)"
            exit 1
          fi
          echo "✅ Health check passed"

          # Test 2: API root endpoint
          echo "Test 2: API root endpoint..."
          API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $LB_URL/api/v1)
          if [ "$API_RESPONSE" != "200" ] && [ "$API_RESPONSE" != "401" ]; then
            echo "❌ API root failed (HTTP $API_RESPONSE)"
            exit 1
          fi
          echo "✅ API root accessible"

          # Test 3: Database connectivity (via health endpoint with db check)
          echo "Test 3: Database connectivity..."
          DB_STATUS=$(curl -s $LB_URL/health | jq -r '.database')
          if [ "$DB_STATUS" != "connected" ]; then
            echo "❌ Database connection failed"
            exit 1
          fi
          echo "✅ Database connected"

          # Test 4: Response time check
          echo "Test 4: Response time check..."
          RESPONSE_TIME=$(curl -s -o /dev/null -w "%{time_total}" $LB_URL/health)
          if (( $(echo "$RESPONSE_TIME > 2.0" | bc -l) )); then
            echo "⚠️  Warning: Slow response time ($RESPONSE_TIME seconds)"
          else
            echo "✅ Response time acceptable ($RESPONSE_TIME seconds)"
          fi

          echo "🎉 All smoke tests passed!"
```

**Impact**:
- Validates deployment before marking as successful
- Tests health, API, database, and performance
- Prevents deploying broken versions
- Immediate feedback on deployment issues

---

### P1.4: Configure ALB Error Rate Alarms ⚡ CRITICAL

**File**: `infrastructure/templates/erp-application.yaml`

**Add after DatabaseConnectionsAlarm (line ~796)**:

```yaml
  ALBTargetResponseTimeAlarm:
    Type: AWS::CloudWatch::Alarm
    Properties:
      AlarmName: !Sub '${FPOIdentifier}-${Environment}-erp-alb-response-time'
      AlarmDescription: !Sub 'High response time for ${FPOIdentifier} ${Environment} ERP ALB'
      MetricName: TargetResponseTime
      Namespace: AWS/ApplicationELB
      Statistic: Average
      Period: 300
      EvaluationPeriods: 2
      Threshold: 2
      ComparisonOperator: GreaterThanThreshold
      AlarmActions:
        - !Ref AlertTopic
      OKActions:
        - !Ref AlertTopic
      Dimensions:
        - Name: LoadBalancer
          Value: !GetAtt ApplicationLoadBalancer.LoadBalancerFullName

  ALB5XXErrorAlarm:
    Type: AWS::CloudWatch::Alarm
    Properties:
      AlarmName: !Sub '${FPOIdentifier}-${Environment}-erp-alb-5xx-errors'
      AlarmDescription: !Sub 'High 5XX error rate for ${FPOIdentifier} ${Environment} ERP ALB'
      MetricName: HTTPCode_Target_5XX_Count
      Namespace: AWS/ApplicationELB
      Statistic: Sum
      Period: 300
      EvaluationPeriods: 1
      Threshold: 10
      ComparisonOperator: GreaterThanThreshold
      AlarmActions:
        - !Ref AlertTopic
      OKActions:
        - !Ref AlertTopic
      Dimensions:
        - Name: LoadBalancer
          Value: !GetAtt ApplicationLoadBalancer.LoadBalancerFullName

  ALB4XXErrorAlarm:
    Type: AWS::CloudWatch::Alarm
    Properties:
      AlarmName: !Sub '${FPOIdentifier}-${Environment}-erp-alb-4xx-errors'
      AlarmDescription: !Sub 'High 4XX error rate for ${FPOIdentifier} ${Environment} ERP ALB'
      MetricName: HTTPCode_Target_4XX_Count
      Namespace: AWS/ApplicationELB
      Statistic: Sum
      Period: 300
      EvaluationPeriods: 2
      Threshold: 50
      ComparisonOperator: GreaterThanThreshold
      AlarmActions:
        - !Ref AlertTopic
      OKActions:
        - !Ref AlertTopic
      Dimensions:
        - Name: LoadBalancer
          Value: !GetAtt ApplicationLoadBalancer.LoadBalancerFullName

  ALBUnhealthyTargetAlarm:
    Type: AWS::CloudWatch::Alarm
    Properties:
      AlarmName: !Sub '${FPOIdentifier}-${Environment}-erp-alb-unhealthy-targets'
      AlarmDescription: !Sub 'Unhealthy targets detected for ${FPOIdentifier} ${Environment} ERP ALB'
      MetricName: UnHealthyHostCount
      Namespace: AWS/ApplicationELB
      Statistic: Average
      Period: 60
      EvaluationPeriods: 2
      Threshold: 1
      ComparisonOperator: GreaterThanOrEqualToThreshold
      AlarmActions:
        - !Ref AlertTopic
      OKActions:
        - !Ref AlertTopic
      Dimensions:
        - Name: TargetGroup
          Value: !GetAtt ALBTargetGroup.TargetGroupFullName
        - Name: LoadBalancer
          Value: !GetAtt ApplicationLoadBalancer.LoadBalancerFullName
```

**Impact**:
- Response time > 2 seconds → Alert
- 5XX errors > 10 in 5 minutes → Alert
- 4XX errors > 50 in 5 minutes → Alert (sustained client errors)
- Any unhealthy target → Immediate alert

---

### P1.5: Implement Application-Level Metrics ⚡ MEDIUM PRIORITY

**Create new file**: `internal/monitoring/cloudwatch.go`

```go
package monitoring

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
    "time"
)

type CloudWatchClient struct {
    client    *cloudwatch.Client
    namespace string
}

func NewCloudWatchClient(cfg aws.Config, namespace string) *CloudWatchClient {
    return &CloudWatchClient{
        client:    cloudwatch.NewFromConfig(cfg),
        namespace: namespace,
    }
}

// PublishSaleMetric publishes a sale transaction metric
func (c *CloudWatchClient) PublishSaleMetric(amount float64, itemCount int, environment string) error {
    _, err := c.client.PutMetricData(context.TODO(), &cloudwatch.PutMetricDataInput{
        Namespace: aws.String(c.namespace),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String("SaleAmount"),
                Value:      aws.Float64(amount),
                Unit:       types.StandardUnitNone,
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []types.Dimension{
                    {Name: aws.String("Environment"), Value: aws.String(environment)},
                },
            },
            {
                MetricName: aws.String("SaleItemCount"),
                Value:      aws.Float64(float64(itemCount)),
                Unit:       types.StandardUnitCount,
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []types.Dimension{
                    {Name: aws.String("Environment"), Value: aws.String(environment)},
                },
            },
        },
    })
    return err
}

// PublishInventoryMetric publishes inventory batch operations
func (c *CloudWatchClient) PublishInventoryMetric(operation string, quantity int, environment string) error {
    _, err := c.client.PutMetricData(context.TODO(), &cloudwatch.PutMetricDataInput{
        Namespace: aws.String(c.namespace),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String("InventoryOperation"),
                Value:      aws.Float64(float64(quantity)),
                Unit:       types.StandardUnitCount,
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []types.Dimension{
                    {Name: aws.String("Environment"), Value: aws.String(environment)},
                    {Name: aws.String("Operation"), Value: aws.String(operation)},
                },
            },
        },
    })
    return err
}

// PublishErrorMetric publishes application errors
func (c *CloudWatchClient) PublishErrorMetric(errorType string, environment string) error {
    _, err := c.client.PutMetricData(context.TODO(), &cloudwatch.PutMetricDataInput{
        Namespace: aws.String(c.namespace),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String("ApplicationError"),
                Value:      aws.Float64(1),
                Unit:       types.StandardUnitCount,
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []types.Dimension{
                    {Name: aws.String("Environment"), Value: aws.String(environment)},
                    {Name: aws.String("ErrorType"), Value: aws.String(errorType)},
                },
            },
        },
    })
    return err
}
```

**Usage in services** (example: `internal/services/sales_service.go`):

```go
// In CreateSale function, after successful sale creation:
if s.cwClient != nil {
    go s.cwClient.PublishSaleMetric(
        result.GrandTotal,
        len(result.Items),
        os.Getenv("SERVER_MODE"),
    )
}
```

**Update CloudWatch Dashboard** to include business metrics widgets.

**Impact**:
- Track sales volume, revenue, inventory operations
- Business-level insights in CloudWatch
- Correlate business metrics with infrastructure metrics
- Foundation for advanced analytics

---

### P1.6: Add RDS Read Replica ⚡ HIGH PRIORITY (Production Only)

**File**: `infrastructure/templates/erp-application.yaml`

**Add after DBInstance (line ~273)**:

```yaml
  DBReadReplica:
    Type: AWS::RDS::DBInstance
    Condition: IsProduction
    Properties:
      DBInstanceIdentifier: !Sub '${FPOIdentifier}-${Environment}-erp-db-replica'
      SourceDBInstanceIdentifier: !Ref DBInstance
      DBInstanceClass: !Ref DBInstanceClass
      PubliclyAccessible: false
      Tags:
        - Key: Name
          Value: !Sub '${FPOIdentifier}-${Environment}-erp-db-replica'
        - Key: Environment
          Value: !Ref Environment
        - Key: FPO
          Value: !Ref FPOIdentifier
        - Key: Role
          Value: ReadReplica
```

**Add to Outputs**:

```yaml
  DBReadReplicaEndpoint:
    Condition: IsProduction
    Description: 'RDS read replica endpoint (production only)'
    Value: !GetAtt DBReadReplica.Endpoint.Address
    Export:
      Name: !Sub '${AWS::StackName}-DBReadReplicaEndpoint'
```

**Impact**:
- Read scaling for production workloads
- Offload reporting/analytics queries to replica
- Improved production database performance
- Zero downtime for read queries during maintenance

---

### P1.7: Configure AWS Backup Plan ⚡ HIGH PRIORITY

**File**: `infrastructure/templates/erp-application.yaml`

**Add after AuditTrail (line ~385)**:

```yaml
  # ========================================
  # AWS Backup Configuration
  # ========================================
  BackupVault:
    Type: AWS::Backup::BackupVault
    Properties:
      BackupVaultName: !Sub '${FPOIdentifier}-${Environment}-erp-backup-vault'
      EncryptionKeyArn: !GetAtt BackupKMSKey.Arn

  BackupKMSKey:
    Type: AWS::KMS::Key
    Properties:
      Description: !Sub 'KMS key for ${FPOIdentifier} ${Environment} ERP backups'
      KeyPolicy:
        Version: '2012-10-17'
        Statement:
          - Sid: Enable IAM policies
            Effect: Allow
            Principal:
              AWS: !Sub 'arn:aws:iam::${AWS::AccountId}:root'
            Action: 'kms:*'
            Resource: '*'
          - Sid: Allow AWS Backup
            Effect: Allow
            Principal:
              Service: backup.amazonaws.com
            Action:
              - 'kms:CreateGrant'
              - 'kms:DescribeKey'
            Resource: '*'

  BackupPlan:
    Type: AWS::Backup::BackupPlan
    Properties:
      BackupPlan:
        BackupPlanName: !Sub '${FPOIdentifier}-${Environment}-erp-backup-plan'
        BackupPlanRule:
          - RuleName: DailyBackup
            TargetBackupVault: !Ref BackupVault
            ScheduleExpression: 'cron(0 3 * * ? *)'  # 3 AM UTC daily
            StartWindowMinutes: 60
            CompletionWindowMinutes: 120
            Lifecycle:
              DeleteAfterDays: !If [IsProduction, 30, 7]
              MoveToColdStorageAfterDays: !If [IsProduction, 7, !Ref 'AWS::NoValue']
          - RuleName: WeeklyBackup
            Condition: IsProduction
            TargetBackupVault: !Ref BackupVault
            ScheduleExpression: 'cron(0 4 ? * SUN *)'  # 4 AM UTC every Sunday
            StartWindowMinutes: 60
            CompletionWindowMinutes: 180
            Lifecycle:
              DeleteAfterDays: 90
              MoveToColdStorageAfterDays: 30

  BackupSelection:
    Type: AWS::Backup::BackupSelection
    Properties:
      BackupPlanId: !Ref BackupPlan
      BackupSelection:
        SelectionName: !Sub '${FPOIdentifier}-${Environment}-erp-resources'
        IamRoleArn: !Sub 'arn:aws:iam::${AWS::AccountId}:role/service-role/AWSBackupDefaultServiceRole'
        Resources:
          - !Sub 'arn:aws:rds:${AWS::Region}:${AWS::AccountId}:db:${DBInstance}'
        Conditions:
          StringEquals:
            - ConditionKey: 'aws:ResourceTag/FPO'
              ConditionValue: !Ref FPOIdentifier
```

**Impact**:
- Daily backups at 3 AM UTC
- Weekly backups on Sunday (production only)
- 30-day retention (production) / 7-day (staging)
- Cold storage after 7 days (cost optimization)
- Encrypted backups via KMS
- Centralized backup management

---

## 🔧 Phase 3: P2 - Hardening & Optimization (FUTURE)

### P2.1: Add AWS WAF Protection

- Protects against SQL injection, XSS, rate limiting
- Managed rule sets (OWASP Top 10, Known bad inputs)
- Geographic restrictions if needed
- Bot control

### P2.2: Implement Distributed Tracing (AWS X-Ray)

- Request tracing across ECS, ALB, RDS
- Performance bottleneck identification
- Microservice dependency mapping
- Latency analysis

### P2.3: Create docker-compose.yml

- Local development environment
- PostgreSQL, Redis, MinIO (S3 emulation)
- AAA service mock (optional)
- One-command local setup

### P2.4: Add Local Setup Script

- `infrastructure/scripts/local-setup.sh`
- Automated Docker Compose setup
- Database migrations
- Seed data loading
- Health checks

### P2.5: Enable VPC Flow Logs

- Network traffic monitoring
- Security analysis
- Troubleshooting connectivity issues
- Compliance requirement

### P2.6: Configure AWS Config Rules

- Compliance monitoring (encrypted volumes, public access)
- Configuration drift detection
- Resource inventory
- Security best practices enforcement

### P2.7: Restrict CORS Origins

- Replace `CORS_ALLOWED_ORIGINS: '*'` with specific domains
- Production: `https://erp.kisanlink.com,https://admin.kisanlink.com`
- Staging: `https://erp-staging.kisanlink.com`
- Prevents unauthorized API access

---

## 🎯 Phase 4: P3 - Excellence & Best Practices (OPTIONAL)

### P3.1: Define SLI/SLO Metrics

Create `docs/SLI_SLO.md`:
- Availability SLO: 99.9% uptime
- Latency SLO: p99 < 500ms
- Error rate SLO: < 0.1%
- Dashboard with SLI tracking

### P3.2: Add VPC Endpoint for S3

- Private connectivity to S3
- No data transfer charges within AZ
- Enhanced security (no internet gateway needed)
- Improved performance

### P3.3: Implement Blue-Green Deployments

- Zero-downtime deployments
- Instant rollback capability
- Traffic shifting via ALB weighted targets
- Automated validation before cutover

### P3.4: Create Runbook Documentation

Create `docs/RUNBOOKS/`:
- `high-cpu-incident.md`
- `database-connection-issue.md`
- `deployment-rollback.md`
- `secrets-rotation.md`
- Step-by-step incident response procedures

---

## 📈 Expected Compliance Scores After Implementation

| Phase | Score Before | Score After | Items | Impact |
|-------|-------------|-------------|-------|--------|
| P0 (Complete) | 17/50 (34%) | 33/50 (66%) | 5/6 | Security, Monitoring |
| P1 (Next) | 33/50 (66%) | 44/50 (88%) | 7/7 | HA, Performance |
| P2 (Future) | 44/50 (88%) | 48/50 (96%) | 7/7 | Hardening, DX |
| P3 (Optional) | 48/50 (96%) | 50/50 (100%) | 4/4 | Excellence |

---

## 🚀 Implementation Priority Order

**Immediate (This Week)**:
1. ✅ P0.1-P0.5 (DONE)
2. P1.4: ALB Error Rate Alarms (30 min)
3. P1.1: ECS Autoscaling (30 min)
4. P1.2: Docker Image Scanning (20 min)

**High Priority (Next Week)**:
5. P1.3: Smoke Tests (1 hour)
6. P1.6: RDS Read Replica (30 min)
7. P1.7: AWS Backup Plan (45 min)

**Medium Priority (Month 1)**:
8. P1.5: Application Metrics (2-3 hours)
9. P2.1: AWS WAF (1 hour)
10. P2.7: Restrict CORS (10 min)

**Future (Month 2+)**:
- P2.2-P2.6: Hardening items
- P3.1-P3.4: Excellence items

---

## 📝 Testing Strategy

After each implementation:
1. ✅ Validate CloudFormation template syntax
2. ✅ Deploy to test/staging environment first
3. ✅ Verify new resources in AWS Console
4. ✅ Test alarms by triggering conditions
5. ✅ Update documentation
6. ✅ Git commit with descriptive message
7. ✅ Deploy to production only after staging validation

---

## 🎓 Knowledge Transfer

**Team Training Required**:
- CloudWatch Dashboard usage
- SNS alert response procedures
- CloudTrail log analysis
- Secret rotation workflows
- Backup restoration procedures
- Autoscaling behavior understanding

**Documentation to Create**:
- Operational runbooks
- Incident response procedures
- Monitoring alert playbooks
- Disaster recovery plan

---

## 📞 Support & Escalation

**For Issues During Implementation**:
1. Check CloudFormation stack events
2. Review CloudWatch logs
3. Verify IAM permissions
4. Consult AWS documentation
5. Contact DevOps team lead

**Rollback Procedures**:
- CloudFormation: Previous stack version
- Docker: Previous ECR image tag
- Database: AWS Backup restoration

---

**Document Version**: 1.0
**Last Updated**: January 2025
**Owner**: DevOps Team
**Next Review**: After P1 completion
