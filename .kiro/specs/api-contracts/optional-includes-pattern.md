# Optional Includes Pattern Specification

## Overview

The Optional Includes Pattern provides a flexible way for API consumers to specify exactly which related resources they need, reducing over-fetching and improving performance.

---

## Pattern Description

### Query Parameter Syntax

```
GET /api/v1/{resource}/{id}?include={relation1},{relation2},{relation3}
```

**Example**:
```
GET /api/v1/products/PROD_123?include=variants,prices,inventory
GET /api/v1/collaborators/CLAB_456?include=products,purchase_orders,addresses
GET /api/v1/sales/SALE_789?include=customer,items,payments,refunds
```

---

## Implementation Guidelines

### 1. Default Behavior

**Option A: Include Nothing by Default**
```
GET /api/v1/products/PROD_123
→ Returns only product core fields
```

**Option B: Include Everything by Default** (Recommended)
```
GET /api/v1/products/PROD_123
→ Returns product with all related data

GET /api/v1/products/PROD_123?include=none
→ Returns only product core fields
```

**Recommendation**: Use Option B for backward compatibility

### 2. Supported Include Values

#### Per-Endpoint Supported Includes

**Products API**:
- `variants` - All product variants
- `prices` - Pricing information
- `inventory` - Stock levels across warehouses
- `collaborators` - Supplier/vendor info
- `taxes` - Tax configuration
- `discounts` - Applicable discounts
- `images` - Product images (high-res)

**Collaborators API**:
- `products` - Products supplied by this collaborator
- `purchase_orders` - All POs with this collaborator
- `addresses` - Multiple addresses (billing, shipping)
- `contacts` - Contact persons
- `financial` - Credit limit, outstanding balance

**Sales API**:
- `customer` - Customer information
- `items` - Sale line items with product details
- `payments` - Payment transactions
- `refunds` - Refund records
- `shipments` - Shipping information
- `warehouse` - Source warehouse details

**Purchase Orders API**:
- `collaborator` - Vendor details
- `warehouse` - Destination warehouse
- `items` - PO line items
- `grns` - Goods Receipt Notes
- `inventory` - Inventory batches created
- `payments` - Payment history

### 3. Response Structure

#### Flat Structure (Simple)

```json
{
  "product": {...},
  "variants": [...],
  "prices": [...],
  "inventory": {...}
}
```

#### Nested Structure (Hierarchical)

```json
{
  "product": {
    "id": "PROD_123",
    "name": "...",
    "variants": [
      {
        "id": "VAR_001",
        "prices": [...],
        "inventory": {...}
      }
    ]
  }
}
```

**Recommendation**: Use nested structure for better semantic clarity

---

## Advanced Patterns

### 1. Nested Includes

Request specific nested relationships:

```
GET /api/v1/products/PROD_123?include=variants.prices,variants.inventory
```

**Response**:
```json
{
  "product": {...},
  "variants": [
    {
      "id": "VAR_001",
      "prices": {...},    // ← included
      "inventory": {...}  // ← included
      // other variant fields excluded
    }
  ]
}
```

### 2. Field Selection (Future Enhancement)

Combine with field selection for ultra-specific queries:

```
GET /api/v1/products/PROD_123?include=variants,prices&fields=id,name,variants.id,variants.sku,prices.retail_price
```

### 3. Include All Shorthand

```
GET /api/v1/products/PROD_123?include=all
→ Returns all supported includes

GET /api/v1/products/PROD_123?include=*
→ Alternative syntax for all includes
```

### 4. Exclude Pattern

```
GET /api/v1/products/PROD_123?exclude=images,inventory
→ Returns everything except images and inventory
```

---

## Backend Implementation

### Service Layer Pattern

```go
type IncludeOptions struct {
    Variants      bool
    Prices        bool
    Inventory     bool
    Collaborators bool
    Taxes         bool
}

func ParseIncludeOptions(includeParam string) IncludeOptions {
    if includeParam == "" || includeParam == "all" {
        return IncludeOptions{
            Variants:      true,
            Prices:        true,
            Inventory:     true,
            Collaborators: true,
            Taxes:         true,
        }
    }

    includes := strings.Split(includeParam, ",")
    options := IncludeOptions{}

    for _, include := range includes {
        switch strings.TrimSpace(include) {
        case "variants":
            options.Variants = true
        case "prices":
            options.Prices = true
        case "inventory":
            options.Inventory = true
        case "collaborators":
            options.Collaborators = true
        case "taxes":
            options.Taxes = true
        }
    }

    return options
}

func (s *ProductService) GetProductDetail(productID string, includes IncludeOptions) (*ProductDetailResponse, error) {
    // Fetch product
    product, err := s.productRepo.GetByID(productID)
    if err != nil {
        return nil, err
    }

    response := &ProductDetailResponse{
        Product: product,
    }

    // Conditionally fetch related data
    if includes.Variants {
        variants, err := s.variantRepo.GetByProductID(productID)
        if err != nil {
            return nil, err
        }
        response.Variants = variants
    }

    if includes.Prices {
        prices, err := s.priceRepo.GetByProductID(productID)
        if err != nil {
            return nil, err
        }
        response.Prices = prices
    }

    // ... more includes

    return response, nil
}
```

### Handler Layer Pattern

```go
func (h *ProductHandler) GetProductDetail(c *gin.Context) {
    productID := c.Param("id")
    includeParam := c.Query("include")

    // Parse includes
    includes := ParseIncludeOptions(includeParam)

    // Call service
    response, err := h.productService.GetProductDetail(productID, includes)
    if err != nil {
        utils.ErrorResponse(c, err)
        return
    }

    utils.SuccessResponse(c, http.StatusOK, response)
}
```

---

## Performance Optimizations

### 1. Parallel Data Fetching

Fetch independent includes in parallel:

```go
func (s *ProductService) GetProductDetail(productID string, includes IncludeOptions) (*ProductDetailResponse, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    response := &ProductDetailResponse{}
    errors := []error{}

    // Fetch product (required)
    product, err := s.productRepo.GetByID(productID)
    if err != nil {
        return nil, err
    }
    response.Product = product

    // Parallel includes
    if includes.Variants {
        wg.Add(1)
        go func() {
            defer wg.Done()
            variants, err := s.variantRepo.GetByProductID(productID)
            mu.Lock()
            if err != nil {
                errors = append(errors, err)
            } else {
                response.Variants = variants
            }
            mu.Unlock()
        }()
    }

    if includes.Prices {
        wg.Add(1)
        go func() {
            defer wg.Done()
            prices, err := s.priceRepo.GetByProductID(productID)
            mu.Lock()
            if err != nil {
                errors = append(errors, err)
            } else {
                response.Prices = prices
            }
            mu.Unlock()
        }()
    }

    wg.Wait()

    if len(errors) > 0 {
        return nil, errors[0]
    }

    return response, nil
}
```

### 2. Batch Loading (DataLoader Pattern)

For list endpoints, use batch loading to avoid N+1:

```go
// Instead of:
for _, product := range products {
    prices, _ := priceRepo.GetByProductID(product.ID) // N queries
}

// Use:
productIDs := extractIDs(products)
pricesMap := priceRepo.GetByProductIDs(productIDs) // 1 query
for _, product := range products {
    product.Prices = pricesMap[product.ID]
}
```

### 3. Smart Caching

Cache expensive includes separately:

```go
cacheKey := fmt.Sprintf("product:%s:variants", productID)
variants, found := cache.Get(cacheKey)
if !found {
    variants = variantRepo.GetByProductID(productID)
    cache.Set(cacheKey, variants, 5*time.Minute)
}
```

---

## Error Handling

### Invalid Include Value

**Request**:
```
GET /api/v1/products/PROD_123?include=invalid,variants
```

**Response**:
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_INCLUDE",
    "message": "Invalid include parameter: 'invalid'",
    "details": {
      "invalid_includes": ["invalid"],
      "supported_includes": ["variants", "prices", "inventory", "collaborators", "taxes"]
    }
  }
}
```

### Partial Failure Handling

**Option A: Fail Fast**
```
If any include fails → Return 500 error
```

**Option B: Graceful Degradation (Recommended)**
```json
{
  "product": {...},
  "variants": [...],
  "prices": null,
  "errors": [
    {
      "include": "prices",
      "error": "Failed to fetch prices: service unavailable"
    }
  ]
}
```

---

## Documentation Standards

### OpenAPI/Swagger Specification

```yaml
/api/v1/products/{id}:
  get:
    summary: Get product details
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
      - name: include
        in: query
        required: false
        description: |
          Comma-separated list of related resources to include.
          Supported values: variants, prices, inventory, collaborators, taxes
          Default: all
        schema:
          type: string
          example: "variants,prices,inventory"
    responses:
      '200':
        description: Product details with optional includes
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ProductDetailResponse'
```

### API Documentation Example

```markdown
## Query Parameter: include

**Type**: String (comma-separated)
**Required**: No
**Default**: all

**Description**: Specifies which related resources to include in the response.

**Supported Values**:
- `variants` - Product variants
- `prices` - Pricing information
- `inventory` - Stock availability
- `collaborators` - Supplier details
- `taxes` - Tax configuration

**Examples**:
- `?include=variants` - Only include variants
- `?include=variants,prices` - Include variants and prices
- `?include=all` - Include all related resources (default)
- `?include=none` - Exclude all related resources
```

---

## Frontend Best Practices

### 1. Request Only What You Need

```typescript
// Product listing page - need minimal data
const products = await fetch('/api/v1/products?include=none');

// Product detail page - need everything
const product = await fetch('/api/v1/products/PROD_123?include=all');

// Checkout page - need specific data
const product = await fetch('/api/v1/products/PROD_123?include=prices,inventory');
```

### 2. TypeScript Type Safety

```typescript
interface ProductDetailOptions {
  include?: ('variants' | 'prices' | 'inventory' | 'collaborators' | 'taxes' | 'all' | 'none')[];
}

function fetchProduct(id: string, options: ProductDetailOptions = {}) {
  const includeParam = options.include?.join(',') || 'all';
  return fetch(`/api/v1/products/${id}?include=${includeParam}`);
}

// Usage
fetchProduct('PROD_123', { include: ['variants', 'prices'] });
```

### 3. React Hook Example

```typescript
function useProduct(id: string, includes: string[] = ['all']) {
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const includeParam = includes.join(',');
    fetch(`/api/v1/products/${id}?include=${includeParam}`)
      .then(res => res.json())
      .then(data => {
        setProduct(data);
        setLoading(false);
      });
  }, [id, includes.join(',')]);

  return { product, loading };
}

// Usage in component
const ProductPage = ({ productId }) => {
  const { product, loading } = useProduct(productId, ['variants', 'prices', 'inventory']);

  if (loading) return <Spinner />;
  return <ProductDetail product={product} />;
};
```

---

## Migration Path

### Phase 1: Add Support
- Implement `include` parameter handling
- Keep default behavior (all includes)
- Deploy with backward compatibility

### Phase 2: Educate
- Update API documentation
- Provide frontend examples
- Conduct developer training

### Phase 3: Optimize
- Monitor include parameter usage
- Identify unused includes
- Optimize based on real-world patterns

### Phase 4: Enforce (Optional)
- Consider making `include` required
- Deprecate "all by default" behavior
- Force explicit include specifications

---

## Monitoring & Analytics

### Metrics to Track

1. **Include Parameter Usage**:
   - Which includes are most popular?
   - Which includes are rarely used?
   - Correlation between includes and response time

2. **Performance Impact**:
   - Response time by include combinations
   - Cache hit rates per include
   - Database query count per include

3. **Error Rates**:
   - Invalid include values
   - Partial failure rates
   - Timeout rates by include combination

---

## Benefits Summary

1. **Reduced Over-fetching**: Clients request only what they need
2. **Improved Performance**: Fewer database queries for unused data
3. **Better Caching**: Granular cache invalidation
4. **Flexibility**: Same endpoint serves multiple use cases
5. **Backward Compatible**: Default behavior maintains compatibility
6. **Developer Friendly**: Clear, intuitive API design

---

## Related Patterns

- **GraphQL**: More flexible but higher complexity
- **JSON:API**: Standardized include/relationship handling
- **OData**: Microsoft's query language standard
- **REST Partial Response**: Google's fields selector pattern

---

## References

- [JSON:API Specification - Inclusion of Related Resources](https://jsonapi.org/format/#fetching-includes)
- [GitHub API - Conditional Requests](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests)
- [Stripe API - Expanding Responses](https://stripe.com/docs/api/expanding_objects)
