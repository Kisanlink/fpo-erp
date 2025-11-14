# Branch Protection Requirements

## Overview
This document outlines the branch protection rules and requirements for the Kisanlink ERP repository to ensure code quality and security.

## Protected Branches

### Main Branch (`main`)
- **Purpose**: Production-ready code
- **Protection Level**: Maximum
- **Deployment**: Auto-deploys to production

### Development Branch (`development`)
- **Purpose**: Integration and staging
- **Protection Level**: High
- **Deployment**: Auto-deploys to staging

### Staging Branch (`staging`)
- **Purpose**: Pre-production testing
- **Protection Level**: Medium
- **Deployment**: Manual deployment

## Required Branch Protection Rules

### 1. Require Pull Request Reviews

**Settings:**
- ✅ Require a pull request before merging
- ✅ Required number of approvals: **1** (for main), **1** (for development)
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from Code Owners (if CODEOWNERS file exists)

**Approvers:**
- Backend Engineering Team Lead
- DevOps Principal (for infrastructure changes)
- Security Team (for security-related changes)

### 2. Require Status Checks to Pass

**Required Status Checks:**
- ✅ `validate-iac` - Infrastructure validation
- ✅ `validate-dockerfile` - Dockerfile linting (if Dockerfile changed)
- ✅ `build-and-push` - Build and test pipeline
- ✅ `validate-pr` - PR validation workflow

**Settings:**
- ✅ Require branches to be up to date before merging
- ✅ Require status checks to pass before merging

### 3. Require Conversation Resolution

**Settings:**
- ✅ Require all conversations on code to be resolved before merging
- ✅ Require linear history (no merge commits)

### 4. Restrict Who Can Push

**Settings:**
- ✅ Restrict pushes that create matching branches
- ✅ Do not allow force pushes
- ✅ Do not allow deletions

**Allowed Actions:**
- Only repository administrators can bypass protection
- All other users must use pull requests

### 5. Require Signed Commits (Optional but Recommended)

**Settings:**
- ⚠️ Require signed commits (if GPG signing is enforced)
- This prevents commit spoofing

## GitHub Actions Integration

### PR Validation Workflow
The `.github/workflows/validate-pr.yml` workflow automatically runs on pull requests and must pass before merging.

**Checks Performed:**
1. CloudFormation template validation
2. Security scanning (Checkov)
3. Dockerfile linting (Hadolint)
4. YAML syntax validation

### Build and Push Workflow
The `.github/workflows/build-and-push.yml` workflow runs on:
- Push to main/development branches
- Pull requests to main/development
- Manual workflow dispatch

**Checks Performed:**
1. Unit tests
2. Integration tests
3. Docker image build
4. Trivy vulnerability scanning
5. Image push to ECR

## Code Review Guidelines

### What Requires Review

**Mandatory Review Required:**
- Infrastructure changes (CloudFormation, Terraform)
- Security-related changes (IAM, secrets, authentication)
- Database migrations
- API changes
- Configuration changes
- Dockerfile changes

**Recommended Review:**
- Business logic changes
- Test updates
- Documentation updates

### Review Checklist

**For Infrastructure Changes:**
- [ ] CloudFormation template validated
- [ ] Security scanning passed (Checkov)
- [ ] No hardcoded secrets
- [ ] IAM permissions follow least privilege
- [ ] Cost implications reviewed
- [ ] Rollback plan documented

**For Application Changes:**
- [ ] Tests added/updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Documentation updated
- [ ] Backward compatibility maintained

## Enforcement

### Automatic Enforcement
- GitHub branch protection rules enforce all requirements
- GitHub Actions workflows must pass
- Status checks block merging if failed

### Manual Enforcement
- Repository administrators review exceptions
- Security team approval for security changes
- DevOps team approval for infrastructure changes

## Exceptions

### Emergency Hotfixes
For critical production issues:
1. Create hotfix branch from main
2. Make minimal fix
3. Create PR with "EMERGENCY" label
4. Get expedited review (1 approver)
5. Deploy immediately after merge
6. Follow up with proper PR within 24 hours

### Documentation Only Changes
- Documentation-only PRs may be merged with single approval
- No code changes allowed in documentation PRs

## Compliance

### Audit Trail
- All merges are logged in GitHub audit log
- CloudTrail logs infrastructure changes
- All deployments are tracked in GitHub Actions

### Regular Reviews
- Monthly review of branch protection rules
- Quarterly audit of merge patterns
- Annual security review of access controls

## Setup Instructions

### For Repository Administrators

1. **Navigate to Repository Settings**
   - Go to repository → Settings → Branches

2. **Add Branch Protection Rule**
   - Click "Add rule"
   - Branch name pattern: `main`
   - Configure all required settings above

3. **Repeat for Other Branches**
   - `development`
   - `staging`

4. **Configure Required Status Checks**
   - Add all required status checks
   - Ensure "Require branches to be up to date" is enabled

5. **Set Up CODEOWNERS File** (Optional)
   ```
   # Infrastructure
   infrastructure/ @devops-team
   
   # Security
   internal/aaa/ @security-team
   
   # Application
   internal/ @backend-team
   ```

## Monitoring

### Metrics to Track
- Average time to merge PR
- Number of failed status checks
- Number of force pushes (should be 0)
- Number of bypassed protections (should be minimal)

### Alerts
- Failed status checks on main/development
- Force push attempts
- Protection bypass events

## Related Documentation
- [Deployment Procedures](./ops/runbooks/deployment.md)
- [Security Guidelines](./SECURITY.md) (if exists)
- [Contributing Guide](./CONTRIBUTING.md) (if exists)

---

**Last Updated:** January 2025  
**Maintained By:** DevOps Team

