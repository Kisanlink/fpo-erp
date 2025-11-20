# Database Indexing Strategy for API Aggregation

## Overview

This document defines comprehensive indexing strategies for optimizing query performance across all aggregated APIs, with focus on regularly accessed columns and common query patterns.

---

## Index Creation Principles

### 1. Index Only What You Query
- Index columns used in WHERE, JOIN, ORDER BY clauses
- Avoid over-indexing (maintenance cost)
- Monitor actual query patterns before adding indexes

### 2. Column Order Matters
- Most selective columns first
- Match query WHERE clause order
- Consider cardinality (unique values)

### 3. Partial Indexes for Filtered Data
- Index only active records: `WHERE is_active = true`
- Index only non-zero quantities: `WHERE quantity > 0`
- Reduces index size and improves performance

### 4. Covering Indexes
- Include frequently accessed columns in index
- Avoid table lookups (index-only scans)
- Balance between coverage and size

---

## Core Table Indexes

### Products Table

```sql
-- Primary key (auto-created)
CREATE UNIQUE INDEX products_pkey ON products(id);

-- Natural key for external systems
CREATE UNIQUE INDEX idx_products_organization_name ON products(organization_id, name);

-- Common filtering and sorting
CREATE INDEX idx_products_category_active ON products(category, is_active)
WHERE is_active = true;

-- Date range queries (BRIN for time-series data)
CREATE INDEX idx_products_created_at_brin ON products USING BRIN(created_at);

-- Full-text search
CREATE INDEX idx_products_search_gin ON products
USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Organization scoping (multi-tenancy)
CREATE INDEX idx_products_organization_id ON products(organization_id);

-- Covering index for list queries
CREATE INDEX idx_products_list_covering ON products(
    organization_id,
    is_active,
    created_at DESC,
    id
) INCLUDE (name, category, description)
WHERE is_active = true;
```

### Product Variants Table

```sql
-- Primary key
CREATE UNIQUE INDEX product_variants_pkey ON product_variants(id);

-- Foreign key to products
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

-- SKU lookup (unique per organization)
CREATE UNIQUE INDEX idx_product_variants_sku ON product_variants(sku)
WHERE sku IS NOT NULL;

-- Barcode lookup
CREATE INDEX idx_product_variants_barcode ON product_variants(barcode)
WHERE barcode IS NOT NULL;

-- External system integration
CREATE INDEX idx_product_variants_external_id ON product_variants(external_id)
WHERE external_id IS NOT NULL;

-- Product detail aggregation
CREATE INDEX idx_product_variants_product_active ON product_variants(product_id, is_active)
WHERE is_active = true;

-- Brand filtering
CREATE INDEX idx_product_variants_brand ON product_variants(brand_name)
WHERE brand_name IS NOT NULL;

-- Fuzzy search (trigram)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_product_variants_name_trgm ON product_variants USING GIN(variant_name gin_trgm_ops);
CREATE INDEX idx_product_variants_sku_trgm ON product_variants USING GIN(sku gin_trgm_ops);

-- Covering index for list queries
CREATE INDEX idx_product_variants_list_covering ON product_variants(
    product_id,
    is_active,
    created_at DESC
) INCLUDE (variant_name, sku, brand_name, pack_size, quantity);
```

### Product Prices Table

```sql
-- Primary key
CREATE UNIQUE INDEX product_prices_pkey ON product_prices(id);

-- Variant lookup with active prices
CREATE INDEX idx_product_prices_variant_active ON product_prices(
    variant_id,
    is_active,
    price_type,
    effective_from DESC
) WHERE is_active = true;

-- Date range effective prices
CREATE INDEX idx_product_prices_effective_dates ON product_prices(
    variant_id,
    effective_from,
    effective_to
) WHERE is_active = true AND effective_from <= CURRENT_DATE AND (effective_to IS NULL OR effective_to >= CURRENT_DATE);

-- Price type filtering
CREATE INDEX idx_product_prices_type ON product_prices(price_type, is_active)
WHERE is_active = true;

-- Organization scoping
CREATE INDEX idx_product_prices_organization ON product_prices(organization_id, variant_id);

-- Covering index for price lookups
CREATE INDEX idx_product_prices_lookup_covering ON product_prices(
    variant_id,
    is_active,
    price_type,
    effective_from DESC
) INCLUDE (price, currency, effective_to)
WHERE is_active = true;
```

### Inventory Batches Table

```sql
-- Primary key
CREATE UNIQUE INDEX inventory_batches_pkey ON inventory_batches(id);

-- FEFO (First Expired, First Out) sorting
CREATE INDEX idx_inventory_batches_fefo ON inventory_batches(
    variant_id,
    expiry_date ASC,
    total_quantity DESC
) WHERE total_quantity > 0;

-- Warehouse inventory
CREATE INDEX idx_inventory_batches_warehouse ON inventory_batches(
    warehouse_id,
    variant_id,
    total_quantity DESC
) WHERE total_quantity > 0;

-- Stock availability check
CREATE INDEX idx_inventory_batches_stock ON inventory_batches(
    variant_id,
    total_quantity DESC
) WHERE total_quantity > 0;

-- Expiry monitoring
CREATE INDEX idx_inventory_batches_expiry ON inventory_batches(expiry_date ASC)
WHERE total_quantity > 0 AND expiry_date IS NOT NULL;

-- Low stock alerts
CREATE INDEX idx_inventory_batches_low_stock ON inventory_batches(
    warehouse_id,
    total_quantity ASC
) WHERE total_quantity > 0 AND total_quantity < 100;

-- Batch number lookup
CREATE INDEX idx_inventory_batches_batch_number ON inventory_batches(batch_number);

-- Organization scoping
CREATE INDEX idx_inventory_batches_organization ON inventory_batches(organization_id);

-- Covering index for inventory list
CREATE INDEX idx_inventory_batches_list_covering ON inventory_batches(
    warehouse_id,
    variant_id,
    expiry_date ASC
) INCLUDE (total_quantity, cost_price, batch_number, manufacturing_date, cgst_rate, sgst_rate)
WHERE total_quantity > 0;

-- Aggregation queries (warehouse summary)
CREATE INDEX idx_inventory_batches_warehouse_agg ON inventory_batches(warehouse_id)
INCLUDE (variant_id, total_quantity, cost_price);
```

### Warehouses Table

```sql
-- Primary key
CREATE UNIQUE INDEX warehouses_pkey ON warehouses(id);

-- Organization scoping
CREATE INDEX idx_warehouses_organization_active ON warehouses(organization_id, is_active)
WHERE is_active = true;

-- Name lookup
CREATE INDEX idx_warehouses_name ON warehouses(name);

-- Location-based queries (if geospatial needed)
-- CREATE INDEX idx_warehouses_location ON warehouses USING GIST(location);
```

### Collaborators Table

```sql
-- Primary key
CREATE UNIQUE INDEX collaborators_pkey ON collaborators(id);

-- Organization scoping
CREATE INDEX idx_collaborators_organization_active ON collaborators(organization_id, is_active)
WHERE is_active = true;

-- Company name search
CREATE INDEX idx_collaborators_company_name_trgm ON collaborators USING GIN(company_name gin_trgm_ops);

-- Contact search
CREATE INDEX idx_collaborators_contact_person ON collaborators(contact_person);

-- Email lookup
CREATE INDEX idx_collaborators_email ON collaborators(email)
WHERE email IS NOT NULL;

-- Phone lookup
CREATE INDEX idx_collaborators_phone ON collaborators(phone)
WHERE phone IS NOT NULL;

-- GSTIN lookup (India-specific)
CREATE UNIQUE INDEX idx_collaborators_gstin ON collaborators(gstin)
WHERE gstin IS NOT NULL;

-- Type filtering
CREATE INDEX idx_collaborators_type ON collaborators(collaborator_type, is_active)
WHERE is_active = true;
```

### Sales Table

```sql
-- Primary key
CREATE UNIQUE INDEX sales_pkey ON sales(id);

-- Sale number lookup
CREATE UNIQUE INDEX idx_sales_sale_number ON sales(sale_number);

-- Organization scoping
CREATE INDEX idx_sales_organization ON sales(organization_id);

-- Warehouse filtering
CREATE INDEX idx_sales_warehouse ON sales(warehouse_id);

-- Date range queries
CREATE INDEX idx_sales_sale_date_brin ON sales USING BRIN(sale_date);
CREATE INDEX idx_sales_created_at ON sales(created_at DESC);

-- Status filtering
CREATE INDEX idx_sales_status_date ON sales(status, sale_date DESC)
WHERE status IN ('pending', 'completed');

-- Payment status
CREATE INDEX idx_sales_payment_status ON sales(payment_status, sale_date DESC);

-- Customer filtering (if applicable)
CREATE INDEX idx_sales_customer ON sales(customer_id, sale_date DESC)
WHERE customer_id IS NOT NULL;

-- Financial reporting
CREATE INDEX idx_sales_date_amount ON sales(
    sale_date DESC,
    total_amount DESC
);

-- Covering index for sales list
CREATE INDEX idx_sales_list_covering ON sales(
    organization_id,
    sale_date DESC,
    id
) INCLUDE (sale_number, warehouse_id, total_amount, payment_status, status);
```

### Sale Items Table

```sql
-- Primary key
CREATE UNIQUE INDEX sale_items_pkey ON sale_items(id);

-- Sale aggregation
CREATE INDEX idx_sale_items_sale_id ON sale_items(sale_id);

-- Variant analysis
CREATE INDEX idx_sale_items_variant ON sale_items(variant_id);

-- Batch tracking
CREATE INDEX idx_sale_items_batch ON sale_items(batch_id);

-- Covering index for sale detail
CREATE INDEX idx_sale_items_sale_covering ON sale_items(sale_id)
INCLUDE (variant_id, batch_id, quantity, unit_price, total_price, tax_amount);
```

### Purchase Orders Table

```sql
-- Primary key
CREATE UNIQUE INDEX purchase_orders_pkey ON purchase_orders(id);

-- PO number lookup
CREATE UNIQUE INDEX idx_purchase_orders_po_number ON purchase_orders(po_number);

-- Organization scoping
CREATE INDEX idx_purchase_orders_organization ON purchase_orders(organization_id);

-- Collaborator filtering
CREATE INDEX idx_purchase_orders_collaborator ON purchase_orders(collaborator_id, order_date DESC);

-- Warehouse filtering
CREATE INDEX idx_purchase_orders_warehouse ON purchase_orders(warehouse_id, order_date DESC);

-- Status filtering
CREATE INDEX idx_purchase_orders_status_date ON purchase_orders(status, order_date DESC);

-- Expected delivery tracking
CREATE INDEX idx_purchase_orders_expected_delivery ON purchase_orders(expected_delivery_date ASC)
WHERE status IN ('approved', 'partially_received');

-- Date range queries
CREATE INDEX idx_purchase_orders_order_date_brin ON purchase_orders USING BRIN(order_date);

-- Covering index for PO list
CREATE INDEX idx_purchase_orders_list_covering ON purchase_orders(
    organization_id,
    order_date DESC,
    id
) INCLUDE (po_number, collaborator_id, warehouse_id, status, total_amount, paid_amount);
```

### GRN (Goods Receipt Note) Table

```sql
-- Primary key
CREATE UNIQUE INDEX grns_pkey ON grns(id);

-- GRN number lookup
CREATE UNIQUE INDEX idx_grns_grn_number ON grns(grn_number);

-- Purchase order relationship
CREATE INDEX idx_grns_purchase_order ON grns(purchase_order_id, received_date DESC);

-- Status filtering
CREATE INDEX idx_grns_status ON grns(status, received_date DESC);

-- Warehouse filtering
CREATE INDEX idx_grns_warehouse ON grns(warehouse_id, received_date DESC);

-- Date range queries
CREATE INDEX idx_grns_received_date_brin ON grns USING BRIN(received_date);
```

### Taxes Table

```sql
-- Primary key
CREATE UNIQUE INDEX taxes_pkey ON taxes(id);

-- Organization scoping
CREATE INDEX idx_taxes_organization_active ON taxes(organization_id, is_active)
WHERE is_active = true;

-- Tax type filtering
CREATE INDEX idx_taxes_type_active ON taxes(tax_type, is_active)
WHERE is_active = true;

-- HSN code lookup
CREATE INDEX idx_taxes_hsn_code ON taxes(hsn_code)
WHERE hsn_code IS NOT NULL;
```

---

## Composite Index Strategy

### Multi-Column Index Guidelines

**Rule 1: Equality Before Range**
```sql
-- Good: Equality first, then range
CREATE INDEX idx_good ON sales(status, sale_date);
WHERE status = 'completed' AND sale_date >= '2024-01-01'

-- Bad: Range first limits index effectiveness
CREATE INDEX idx_bad ON sales(sale_date, status);
WHERE status = 'completed' AND sale_date >= '2024-01-01'
```

**Rule 2: Match Query Column Order**
```sql
-- Query pattern
WHERE organization_id = 'ORG_001'
  AND category = 'Grains'
  AND is_active = true
ORDER BY created_at DESC

-- Matching index
CREATE INDEX idx_products_org_cat_active_created ON products(
    organization_id,
    category,
    is_active,
    created_at DESC
);
```

**Rule 3: Consider Cardinality**
```sql
-- High cardinality first (most selective)
CREATE INDEX idx_variants_sku_product ON product_variants(sku, product_id);

-- Low cardinality last (least selective)
CREATE INDEX idx_products_active_category ON products(is_active, category)
WHERE is_active = true;
```

---

## Specialized Index Types

### 1. Partial Indexes (Filtered)

Reduce index size by indexing only relevant rows:

```sql
-- Only active products
CREATE INDEX idx_products_active_partial ON products(category, name)
WHERE is_active = true;

-- Only items in stock
CREATE INDEX idx_inventory_in_stock_partial ON inventory_batches(variant_id, expiry_date)
WHERE total_quantity > 0;

-- Only pending/completed sales
CREATE INDEX idx_sales_active_partial ON sales(sale_date DESC)
WHERE status IN ('pending', 'completed');

-- Only upcoming deliveries
CREATE INDEX idx_po_upcoming_partial ON purchase_orders(expected_delivery_date)
WHERE status = 'approved' AND expected_delivery_date >= CURRENT_DATE;
```

**Benefits**:
- Smaller index size (faster scans)
- Lower maintenance cost
- Better cache utilization

### 2. Expression Indexes (Computed)

Index computed values:

```sql
-- Lowercase search
CREATE INDEX idx_products_name_lower ON products(LOWER(name));

-- Date truncation
CREATE INDEX idx_sales_month ON sales(DATE_TRUNC('month', sale_date));

-- JSON field extraction
CREATE INDEX idx_product_variants_images_count ON product_variants((images::jsonb -> 'count'));

-- Calculated fields
CREATE INDEX idx_sales_net_amount ON sales((total_amount - discount_amount));
```

### 3. GIN Indexes (Full-Text & Arrays)

For text search and array operations:

```sql
-- Full-text search
CREATE INDEX idx_products_fts ON products
USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- JSONB columns
CREATE INDEX idx_product_variants_images_gin ON product_variants USING GIN(images);

-- Array columns
CREATE INDEX idx_taxes_custom_tax_ids_gin ON inventory_batches USING GIN(custom_tax_ids);

-- Trigram similarity
CREATE INDEX idx_collaborators_name_gin_trgm ON collaborators USING GIN(company_name gin_trgm_ops);
```

### 4. BRIN Indexes (Large Sequential Data)

For time-series and sequential data:

```sql
-- Time-series data (very efficient for large tables)
CREATE INDEX idx_sales_created_at_brin ON sales USING BRIN(created_at);
CREATE INDEX idx_inventory_batches_created_brin ON inventory_batches USING BRIN(created_at);

-- Sequential IDs
CREATE INDEX idx_sales_id_brin ON sales USING BRIN(id);

-- Benefits: 100x-1000x smaller than B-tree, fast for sequential scans
-- Use case: Tables with >1M rows, naturally ordered data
```

---

## Index Maintenance

### 1. Monitoring Index Usage

```sql
-- Check unused indexes
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%_pkey'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Check index bloat
SELECT
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    idx_scan,
    round(100.0 * idx_scan / NULLIF(seq_scan + idx_scan, 0), 2) AS idx_scan_pct
FROM pg_stat_user_indexes
JOIN pg_stat_user_tables USING (schemaname, tablename)
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### 2. Reindexing Strategy

```sql
-- Rebuild bloated indexes (low-traffic hours)
REINDEX INDEX CONCURRENTLY idx_products_category_active;

-- Rebuild all indexes for a table
REINDEX TABLE CONCURRENTLY products;

-- Update statistics
ANALYZE products;
ANALYZE product_variants;
ANALYZE inventory_batches;
```

**Schedule**:
- Daily: ANALYZE on frequently updated tables
- Weekly: Check for bloat
- Monthly: REINDEX heavily used indexes
- Quarterly: Full database VACUUM ANALYZE

### 3. Index Creation Best Practices

```sql
-- Always use CONCURRENTLY to avoid table locks
CREATE INDEX CONCURRENTLY idx_new_index ON table_name(column);

-- If creation fails, clean up invalid index
DROP INDEX CONCURRENTLY IF EXISTS idx_new_index;

-- For large tables, increase maintenance_work_mem
SET maintenance_work_mem = '2GB';
CREATE INDEX CONCURRENTLY idx_large_table ON large_table(column);
RESET maintenance_work_mem;
```

---

## Performance Monitoring

### Query Execution Plans

```sql
-- Explain query plan
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT p.*, pv.*, pp.price
FROM products p
JOIN product_variants pv ON p.id = pv.product_id
JOIN product_prices pp ON pv.id = pp.variant_id
WHERE p.organization_id = 'ORG_001'
  AND p.category = 'Grains'
  AND p.is_active = true
  AND pp.is_active = true
ORDER BY p.created_at DESC
LIMIT 50;
```

**Look for**:
- Seq Scan (should be Index Scan)
- High execution time
- High buffer usage
- Missing indexes suggestions

### Slow Query Log

```sql
-- Enable slow query logging (postgresql.conf)
log_min_duration_statement = 1000  -- Log queries > 1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_statement = 'all'

-- Analyze slow queries
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 1000  -- Queries averaging > 1s
ORDER BY mean_time DESC
LIMIT 20;
```

---

## Index Size Management

### Current Index Sizes

```sql
SELECT
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC
LIMIT 20;
```

### Recommendations

- **Total index size < 2x table size** is healthy
- **Index size > table size** may indicate over-indexing
- **Unused indexes** should be dropped
- **Duplicate indexes** should be consolidated

---

## Implementation Checklist

### Phase 1: Core Indexes (Week 1)
- [ ] Create primary and foreign key indexes
- [ ] Add organization scoping indexes
- [ ] Create basic active/status filtering indexes

### Phase 2: Query Optimization (Week 2)
- [ ] Add indexes for common WHERE clauses
- [ ] Create covering indexes for list queries
- [ ] Add ORDER BY optimization indexes

### Phase 3: Advanced Indexes (Week 3)
- [ ] Implement full-text search indexes
- [ ] Add partial indexes for filtered queries
- [ ] Create BRIN indexes for time-series data

### Phase 4: Monitoring & Tuning (Week 4)
- [ ] Set up pg_stat_statements
- [ ] Configure slow query logging
- [ ] Create monitoring dashboards
- [ ] Establish maintenance schedule

---

## Related Documents

- [Pagination and Search Specification](./pagination-and-search.md)
- [Aggregated Product API](./aggregated-product-api.md)
- [Sales Context API](./sales-context-api.md)
- [Inventory List API](./inventory-list-api.md)
