package interfaces

import (
	"context"
)

// InvoiceServiceInterface defines the contract for invoice PDF generation
type InvoiceServiceInterface interface {
	// GenerateInvoicePDF generates a PDF invoice for a sale
	// Returns the PDF bytes, content type, filename, and any error
	GenerateInvoicePDF(ctx context.Context, saleID string) ([]byte, string, string, error)

	// CheckInvoiceRequirements checks if all required settings exist for invoice generation
	// Returns (ready bool, missing []string, error)
	CheckInvoiceRequirements(ctx context.Context) (bool, []string, error)
}
