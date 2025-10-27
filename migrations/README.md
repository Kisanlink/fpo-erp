# Database Migrations

This directory contains manual database migration scripts for schema changes that cannot be fully handled by GORM AutoMigrate.

## Migration Files

### `20250128_refactor_attachments_to_entity.sql`

**Purpose**: Converts the attachments table from sale/return-specific to a generic entity-based system.

**Changes**:
- Adds `entity_type` and `entity_id` columns
- Migrates existing data from `sale_id`/`return_id` to the new structure
- Removes old `sale_id` and `return_id` columns
- Creates composite index on `(entity_type, entity_id)`

**When to Run**:
- **For new installations**: Not needed - GORM AutoMigrate will create the correct schema
- **For existing installations**: Run this migration ONCE if you have existing attachment data and the old `sale_id`/`return_id` columns

## How to Run Migrations

### Option 1: Using psql (Recommended)

```bash
# For local development database
psql -U kisanlink_user -d kisanlink_erp -f migrations/20250128_refactor_attachments_to_entity.sql

# For specific FPO database
psql -h <rds-endpoint> -U <username> -d <database> -f migrations/20250128_refactor_attachments_to_entity.sql
```

### Option 2: Using database client

1. Connect to your PostgreSQL database
2. Open and execute the SQL file: `migrations/20250128_refactor_attachments_to_entity.sql`

## GORM AutoMigrate Behavior

The project uses GORM AutoMigrate which runs automatically on server startup. GORM will:

✅ **What GORM does**:
- Add new columns
- Create indexes
- Update column types (with limitations)
- Add constraints

❌ **What GORM doesn't do**:
- Remove old columns
- Migrate existing data
- Drop foreign key constraints
- Rename columns

## Migration Process for Attachments Refactor

### Scenario 1: Fresh Installation (No existing data)
**Action**: No manual migration needed
- Just start the server
- GORM AutoMigrate will create the correct schema

### Scenario 2: Existing Installation with Data
**Action**: Run manual migration script

1. **Backup your database first!**
   ```bash
   pg_dump -U kisanlink_user kisanlink_erp > backup_$(date +%Y%m%d).sql
   ```

2. **Run the migration script**
   ```bash
   psql -U kisanlink_user -d kisanlink_erp -f migrations/20250128_refactor_attachments_to_entity.sql
   ```

3. **Verify the migration**
   ```sql
   -- Check table structure
   \d attachments

   -- Verify data migrated correctly
   SELECT entity_type, COUNT(*) FROM attachments GROUP BY entity_type;
   ```

4. **Start the server**
   - GORM AutoMigrate will run but find everything already correct

### Scenario 3: Development with Docker Compose
**Action**: Depends on whether you want to preserve data

**If preserving data**:
- Follow Scenario 2 steps

**If starting fresh**:
- Stop containers: `docker-compose down -v` (removes volumes)
- Start containers: `docker-compose up -d`
- Start server - GORM AutoMigrate creates correct schema

## Rollback

If you need to rollback the attachment changes:

```sql
-- WARNING: This will lose entity_type/entity_id data

-- Add old columns back
ALTER TABLE attachments ADD COLUMN sale_id VARCHAR(100);
ALTER TABLE attachments ADD COLUMN return_id VARCHAR(100);

-- Optionally migrate data back (only works for sale/return types)
UPDATE attachments SET sale_id = entity_id WHERE entity_type = 'sale';
UPDATE attachments SET return_id = entity_id WHERE entity_type = 'return';

-- Remove new columns
DROP INDEX IF EXISTS idx_attachment_entity;
ALTER TABLE attachments DROP COLUMN entity_type;
ALTER TABLE attachments DROP COLUMN entity_id;
```

## Notes

- All migration scripts use PostgreSQL's `DO $$ ... END $$` blocks for idempotency
- Scripts can be run multiple times safely - they check for existence before making changes
- Always test migrations on a development/staging environment first
- The verification query at the end of each migration shows the final table structure

## Support

For issues with migrations, check:
1. PostgreSQL server logs
2. GORM AutoMigrate logs (shown during server startup)
3. Application logs in `logs/` directory
