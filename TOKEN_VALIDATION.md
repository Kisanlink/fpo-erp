# AAA Token Validation Tool

## Overview
The `validate_token.go` tool validates JWT tokens from the AAA service to ensure the updated claims structure works correctly.

## Usage

### Command Line
```bash
# Validate a token directly
go run validate_token.go "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Interactive mode (prompts for token)
go run validate_token.go
```

### Environment Setup
The tool uses the JWT secret from environment variable:
```bash
export AAA_JWT_SECRET="your_aaa_jwt_secret_here"
```

Or it will use the default from `.env` file.

## What It Validates

### ✅ Token Structure
- JWT format validation (header.payload.signature)
- Signature verification using AAA_JWT_SECRET
- Expiration check

### ✅ Updated Claims Fields
- `user_id` → User identifier
- `username` → User name
- `isvalidate` → Validation status (updated field name)
- `roleIds` → Array of role objects (updated field name)

### ✅ Standard JWT Claims
- `sub` → Subject (user ID)
- `iss` → Issuer (aaa-service)
- `aud` → Audience (aaa-frontend)
- `exp` → Expiration time
- `iat` → Issued at time
- `nbf` → Not before time

### ✅ Role Information
- Role ID and user associations
- Role activation status
- Creation timestamps
- Nested role definitions

### ✅ Permission Extraction
- Extracts permissions from nested `role.Role.Permissions`
- Handles both `[]string` and `[]interface{}` permission formats
- Shows total permission count

### ✅ User Profile Data
- User profile information (if available)
- Address details
- Contact information

## Example Output

```
=== AAA Token Validation Tool ===

🔍 Validating token (length: 1234 characters)...

✅ Token is VALID!

=== TOKEN DETAILS ===
User ID:      USER00000003
Username:     erpone
IsValidated:  true
Subject:      USER00000003
Issuer:       aaa-service
Audience:     [aaa-frontend]
Issued At:    2025-09-21 15:49:13 UTC
Expires At:   2025-09-22 15:50:13 UTC
✅ Token expires in: 23h59m

=== ROLES ===
Total Roles:  1

Role 1:
  ID:         USERROLE1758288753937644000
  User ID:    USER00000003
  Role ID:    ROLE000005
  Is Active:  true
  Created:    2025-09-21 21:20:13

=== PERMISSIONS ===
Total Permissions: 15
 1. products:read
 2. products:write
 3. warehouses:read
 4. warehouses:write
 5. sales:read
 6. sales:write
 ...

=== VALIDATION STATUS ===
✅ All fields successfully parsed with updated structure
✅ Token format matches AAA service specification
✅ Claims extraction working correctly
```

## Error Handling

### Invalid Token
```
❌ Token parsing failed: token is malformed
```

### Expired Token
```
⚠️ Token is EXPIRED!
```

### Wrong Secret
```
❌ Token parsing failed: signature is invalid
```

## Testing Your AAA Integration

1. **Get a token from your AAA service**
2. **Run the validation tool:**
   ```bash
   go run validate_token.go "your_token_here"
   ```
3. **Check the output:**
   - ✅ Green checkmarks = Everything working
   - ❌ Red X marks = Issues found
   - ⚠️ Yellow warnings = Token expired/other issues

## Troubleshooting

### Token Parsing Fails
- Check if `AAA_JWT_SECRET` matches your AAA service
- Verify token hasn't been corrupted during copy/paste
- Ensure token is complete (no truncation)

### Claims Missing
- Verify AAA service is sending all required fields
- Check if field names match exactly (`isvalidate`, `roleIds`)
- Confirm role structure includes nested permissions

### Permissions Not Found
- Check if `role.Role.Permissions` contains data
- Verify permission format (array of strings)
- Ensure roles are active and properly configured

## Integration Test
Use this tool to verify the ERP system can correctly parse real AAA tokens before deploying to production.