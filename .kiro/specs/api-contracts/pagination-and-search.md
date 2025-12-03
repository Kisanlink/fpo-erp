# Pagination and Search Query Parameters Specification

## Overview

This document defines standard query parameters for pagination, filtering, sorting, and search across all aggregated APIs to ensure efficient data retrieval and consistent user experience.

---

## Standard Query Parameters

### 1. Pagination Parameters

#### Offset-Based Pagination (Default)

```
GET /api/v1/products?limit=50&offset=100
```

| Parameter | Type | Required | Default | Max | Description |
|-----------|------|----------|---------|-----|-------------|
| `limit` | integer | No | 50 | 200 | Number of records per page |
| `offset` | integer | No | 0 | - | Number of records to skip |

**Response Structure**:
```json
{
  "data": [...],
  "pagination": {
    "total": 1250,
    "limit": 50,
    "offset": 100,
    "has_more": true,
    "next_offset": 150,
    "prev_offset": 50
  }
}
```

**Advantages**:
- Simple to implement
- Easy to jump to specific page
- Familiar to users

**Disadvantages**:
- Performance degrades with large offsets
- Data consistency issues with concurrent updates

#### Cursor-Based Pagination (Recommended for Large Datasets)

```
GET /api/v1/inventory/batches?limit=50&after=BTCH_1234
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 50 | Number of records per page |
| `after` | string | No | null | Cursor pointing to last record of previous page |
| `before` | string | No | null | Cursor for reverse pagination |

**Response Structure**:
```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "next_cursor": "BTCH_1284",
    "prev_cursor": "BTCH_1234",
    "total_count": null
  }
}
```

**Advantages**:
- Consistent performance regardless of page depth
- No data skipping or duplication
- Works well with real-time data

**Disadvantages**:
- Cannot jump to arbitrary page
- More complex implementation

**Implementation Example**:
```go
func (r *InventoryRepository) GetBatchesCursorPaginated(afterCursor string, limit int) ([]models.InventoryBatch, string, error) {
    query := r.db.Table("inventory_batches").
        Where("total_quantity > 0").
        Order("id ASC").
        Limit(limit + 1) // Fetch one extra to determine has_more

    if afterCursor != "" {
        query = query.Where("id > ?", afterCursor)
    }

    var batches []models.InventoryBatch
    if err := query.Find(&batches).Error; err != nil {
        return nil, "", err
    }

    hasMore := len(batches) > limit
    if hasMore {
        batches = batches[:limit] // Remove extra record
    }

    var nextCursor string
    if hasMore && len(batches) > 0 {
        nextCursor = batches[len(batches)-1].ID
    }

    return batches, nextCursor, nil
}
```

---

### 2. Sorting Parameters

```
GET /api/v1/products?sort_by=created_at&sort_order=desc
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sort_by` | string | No | created_at | Field to sort by |
| `sort_order` | string | No | desc | Sort direction: `asc` or `desc` |

**Supported Sort Fields by Endpoint**:

**Products**:
- `created_at` - Creation date
- `updated_at` - Last modified date
- `name` - Product name (alphabetical)
- `category` - Product category
- `price` - Base price
- `stock_quantity` - Total stock across warehouses

**Inventory Batches**:
- `expiry_date` - Expiry date (FEFO sorting)
- `quantity` - Available quantity
- `cost_price` - Cost per unit
- `created_at` - Batch creation date
- `manufacturing_date` - Manufacturing date

**Sales**:
- `sale_date` - Date of sale
- `total_amount` - Sale total value
- `payment_status` - Payment status
- `customer_name` - Customer name

**Purchase Orders**:
- `order_date` - PO creation date
- `expected_delivery_date` - Expected delivery
- `total_amount` - PO total value
- `status` - PO status

**Multi-Field Sorting** (Advanced):
```
GET /api/v1/inventory/batches?sort=expiry_date:asc,quantity:desc
```

---

### 3. Filtering Parameters

#### Basic Filters

```
GET /api/v1/products?category=Grains&is_active=true
```

**Common Filters**:
| Parameter | Type | Example | Description |
|-----------|------|---------|-------------|
| `is_active` | boolean | true | Active/inactive status |
| `category` | string | Grains | Product category |
| `warehouse_id` | string | WH_001 | Specific warehouse |
| `collaborator_id` | string | CLAB_789 | Specific collaborator/vendor |
| `created_after` | ISO date | 2024-01-01 | Created after date |
| `created_before` | ISO date | 2024-12-31 | Created before date |

#### Range Filters

```
GET /api/v1/products?min_price=50&max_price=500
GET /api/v1/inventory/batches?min_quantity=100&max_quantity=1000
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `min_price` | float | Minimum price filter |
| `max_price` | float | Maximum price filter |
| `min_quantity` | integer | Minimum quantity filter |
| `max_quantity` | integer | Maximum quantity filter |
| `min_expiry_days` | integer | Minimum days until expiry |
| `max_expiry_days` | integer | Maximum days until expiry |

#### Status Filters

```
GET /api/v1/sales?status=completed,pending
GET /api/v1/purchase-orders?status=approved
```

**Supported Status Values by Endpoint**:

**Sales**:
- `draft`, `pending`, `completed`, `cancelled`, `refunded`

**Purchase Orders**:
- `draft`, `submitted`, `approved`, `partially_received`, `fully_received`, `closed`, `cancelled`

**Inventory Batches**:
- `good`, `expiring_soon`, `expired`, `low_stock`

#### Special Filters

```
GET /api/v1/inventory/batches?in_stock_only=true&expiring_soon=true
GET /api/v1/products?has_active_price=true&has_stock=true
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `in_stock_only` | boolean | Only items with quantity > 0 |
| `expiring_soon` | boolean | Expiring within 30 days |
| `low_stock` | boolean | Below configured threshold |
| `has_active_price` | boolean | Has active pricing |
| `has_images` | boolean | Has product images |

---

### 4. Search Parameters

#### Full-Text Search

```
GET /api/v1/products?search=rice basmati
GET /api/v1/collaborators?search=punjab mills
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `search` | string | Full-text search query |
| `search_fields` | string[] | Fields to search (default: all) |

**Implementation with PostgreSQL**:
```sql
SELECT *
FROM products
WHERE
    to_tsvector('english', name || ' ' || description || ' ' || category) @@
    plainto_tsquery('english', 'rice basmati')
ORDER BY
    ts_rank(to_tsvector('english', name || ' ' || description), plainto_tsquery('english', 'rice basmati')) DESC
LIMIT 50;
```

**Required Index**:
```sql
CREATE INDEX idx_products_search ON products
USING GIN (to_tsvector('english', name || ' ' || description || ' ' || category));
```

#### Field-Specific Search

```
GET /api/v1/products?name_contains=rice&category=Grains
GET /api/v1/inventory/batches?sku=PBR-1KG&barcode_starts_with=8901
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `{field}_contains` | string | Partial match (case-insensitive) |
| `{field}_starts_with` | string | Prefix match |
| `{field}_ends_with` | string | Suffix match |
| `{field}_exact` | string | Exact match (case-sensitive) |

**Performance Considerations**:
- Use `ILIKE` with caution (can be slow)
- Prefer prefix searches (`starts_with`) which can use indexes
- Avoid suffix searches (`ends_with`) when possible
- Use full-text search for complex queries

**Optimized Query Example**:
```go
func (r *ProductRepository) SearchProducts(searchTerm string, limit, offset int) ([]models.Product, error) {
    query := r.db.Table("products")

    if searchTerm != "" {
        // Use trigram similarity for fuzzy matching
        query = query.Where(
            "name % ? OR sku % ?",
            searchTerm, searchTerm,
        ).Order("similarity(name, ?) DESC", searchTerm)
    }

    var products []models.Product
    err := query.Limit(limit).Offset(offset).Find(&products).Error
    return products, err
}
```

**Required Extension & Index**:
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
CREATE INDEX idx_products_sku_trgm ON products USING gin (sku gin_trgm_ops);
```

---

### 5. Aggregation Parameters

```
GET /api/v1/sales?group_by=date&aggregate=sum(total_amount)
GET /api/v1/inventory/batches?group_by=warehouse_id&aggregate=count
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `group_by` | string | Field to group results by |
| `aggregate` | string | Aggregation function: `count`, `sum`, `avg`, `min`, `max` |
| `having` | string | Filter on aggregated results |

**Example Response**:
```json
{
  "aggregations": [
    {
      "warehouse_id": "WH_001",
      "warehouse_name": "Main Warehouse",
      "total_quantity": 125000,
      "total_value": 6875000.00,
      "batch_count": 234
    },
    {
      "warehouse_id": "WH_002",
      "warehouse_name": "Branch Warehouse",
      "total_quantity": 87500,
      "total_value": 4593750.00,
      "batch_count": 156
    }
  ],
  "summary": {
    "grand_total_quantity": 212500,
    "grand_total_value": 11468750.00,
    "total_batches": 390
  }
}
```

---

## Complete API Examples

### Example 1: Product Catalog with Advanced Filters

```http
GET /api/v1/products/list?
    category=Grains&
    is_active=true&
    has_stock=true&
    min_price=50&
    max_price=500&
    search=rice&
    sort_by=name&
    sort_order=asc&
    limit=25&
    offset=0&
    include=variants,prices,inventory

Authorization: Bearer {token}
```

**Response**:
```json
{
  "products": [
    {
      "id": "PROD_12345",
      "name": "Premium Basmati Rice",
      "category": "Grains",
      "variants": [...],
      "prices": {...},
      "stock_summary": {...}
    }
  ],
  "pagination": {
    "total": 156,
    "limit": 25,
    "offset": 0,
    "has_more": true,
    "next_offset": 25
  },
  "filters_applied": {
    "category": "Grains",
    "is_active": true,
    "has_stock": true,
    "price_range": "50-500",
    "search_term": "rice"
  }
}
```

### Example 2: Inventory Management Dashboard

```http
GET /api/v1/inventory/batches/list?
    warehouse_id=WH_001&
    in_stock_only=true&
    expiring_soon=true&
    sort_by=expiry_date&
    sort_order=asc&
    limit=50&
    offset=0&
    include=variant,product,warehouse,prices

Authorization: Bearer {token}
```

### Example 3: Sales Report with Date Range

```http
GET /api/v1/sales?
    created_after=2024-11-01&
    created_before=2024-11-30&
    status=completed&
    min_amount=1000&
    sort_by=sale_date&
    sort_order=desc&
    limit=100&
    offset=0&
    include=customer,items,payments

Authorization: Bearer {token}
```

### Example 4: Cursor-Based Inventory Pagination

```http
GET /api/v1/inventory/batches/stream?
    limit=50&
    after=BTCH_1234&
    warehouse_id=WH_001&
    sort_by=expiry_date&
    include=variant,product

Authorization: Bearer {token}
```

---

## Performance Optimization Strategies

### 1. Query Parameter Validation

```go
type PaginationParams struct {
    Limit  int `form:"limit" binding:"min=1,max=200"`
    Offset int `form:"offset" binding:"min=0"`
}

type FilterParams struct {
    Category      string    `form:"category"`
    IsActive      *bool     `form:"is_active"`
    MinPrice      *float64  `form:"min_price" binding:"omitempty,min=0"`
    MaxPrice      *float64  `form:"max_price" binding:"omitempty,gtefield=MinPrice"`
    CreatedAfter  time.Time `form:"created_after" time_format:"2006-01-02"`
    CreatedBefore time.Time `form:"created_before" time_format:"2006-01-02"`
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
    var pagination PaginationParams
    var filters FilterParams

    if err := c.ShouldBindQuery(&pagination); err != nil {
        utils.BadRequestResponse(c, "Invalid pagination parameters")
        return
    }

    if err := c.ShouldBindQuery(&filters); err != nil {
        utils.BadRequestResponse(c, "Invalid filter parameters")
        return
    }

    // Set defaults
    if pagination.Limit == 0 {
        pagination.Limit = 50
    }

    // Call service
    products, total, err := h.productService.ListProducts(filters, pagination)
    // ...
}
```

### 2. Efficient Query Building

```go
func (r *ProductRepository) BuildFilteredQuery(filters FilterParams) *gorm.DB {
    query := r.db.Table("products p")

    // Apply filters conditionally (avoid unnecessary WHERE clauses)
    if filters.Category != "" {
        query = query.Where("p.category = ?", filters.Category)
    }

    if filters.IsActive != nil {
        query = query.Where("p.is_active = ?", *filters.IsActive)
    }

    if filters.MinPrice != nil {
        query = query.Where("p.base_price >= ?", *filters.MinPrice)
    }

    if filters.MaxPrice != nil {
        query = query.Where("p.base_price <= ?", *filters.MaxPrice)
    }

    if !filters.CreatedAfter.IsZero() {
        query = query.Where("p.created_at >= ?", filters.CreatedAfter)
    }

    if !filters.CreatedBefore.IsZero() {
        query = query.Where("p.created_at <= ?", filters.CreatedBefore)
    }

    return query
}
```

### 3. Count Optimization

For large datasets, counting can be expensive. Use approximate counts:

```go
func (r *ProductRepository) GetApproximateCount() (int64, error) {
    var count int64

    // Use PostgreSQL statistics for approximate count
    err := r.db.Raw(`
        SELECT reltuples::bigint AS approximate_count
        FROM pg_class
        WHERE relname = 'products'
    `).Scan(&count).Error

    return count, err
}

func (r *ProductRepository) GetExactCount(filters FilterParams) (int64, error) {
    var count int64
    query := r.BuildFilteredQuery(filters)
    err := query.Count(&count).Error
    return count, err
}

// Use approximate for unfiltered queries, exact for filtered
func (r *ProductRepository) GetCount(filters FilterParams) (int64, bool, error) {
    hasFilters := filters.Category != "" || filters.IsActive != nil || filters.MinPrice != nil

    if !hasFilters {
        // Use approximate count (fast)
        count, err := r.GetApproximateCount()
        return count, false, err // false = approximate
    }

    // Use exact count (slower but accurate)
    count, err := r.GetExactCount(filters)
    return count, true, err // true = exact
}
```

### 4. Pagination Performance

```go
// Bad: Using OFFSET for large offsets (slow)
query.Offset(10000).Limit(50) // Scans 10,000 rows then discards them

// Good: Using cursor-based pagination
query.Where("id > ?", lastSeenID).Limit(50) // Uses index, always fast

// Good: Using keyset pagination for sorted results
query.Where("(created_at, id) > (?, ?)", lastSeenDate, lastSeenID).
    Order("created_at ASC, id ASC").
    Limit(50)
```

---

## Database Indexing for Query Parameters

See [Database Indexing Strategy](./database-indexing-strategy.md) for comprehensive index definitions.

**Quick Reference - Essential Indexes**:

```sql
-- Pagination (default sort)
CREATE INDEX idx_products_created_at_id ON products(created_at DESC, id DESC);

-- Filtering
CREATE INDEX idx_products_category_active ON products(category, is_active) WHERE is_active = true;
CREATE INDEX idx_products_price_range ON products(base_price) WHERE base_price > 0;

-- Search
CREATE INDEX idx_products_search_gin ON products USING GIN(to_tsvector('english', name || ' ' || description));
CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);

-- Date range queries
CREATE INDEX idx_products_created_at_brin ON products USING BRIN(created_at);
```

---

## Frontend Integration

### React Hook Example

```typescript
interface PaginationState {
  limit: number;
  offset: number;
  total: number;
  hasMore: boolean;
}

interface FilterState {
  category?: string;
  isActive?: boolean;
  minPrice?: number;
  maxPrice?: number;
  search?: string;
}

function useProductList() {
  const [products, setProducts] = useState<Product[]>([]);
  const [pagination, setPagination] = useState<PaginationState>({
    limit: 50,
    offset: 0,
    total: 0,
    hasMore: false,
  });
  const [filters, setFilters] = useState<FilterState>({});
  const [loading, setLoading] = useState(false);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: pagination.limit.toString(),
        offset: pagination.offset.toString(),
        ...Object.fromEntries(
          Object.entries(filters).filter(([_, v]) => v != null)
        ),
      });

      const response = await fetch(`/api/v1/products?${params}`);
      const data = await response.json();

      setProducts(data.products);
      setPagination(data.pagination);
    } catch (error) {
      console.error('Failed to fetch products', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [pagination.offset, filters]);

  const nextPage = () => {
    if (pagination.hasMore) {
      setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }));
    }
  };

  const prevPage = () => {
    if (pagination.offset > 0) {
      setPagination(prev => ({
        ...prev,
        offset: Math.max(0, prev.offset - prev.limit),
      }));
    }
  };

  const updateFilters = (newFilters: Partial<FilterState>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
    setPagination(prev => ({ ...prev, offset: 0 })); // Reset to first page
  };

  return {
    products,
    pagination,
    loading,
    nextPage,
    prevPage,
    updateFilters,
  };
}
```

---

## Best Practices

1. **Always Limit Results**: Never allow unlimited result sets
2. **Use Cursor Pagination for Real-Time Data**: Especially for feeds or streams
3. **Cache Count Queries**: Total counts change slowly, cache aggressively
4. **Validate All Parameters**: Prevent injection attacks and invalid queries
5. **Document All Filters**: Keep API documentation up to date
6. **Monitor Query Performance**: Track slow queries and optimize
7. **Use Appropriate Indexes**: Index all filterable and sortable columns
8. **Provide Sensible Defaults**: Don't require users to specify everything
9. **Support Bulk Operations**: For admin/batch operations
10. **Return Metadata**: Help clients understand pagination state

---

## Related Documents

- [Database Indexing Strategy](./database-indexing-strategy.md)
- [Optional Includes Pattern](./optional-includes-pattern.md)
- [Inventory List API Contract](./inventory-list-api.md)
