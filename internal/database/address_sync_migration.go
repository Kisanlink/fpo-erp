package database

import (
	"context"
	"log"
	"time"

	"kisanlink-erp/internal/aaa"

	"gorm.io/gorm"
)

// SyncExistingAddressesToLocal syncs address data from AAA service to local cache
// for all warehouses and collaborators that have address_id but no local address data.
// This is a one-time migration to populate the local cache for existing records.
//
// NOTE: This requires an active AAA service connection and valid JWT token.
// Run this after deploying the address cache fields to the database.
func SyncExistingAddressesToLocal(db *gorm.DB, addressClient *aaa.AddressGRPCClient, jwtToken string) error {
	log.Println("Running address sync migration to local cache...")

	if addressClient == nil {
		log.Println("WARNING: AAA address client is nil - skipping address sync")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Sync warehouses
	warehouseCount, err := syncWarehouseAddresses(ctx, db, addressClient, jwtToken)
	if err != nil {
		log.Printf("WARNING: Failed to sync warehouse addresses: %v", err)
		// Don't fail the migration - continue with collaborators
	} else {
		log.Printf("Synced %d warehouse addresses to local cache", warehouseCount)
	}

	// Sync collaborators
	collaboratorCount, err := syncCollaboratorAddresses(ctx, db, addressClient, jwtToken)
	if err != nil {
		log.Printf("WARNING: Failed to sync collaborator addresses: %v", err)
	} else {
		log.Printf("Synced %d collaborator addresses to local cache", collaboratorCount)
	}

	log.Println("Address sync migration completed")
	return nil
}

// syncWarehouseAddresses syncs warehouse addresses to local cache
func syncWarehouseAddresses(ctx context.Context, db *gorm.DB, addressClient *aaa.AddressGRPCClient, jwtToken string) (int, error) {
	if !db.Migrator().HasTable("warehouses") {
		log.Println("Warehouses table does not exist - skipping")
		return 0, nil
	}

	// Find warehouses with address_id but no local address data (state is null as indicator)
	type warehouseToSync struct {
		ID        string
		AddressID string
	}

	var warehouses []warehouseToSync
	query := `
		SELECT id, address_id
		FROM warehouses
		WHERE address_id IS NOT NULL
		AND address_id != ''
		AND state IS NULL
		AND deleted_at IS NULL
	`
	if err := db.Raw(query).Scan(&warehouses).Error; err != nil {
		return 0, err
	}

	if len(warehouses) == 0 {
		log.Println("No warehouses need address sync")
		return 0, nil
	}

	log.Printf("Found %d warehouses needing address sync", len(warehouses))

	synced := 0
	for _, w := range warehouses {
		// Fetch address from AAA
		address, err := addressClient.GetAddress(ctx, w.AddressID, jwtToken)
		if err != nil {
			log.Printf("Failed to fetch address %s for warehouse %s: %v", w.AddressID, w.ID, err)
			continue
		}

		// Update local cache fields
		updateSQL := `
			UPDATE warehouses
			SET address_type = ?,
				house = ?,
				street = ?,
				landmark = ?,
				post_office = ?,
				subdistrict = ?,
				district = ?,
				vtc = ?,
				state = ?,
				country = ?,
				pincode = ?
			WHERE id = ?
		`
		if err := db.Exec(updateSQL,
			address.Type,
			address.House,
			address.Street,
			address.Landmark,
			address.PostOffice,
			address.Subdistrict,
			address.District,
			address.VTC,
			address.State,
			address.Country,
			address.Pincode,
			w.ID,
		).Error; err != nil {
			log.Printf("Failed to update warehouse %s: %v", w.ID, err)
			continue
		}

		synced++
		log.Printf("Synced address for warehouse %s", w.ID)
	}

	return synced, nil
}

// syncCollaboratorAddresses syncs collaborator addresses to local cache
func syncCollaboratorAddresses(ctx context.Context, db *gorm.DB, addressClient *aaa.AddressGRPCClient, jwtToken string) (int, error) {
	if !db.Migrator().HasTable("collaborators") {
		log.Println("Collaborators table does not exist - skipping")
		return 0, nil
	}

	// Find collaborators with address_id but no local address data (state is null as indicator)
	type collaboratorToSync struct {
		ID        string
		AddressID string
	}

	var collaborators []collaboratorToSync
	query := `
		SELECT id, address_id
		FROM collaborators
		WHERE address_id IS NOT NULL
		AND address_id != ''
		AND state IS NULL
		AND deleted_at IS NULL
	`
	if err := db.Raw(query).Scan(&collaborators).Error; err != nil {
		return 0, err
	}

	if len(collaborators) == 0 {
		log.Println("No collaborators need address sync")
		return 0, nil
	}

	log.Printf("Found %d collaborators needing address sync", len(collaborators))

	synced := 0
	for _, c := range collaborators {
		// Fetch address from AAA
		address, err := addressClient.GetAddress(ctx, c.AddressID, jwtToken)
		if err != nil {
			log.Printf("Failed to fetch address %s for collaborator %s: %v", c.AddressID, c.ID, err)
			continue
		}

		// Update local cache fields
		updateSQL := `
			UPDATE collaborators
			SET address_type = ?,
				house = ?,
				street = ?,
				landmark = ?,
				post_office = ?,
				subdistrict = ?,
				district = ?,
				vtc = ?,
				state = ?,
				country = ?,
				pincode = ?
			WHERE id = ?
		`
		if err := db.Exec(updateSQL,
			address.Type,
			address.House,
			address.Street,
			address.Landmark,
			address.PostOffice,
			address.Subdistrict,
			address.District,
			address.VTC,
			address.State,
			address.Country,
			address.Pincode,
			c.ID,
		).Error; err != nil {
			log.Printf("Failed to update collaborator %s: %v", c.ID, err)
			continue
		}

		synced++
		log.Printf("Synced address for collaborator %s", c.ID)
	}

	return synced, nil
}
