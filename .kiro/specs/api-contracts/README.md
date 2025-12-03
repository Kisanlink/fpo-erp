# API Aggregation Contracts - Index

## Overview

This directory contains comprehensive API contracts for the new aggregated endpoints that reduce frontend API calls by 70-85% while improving performance and user experience.

---

## 📋 Table of Contents

### Core API Contracts
1. [Aggregated Product API](./aggregated-product-api.md) - Complete product details in one call
2. [Sales Context API](./sales-context-api.md) - Pre-checkout data aggregation
3. [Purchase Order Detail API](./purchase-order-detail-api.md) - Complete PO lifecycle view
4. [Inventory List API](./inventory-list-api.md) - Paginated inventory with full context

### Technical Specifications
5. [Optional Includes Pattern](./optional-includes-pattern.md) - Flexible field selection strategy
6. [Pagination and Search](./pagination-and-search.md) - Query parameters for efficient data retrieval
7. [Database Indexing Strategy](./database-indexing-strategy.md) - Comprehensive index definitions

### Implementation Guide
8. [Migration Guide](./migration-guide.md) - Phased rollout plan with timelines

---

## 🎯 Quick Start

### For Frontend Developers

**Before (Old Approach)**:
```typescript
// 4 sequential API calls
const product = await fetch('/api/v1/products/PROD_123');
const variants = await fetch('/api/v1/product-variants?product_id=PROD_123');
const prices = await fetch('/api/v1/prices?product_id=PROD_123');
const inventory = await fetch('/api/v1/inventory?product_id=PROD_123');
```

**After (New Approach)**:
```typescript
// 1 API call
const productDetail = await fetch('/api/v1/products/PROD_123/detail?include=variants,prices,inventory');
```

**See**: [Aggregated Product API](./aggregated-product-api.md)

### For Backend Developers

1. **Read** [Migration Guide](./migration-guide.md) for implementation phases
2. **Review** [Database Indexing Strategy](./database-indexing-strategy.md) for required indexes
3. **Implement** using patterns in each API contract
4. **Test** according to requirements in each contract

---

## 📊 Expected Impact

| Workflow | Current API Calls | New API Calls | Reduction | Time Saved |
|----------|------------------|---------------|-----------|------------|
| Product Detail Page | 4-5 | 1 | 75-80% | 300-600ms |
| Checkout Flow | 5-6 | 1 | 80-83% | 400-800ms |
| PO Detail View | 5 | 1 | 80% | 400-600ms |
| Inventory List (100 items) | 200+ | 1 | 99%+ | 2-4 seconds |

---

## 🏗️ Architecture Decisions

### Design Principles

1. **Backward Compatibility**: All existing endpoints remain functional
2. **Gradual Migration**: Phased rollout with feature flags
3. **Performance First**: Database-level aggregation with optimized queries
4. **Security**: Multi-level permission checks for each resource
5. **Flexibility**: Optional includes pattern for fine-grained control

### Key Patterns

- **Repository-Level Aggregation**: LATERAL JOINs for efficient queries
- **Cursor-Based Pagination**: For large datasets and real-time data
- **Consistency Tokens**: Optimistic locking for transaction integrity
- **Partial Indexes**: Filtered indexes for better performance
- **Covering Indexes**: Include frequently accessed columns

---

## 🔐 Security Considerations

- **Organization Isolation**: All queries scoped by organization_id
- **Role-Based Filtering**: Different roles see different fields
- **Permission Cascade**: Check permissions for each included resource
- **Audit Logging**: All access logged with user and timestamp
- **Rate Limiting**: Prevent abuse and ensure fair usage

---

## 🚀 Implementation Timeline

```
Phase 1: Foundation (Weeks 1-2)
  └─ Backend development & testing

Phase 2: Frontend Integration (Weeks 3-4)
  └─ Update UI to use new endpoints

Phase 3: Optimization (Weeks 5-6)
  └─ Performance tuning & monitoring

Phase 4: Deprecation (Month 3+)
  └─ Phase out old endpoints
```

**See**: [Migration Guide](./migration-guide.md) for detailed timeline

---

## 📖 Contract Structure

Each API contract includes:

- **Overview**: Problem statement and solution impact
- **API Specification**: Complete endpoint definition
- **Request/Response Examples**: Full JSON examples
- **Business Rules**: Domain logic and constraints
- **Performance Characteristics**: Response time targets and caching
- **Security Considerations**: Authentication, authorization, rate limiting
- **Migration Strategy**: Backward compatibility and rollout plan
- **Testing Requirements**: Unit, integration, and load test requirements
- **Use Cases**: Real-world usage examples
- **Monitoring & Alerts**: Metrics and alert conditions

---

## 🔧 Related Documentation

### In `.kiro/specs/`
- `api-contracts/` - This directory
- `design/` - System design documents
- `tasks/` - Implementation task breakdown

### In `.kiro/steering/`
- `product.md` - Product requirements
- `tech.md` - Technical direction
- `testing.md` - Testing strategy

### In `.kiro/dev-standards/`
- Development guidelines and standards (coming soon)

---

## 🤝 Contributing

When adding new aggregated endpoints:

1. **Create API Contract** using existing contracts as template
2. **Define Business Rules** clearly with examples
3. **Specify Performance Targets** (P50, P95, P99)
4. **Document Security Requirements** thoroughly
5. **Provide Migration Path** for backward compatibility
6. **Include Test Scenarios** covering edge cases

---

## 📞 Questions & Support

- **Technical Questions**: Refer to individual contract documents
- **Implementation Help**: See [Migration Guide](./migration-guide.md)
- **Performance Issues**: Check [Database Indexing Strategy](./database-indexing-strategy.md)

---

## 📝 Change Log

### Version 1.0 (2024-11-21)
- Initial release of all core API contracts
- Comprehensive aggregation strategy
- Database indexing specifications
- Migration guide with phased rollout plan

---

## ✅ Quick Reference

### Endpoints Summary

| Endpoint | Purpose | Aggregates | Status |
|----------|---------|------------|--------|
| `GET /api/v1/products/{id}/detail` | Product detail | Variants, Prices, Inventory | ✅ Specified |
| `GET /api/v1/sales/context` | Checkout data | Inventory, Prices, Taxes | ✅ Specified |
| `GET /api/v1/purchase-orders/{id}/detail` | PO lifecycle | Collaborator, GRN, Inventory | ✅ Specified |
| `GET /api/v1/inventory/batches/list` | Inventory list | Product, Variant, Warehouse, Prices | ✅ Specified |

### Query Parameters

| Parameter | Purpose | Example | Documents |
|-----------|---------|---------|-----------|
| `include` | Select resources | `variants,prices` | All contracts |
| `limit` | Pagination | `50` | Pagination doc |
| `offset` | Pagination | `100` | Pagination doc |
| `sort_by` | Sorting | `created_at` | Pagination doc |
| `search` | Full-text search | `rice basmati` | Pagination doc |

### Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| P50 Response Time | < 100ms | Application monitoring |
| P95 Response Time | < 250ms | Application monitoring |
| P99 Response Time | < 500ms | Application monitoring |
| Error Rate | < 0.5% | Error tracking |
| Cache Hit Rate | > 70% | Redis metrics |

---

**Last Updated**: 2024-11-21
**Version**: 1.0
**Status**: Ready for Implementation
