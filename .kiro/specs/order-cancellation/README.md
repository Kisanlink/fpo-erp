# Order Cancellation with Inventory Return

**Feature**: Cancel sales orders and automatically return inventory to batches
**Status**: Assessment Complete - Awaiting Design Document
**Priority**: High
**Estimated Effort**: 3-5 days

---

## Overview

This feature allows users to cancel pending sales orders, which automatically returns the allocated inventory back to the respective inventory batches. The implementation follows the existing purchase order cancellation pattern and maintains full audit trail through inventory transactions.

**Key Benefits**:
- Prevents inventory leakage from cancelled orders
- Maintains accurate stock levels across warehouses
- Provides complete audit trail for compliance
- Follows FEFO (First-Expired-First-Out) reversal logic

---

## Documentation Structure

### 1. Technical Assessment (`technical-assessment.md`)
**Purpose**: Comprehensive codebase analysis and implementation blueprint

**Contents**:
- Current system analysis (models, services, repositories)
- Detailed file-by-file change requirements
- Existing patterns to follow
- Technical debt identified
- Security considerations
- Testing requirements
- Risk assessment
- Implementation phases

**Audience**: Backend developers, tech leads, architects

**Read this first** if you're implementing the feature.

---

### 2. Quick Reference (`quick-reference.md`)
**Purpose**: Fast lookup guide for implementation details

**Contents**:
- TL;DR summary
- Key files to modify
- Core implementation logic (code snippets)
- API endpoint specification
- Testing checklist
- Security & permissions
- Existing patterns reference

**Audience**: Developers during implementation

**Use this** for quick reference while coding.

---

## Current Status

### ✅ Completed
- [x] Codebase exploration and analysis
- [x] Technical assessment document
- [x] Quick reference guide
- [x] Implementation estimate
- [x] Risk assessment
- [x] Testing requirements defined

### ⏳ Pending (Requires Business Input)
- [ ] Design document with business rules
- [ ] Stakeholder approval for:
  - Cancellation time limits (if any)
  - Which sale statuses can be cancelled
  - Approval workflow requirements
  - Discount usage reversal rules
  - Payment/refund integration
- [ ] API contract review
- [ ] Permission matrix approval

### 🚧 Not Started
- [ ] Database migration script
- [ ] Model updates
- [ ] Service layer implementation
- [ ] API handler and routes
- [ ] Unit tests
- [ ] Integration tests
- [ ] API documentation
- [ ] Runbook creation

---

## Key Decisions Needed

Before implementation can begin, the following business questions must be answered:

### 1. Cancellation Rules
**Question**: Which sale statuses can be cancelled?
**Options**:
- A) Only "pending" sales (simplest, recommended)
- B) "pending" and "completed" sales (requires refund workflow)
- C) All sales except "refunded" (complex)

**Current Recommendation**: Option A (pending only)

### 2. Time Limits
**Question**: Is there a time limit for cancellation?
**Options**:
- A) No time limit (simplest)
- B) Within X hours of sale creation
- C) Same business day only

**Current Recommendation**: Option A (no limit for pending sales)

### 3. Approval Workflow
**Question**: Does cancellation require approval?
**Options**:
- A) Direct cancellation with permission check (simplest, recommended)
- B) Manager approval required for cancellations > X amount
- C) Two-step approval workflow

**Current Recommendation**: Option A (direct with permission)

### 4. Partial Cancellation
**Question**: Support cancelling individual items or full order only?
**Options**:
- A) Full order cancellation only (Phase 1, recommended)
- B) Item-level cancellation (Phase 2)

**Current Recommendation**: Option A (full cancellation)

### 5. Discount Handling
**Question**: What happens to discount usage on cancellation?
**Options**:
- A) Ignore discount usage reversal (Phase 1, simplest)
- B) Release discount usage count (Phase 2)

**Current Recommendation**: Option A initially, B for Phase 2

---

## Implementation Phases

### Phase 1: Foundation (Days 1-2)
**Goal**: Database and model layer ready

**Tasks**:
- Write and test database migration script
- Update `Sale` model with cancellation fields
- Update request/response models
- Add repository methods
- Write model unit tests

**Deliverables**:
- Migration script executed successfully
- Models updated and tested
- No breaking changes

---

### Phase 2: Core Logic (Days 2-3)
**Goal**: Business logic implemented and tested

**Tasks**:
- Implement `CancelSale()` service method
- Add status transition validation
- Implement inventory reversal logic
- Update transaction types
- Write service unit tests
- Write integration tests

**Deliverables**:
- Service method working end-to-end
- All tests passing
- Code coverage > 80%

---

### Phase 3: API Layer (Days 3-4)
**Goal**: API endpoint ready for consumption

**Tasks**:
- Implement handler
- Register routes
- Add permission checks
- Update API documentation (Swagger/OpenAPI)
- Write handler tests

**Deliverables**:
- API endpoint functional
- Documentation updated
- Permission checks working

---

### Phase 4: Polish & Deploy (Days 4-5)
**Goal**: Production-ready feature

**Tasks**:
- Performance testing
- Security review
- Create runbook for operations
- Feature flag setup
- Staging deployment
- Monitoring and alerts setup

**Deliverables**:
- Feature deployed to staging
- Runbook completed
- Metrics configured
- Ready for production rollout

---

## Testing Strategy

### Unit Tests (Target: > 80% coverage)
- Model validation tests
- Repository method tests (using mocks)
- Service method tests (happy path + edge cases)
- Handler tests (request validation + error handling)

### Integration Tests
- End-to-end: Create sale → Cancel → Verify inventory
- Transaction rollback scenarios
- FEFO reversal correctness
- Concurrent cancellation attempts

### Manual Testing
- Postman/curl API tests
- Multiple items across batches
- Permission denied scenarios
- Status transition validation
- Audit trail verification

### Performance Tests
- Benchmark single item cancellation
- Benchmark multi-item cancellation
- Concurrent cancellation load test

---

## Security Considerations

### Authorization
- New permission: `sales:cancel`
- Organization-scoped access (multi-tenant isolation)
- User must belong to warehouse's organization

### Audit Trail
- All cancellations logged with structured logging
- Inventory transactions maintain full history
- User ID and reason captured in database

### Input Validation
- Reason: Required, 10-500 characters
- Sale ID: Valid UUID format
- Status: Must be "pending"

---

## Rollout Plan

### Pre-Production
1. Code review by senior engineer
2. Security review by security team
3. API contract review with frontend team
4. Load testing on staging environment

### Production Rollout
1. Deploy with feature flag disabled
2. Enable for 10% of organizations (canary)
3. Monitor metrics for 24 hours
4. Expand to 50% if no issues
5. Full rollout after 1 week

### Rollback Plan
- Feature flag can disable instantly
- Database migration is backward compatible
- Manual inventory correction procedure documented

---

## Metrics & Monitoring

### Success Metrics
- Cancellation API success rate > 99%
- P95 latency < 200ms
- Zero inventory discrepancies
- Zero transaction rollback failures

### Business Metrics
- Cancellation rate by warehouse
- Most common cancellation reasons
- Average time to cancellation
- Inventory turn impact

### Alerts
- High cancellation rate (> 10% in 1 hour)
- Transaction failures (> 5 in 5 minutes)
- Inventory mismatch detected

---

## Dependencies

### Internal
- AAA Service: User authentication & permissions
- Inventory Service: Batch management
- Sales Service: Order management

### External
- PostgreSQL: Database with GORM ORM
- JWT: Token validation
- Zap: Structured logging

### Optional (Future)
- Payment Gateway: Refund processing
- Notification Service: Customer alerts
- Webhook System: External integrations

---

## Open Questions & Blockers

### Questions for Product Team
1. What are the acceptable cancellation reasons? (free text or predefined list?)
2. Should customers be notified of cancellations? (email/SMS)
3. Do we need analytics on cancellation patterns?
4. Integration with accounting system for journal entries?

### Questions for Operations Team
1. Manual cancellation override process?
2. Bulk cancellation support needed?
3. Cancellation approval SLA?
4. Inventory reconciliation frequency?

### Technical Blockers
- None identified (all dependencies available)

---

## Related Features

### Implemented
- Sale creation with FEFO allocation
- Purchase order cancellation
- Inventory transaction logging
- Multi-tenant authorization

### Future Enhancements
- Refund workflow (for completed sales)
- Partial cancellation (item-level)
- Discount usage reversal
- Customer notification integration
- Bulk cancellation API
- Cancellation analytics dashboard

---

## Contact & Ownership

**Technical Owner**: Backend Engineering Team
**Business Owner**: TBD
**Stakeholders**:
- Operations Team
- Finance Team
- Customer Support Team

**Slack Channels**:
- #backend-development
- #product-erp
- #operations

**Related PRs**: TBD (will be linked when implementation starts)

---

## Additional Resources

### Internal Documentation
- API Contracts: `.kiro/specs/api-contracts/`
- Architecture Overview: `docs/README.md`
- Permission Matrix: `README.md` (main project)

### External References
- GORM Transactions: https://gorm.io/docs/transactions.html
- Go Best Practices: https://go.dev/doc/effective_go
- PostgreSQL ACID: https://www.postgresql.org/docs/current/tutorial-transactions.html

---

**Last Updated**: 2025-11-25
**Version**: 1.0
**Status**: Assessment Complete
