package docs

// GenerateOpenAPISpec generates the OpenAPI 3.0 specification JSON
func GenerateOpenAPISpec() string {
	return `{
  "openapi": "3.0.0",
  "info": {
    "title": "Kisanlink ERP API",
    "version": "1.0.0",
    "description": "Comprehensive ERP system for agricultural cooperatives with multi-tenant architecture",
    "contact": {
      "name": "Kisanlink Support",
      "url": "https://github.com/Kisanlink/fpo-erp",
      "email": "info@kisanlink.in"
    },
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    }
  },
  "servers": [
    {
      "url": "http://localhost:3000/api/v1",
      "description": "Development server"
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "JWT Authorization header using Bearer scheme. Enter 'Bearer' [space] and then your token."
      }
    },
    "schemas": {
      "Response": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "example": true
          },
          "message": {
            "type": "string",
            "example": "Operation completed successfully"
          },
          "data": {
            "type": "object"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "example": "2024-01-15T10:30:00Z"
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "example": false
          },
          "message": {
            "type": "string",
            "example": "Error occurred"
          },
          "error": {
            "type": "string",
            "example": "Detailed error message"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "example": "2024-01-15T10:30:00Z"
          }
        }
      },
      "Warehouse": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "example": "WHSE_12345678"
          },
          "name": {
            "type": "string",
            "example": "Main Warehouse"
          },
          "address_id": {
            "type": "string",
            "example": "ADDR_12345678"
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          },
          "updated_at": {
            "type": "string",
            "format": "date-time"
          }
        }
      },
      "CreateWarehouseRequest": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {
            "type": "string",
            "example": "New Warehouse"
          },
          "address_id": {
            "type": "string",
            "example": "ADDR_12345678"
          },
          "address": {
            "type": "object",
            "properties": {
              "type": {
                "type": "string",
                "example": "WORK"
              },
              "address_line_1": {
                "type": "string",
                "example": "123 Main St"
              },
              "city": {
                "type": "string",
                "example": "Mumbai"
              },
              "state": {
                "type": "string",
                "example": "Maharashtra"
              },
              "postal_code": {
                "type": "string",
                "example": "400001"
              },
              "country": {
                "type": "string",
                "example": "India"
              }
            }
          }
        }
      },
      "UpdateWarehouseRequest": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "example": "Updated Warehouse Name"
          },
          "address_id": {
            "type": "string",
            "example": "ADDR_87654321"
          }
        }
      }
    }
  },
  "security": [
    {
      "BearerAuth": []
    }
  ],
  "tags": [
    {
      "name": "Warehouses",
      "description": "Warehouse management operations"
    },
    {
      "name": "Products",
      "description": "Product catalog and inventory management"
    },
    {
      "name": "Inventory",
      "description": "Inventory batch and transaction management"
    },
    {
      "name": "Sales",
      "description": "Sales transaction management"
    },
    {
      "name": "Returns",
      "description": "Return and refund processing"
    },
    {
      "name": "Taxes",
      "description": "Tax configuration and calculation"
    },
    {
      "name": "Discounts",
      "description": "Discount management and validation"
    },
    {
      "name": "Product Prices",
      "description": "Product pricing management"
    },
    {
      "name": "Attachments",
      "description": "File attachment management"
    },
    {
      "name": "Bank Payments",
      "description": "Bank payment processing"
    },
    {
      "name": "Refund Policies",
      "description": "Refund policy management"
    }
  ],
  "paths": {
    "/warehouses": {
      "get": {
        "tags": ["Warehouses"],
        "summary": "Get All Warehouses",
        "description": "Retrieve all warehouses (requires authentication)",
        "security": [{"BearerAuth": []}],
        "responses": {
          "200": {
            "description": "Warehouses retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {"$ref": "#/components/schemas/Response"},
                    {
                      "type": "object",
                      "properties": {
                        "data": {
                          "type": "array",
                          "items": {"$ref": "#/components/schemas/Warehouse"}
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "401": {
            "description": "Unauthorized",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          },
          "500": {
            "description": "Internal server error",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      },
      "post": {
        "tags": ["Warehouses"],
        "summary": "Create Warehouse",
        "description": "Create a new warehouse (requires authentication)",
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/CreateWarehouseRequest"}
            }
          }
        },
        "responses": {
          "201": {
            "description": "Warehouse created successfully",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/Response"}
              }
            }
          },
          "400": {
            "description": "Bad request",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          },
          "401": {
            "description": "Unauthorized",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      }
    },
    "/warehouses/{id}": {
      "get": {
        "tags": ["Warehouses"],
        "summary": "Get Warehouse by ID",
        "description": "Retrieve a specific warehouse by its ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            },
            "description": "Warehouse ID"
          }
        ],
        "responses": {
          "200": {
            "description": "Warehouse retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {"$ref": "#/components/schemas/Response"},
                    {
                      "type": "object",
                      "properties": {
                        "data": {"$ref": "#/components/schemas/Warehouse"}
                      }
                    }
                  ]
                }
              }
            }
          },
          "404": {
            "description": "Warehouse not found",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      },
      "patch": {
        "tags": ["Warehouses"],
        "summary": "Update Warehouse",
        "description": "Update an existing warehouse",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            },
            "description": "Warehouse ID"
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/UpdateWarehouseRequest"}
            }
          }
        },
        "responses": {
          "200": {
            "description": "Warehouse updated successfully",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/Response"}
              }
            }
          },
          "400": {
            "description": "Bad request",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          },
          "404": {
            "description": "Warehouse not found",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      },
      "delete": {
        "tags": ["Warehouses"],
        "summary": "Delete Warehouse",
        "description": "Delete a warehouse by ID",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            },
            "description": "Warehouse ID"
          }
        ],
        "responses": {
          "200": {
            "description": "Warehouse deleted successfully",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/Response"}
              }
            }
          },
          "404": {
            "description": "Warehouse not found",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      }
    },
    "/warehouses/search": {
      "get": {
        "tags": ["Warehouses"],
        "summary": "Search Warehouses",
        "description": "Search warehouses by query string",
        "parameters": [
          {
            "name": "q",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string"
            },
            "description": "Search query"
          }
        ],
        "responses": {
          "200": {
            "description": "Search results retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {"$ref": "#/components/schemas/Response"},
                    {
                      "type": "object",
                      "properties": {
                        "data": {
                          "type": "array",
                          "items": {"$ref": "#/components/schemas/Warehouse"}
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "description": "Bad request - missing query parameter",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
              }
            }
          }
        }
      }
    }
  }
}`
}
