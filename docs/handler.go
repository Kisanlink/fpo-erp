package docs

import (
	"github.com/gin-gonic/gin"
)

// SetupDocumentationRoutes configures API documentation routes
func SetupDocumentationRoutes(router *gin.Engine) {
	// Scalar API Documentation - Main Route
	router.GET("/docs", ScalarDocumentationHandler)

	// Alternative Scalar Documentation
	router.GET("/docs-alt", ScalarAlternativeHandler)

	// Simple HTML Documentation (backup)
	router.GET("/docs-simple", SimpleDocumentationHandler)

	// OpenAPI JSON specification endpoint
	router.GET("/api/openapi.json", OpenAPISpecHandler)
}

// ScalarDocumentationHandler serves the main Scalar documentation interface
func ScalarDocumentationHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!doctype html>
<html>
<head>
    <title>Kisanlink ERP API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
        body { margin: 0; padding: 0; font-family: Arial, sans-serif; }
        .loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-size: 18px;
            color: #666;
        }
        .error {
            display: none;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-family: Arial, sans-serif;
            text-align: center;
            padding: 20px;
        }
        .error h1 { color: #d32f2f; margin-bottom: 10px; }
        .error p { color: #666; margin-bottom: 20px; }
        .error a { color: #1976d2; text-decoration: none; }
        .error a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="loading" id="loading">Loading API Documentation...</div>
    <div class="error" id="error">
        <h1>Failed to Load Documentation</h1>
        <p>The API documentation could not be loaded.</p>
        <p>
            <a href="/docs-alt">Try Alternative Documentation</a> |
            <a href="/docs-simple">Simple HTML Documentation</a> |
            <a href="/api/openapi.json">View OpenAPI JSON</a>
        </p>
    </div>
    <div id="scalar-container"></div>

    <script type="module">
        async function loadScalar() {
            try {
                // Import Scalar API Reference
                const { ApiReference } = await import('https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest/dist/browser/standalone.js');

                const configuration = {
                    spec: {
                        url: '/api/openapi.json',
                    },
                    theme: 'purple',
                    layout: 'modern',
                    showSidebar: true,
                    hideDownloadButton: false,
                    darkMode: false,
                    customCss: `+"`"+`
                        .scalar-app {
                            height: 100vh;
                        }
                        body {
                            margin: 0;
                            padding: 0;
                        }
                    `+"`"+`
                };

                // Hide loading and show documentation
                document.getElementById('loading').style.display = 'none';
                const container = document.getElementById('scalar-container');

                if (container) {
                    await ApiReference(container, configuration);
                } else {
                    throw new Error('Container not found');
                }

            } catch (error) {
                console.error('Failed to load Scalar:', error);
                document.getElementById('loading').style.display = 'none';
                document.getElementById('error').style.display = 'flex';
            }
        }

        // Load Scalar when DOM is ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', loadScalar);
        } else {
            loadScalar();
        }
    </script>
</body>
</html>`)
}

// ScalarAlternativeHandler serves alternative Scalar documentation
func ScalarAlternativeHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!doctype html>
<html>
<head>
    <title>Kisanlink ERP API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <script src="https://unpkg.com/@scalar/api-reference@latest/dist/browser/standalone.min.js"></script>
</head>
<body>
    <div id="openapi-ui"></div>
    <script>
        window.addEventListener('DOMContentLoaded', function() {
            try {
                if (window.ScalarApiReference && window.ScalarApiReference.ApiReference) {
                    const { ApiReference } = window.ScalarApiReference;
                    ApiReference('#openapi-ui', {
                        spec: { url: '/api/openapi.json' },
                        theme: 'purple',
                        layout: 'modern'
                    });
                } else {
                    throw new Error('Scalar not loaded');
                }
            } catch (e) {
                console.error(e);
                document.getElementById('openapi-ui').innerHTML = `+"`"+`
                    <div style="text-align: center; padding: 50px; font-family: Arial, sans-serif;">
                        <h1 style="color: #d32f2f;">Documentation Load Error</h1>
                        <p style="color: #666;">Failed to load Scalar documentation.</p>
                        <p>
                            <a href="/docs-simple" style="color: #1976d2;">Try Simple Documentation</a> |
                            <a href="/api/openapi.json" style="color: #1976d2;">View OpenAPI JSON</a>
                        </p>
                    </div>
                `+"`"+`;
            }
        });
    </script>
</body>
</html>`)
}

// SimpleDocumentationHandler serves simple HTML documentation as fallback
func SimpleDocumentationHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!doctype html>
<html>
<head>
    <title>Kisanlink ERP API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; line-height: 1.6; }
        .endpoint { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; background: #f9f9f9; }
        .method { display: inline-block; padding: 5px 10px; border-radius: 3px; color: white; font-weight: bold; margin-right: 10px; }
        .get { background-color: #61affe; }
        .post { background-color: #49cc90; }
        .put { background-color: #fca130; }
        .patch { background-color: #50e3c2; }
        .delete { background-color: #f93e3e; }
        .path { font-family: monospace; font-size: 14px; background: #f0f0f0; padding: 2px 4px; border-radius: 3px; }
        .auth { background: #fff3cd; border: 1px solid #ffeaa7; padding: 10px; border-radius: 5px; margin: 10px 0; }
        .section { margin: 30px 0; }
        h1 { color: #333; border-bottom: 2px solid #6c5ce7; padding-bottom: 10px; }
        h2 { color: #444; border-bottom: 1px solid #ddd; padding-bottom: 5px; }
        h3 { color: #555; }
    </style>
</head>
<body>
    <h1>🌾 Kisanlink ERP API Documentation</h1>
    <div class="auth">
        <strong>🔐 Authentication:</strong> Bearer Token (JWT) required for most endpoints<br>
        <strong>🌐 Base URL:</strong> <code>http://localhost:8080/api/v1</code><br>
        <strong>📄 Full OpenAPI Spec:</strong> <a href="/api/openapi.json">JSON Format</a> |
        <a href="/docs">Scalar UI</a> | <a href="/docs-alt">Alternative UI</a>
    </div>

    <div class="section">
        <h2>📦 Warehouse Management</h2>
        <div class="endpoint">
            <span class="method get">GET</span> <span class="path">/warehouses</span>
            <p>Get all warehouses (requires authentication)</p>
        </div>
        <div class="endpoint">
            <span class="method post">POST</span> <span class="path">/warehouses</span>
            <p>Create a new warehouse</p>
        </div>
        <div class="endpoint">
            <span class="method get">GET</span> <span class="path">/warehouses/{id}</span>
            <p>Get a specific warehouse by ID</p>
        </div>
        <div class="endpoint">
            <span class="method patch">PATCH</span> <span class="path">/warehouses/{id}</span>
            <p>Update an existing warehouse</p>
        </div>
        <div class="endpoint">
            <span class="method delete">DELETE</span> <span class="path">/warehouses/{id}</span>
            <p>Delete a warehouse</p>
        </div>
        <div class="endpoint">
            <span class="method get">GET</span> <span class="path">/warehouses/search</span>
            <p>Search warehouses by query string</p>
        </div>
    </div>

    <div class="section">
        <h2>📋 All API Categories</h2>
        <h3>Products & Inventory</h3>
        <ul>
            <li><strong>Products:</strong> /products, /products/{id}, /products/sku/{sku}, /products/{id}/prices</li>
            <li><strong>Inventory:</strong> /batches, /batches/{id}/transactions</li>
        </ul>

        <h3>Sales & Returns</h3>
        <ul>
            <li><strong>Sales:</strong> /sales, /sales/{id}, /sales/customer/{customerID}</li>
            <li><strong>Returns:</strong> /returns, /returns/{id}, /returns/customer/{customerID}</li>
        </ul>

        <h3>Financial</h3>
        <ul>
            <li><strong>Taxes:</strong> /taxes, /taxes/calculate, /taxes/applications/sale/{saleID}</li>
            <li><strong>Discounts:</strong> /discounts, /discounts/validate, /discounts/active</li>
            <li><strong>Prices:</strong> /prices, /prices/expired</li>
            <li><strong>Bank Payments:</strong> /bank-payments, /bank-payments/sale/{saleID}</li>
            <li><strong>Refund Policies:</strong> /refund-policies, /refund-policies/{id}</li>
        </ul>

        <h3>Attachments</h3>
        <ul>
            <li><strong>File Management:</strong> /attachments, /attachments/{id}/download</li>
        </ul>
    </div>

    <div class="section">
        <p><strong>💡 Note:</strong> This is a simplified view. For complete documentation with request/response schemas, authentication details, and interactive testing, use the <a href="/api/openapi.json">OpenAPI JSON specification</a> or the <a href="/docs">Scalar documentation interface</a>.</p>
    </div>
</body>
</html>`)
}

// OpenAPISpecHandler generates and serves the OpenAPI 3.0 specification JSON
func OpenAPISpecHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(200, GenerateOpenAPISpec())
}
