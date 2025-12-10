package models

import "time"

// ========================================
// Common Nested DTOs
// ========================================

// WebhookAddress represents an address in webhook payloads
type WebhookAddress struct {
	AddressID    string  `json:"address_id" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	AddressLine1 string  `json:"address_line_1" binding:"required"`
	AddressLine2 *string `json:"address_line_2"`
	City         string  `json:"city" binding:"required"`
	State        string  `json:"state" binding:"required"`
	PostalCode   string  `json:"postal_code" binding:"required"`
	Country      string  `json:"country" binding:"required"`
}

// WebhookCollaborator represents a collaborator (supplier) in webhook payloads
type WebhookCollaborator struct {
	ExternalID    string  `json:"external_id" binding:"required"`
	CompanyName   string  `json:"company_name"`
	ContactPerson string  `json:"contact_person"`
	ContactNumber string  `json:"contact_number"`
	Email         *string `json:"email"`
	GSTNumber     string  `json:"gst_number"`
	PANNumber     *string `json:"pan_number"`
	BankAccountNo *string `json:"bank_account_no"`
	BankIFSC      *string `json:"bank_ifsc"`
	BankName      *string `json:"bank_name"`
	Experience    *string `json:"experience"`
	AddressID     *string `json:"address_id"` // Reference to AAA address (AAA is source of truth)
}

// WebhookProduct represents a product in webhook payloads
type WebhookProduct struct {
	ExternalID  string  `json:"external_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Category    string  `json:"category" binding:"required"`
	Subcategory *string `json:"subcategory"`
	HSNCode     *string `json:"hsn_code"`
	Unit        string  `json:"unit" binding:"required"`
}

// WebhookVariant represents a product variant in webhook payloads
type WebhookVariant struct {
	ExternalID         string    `json:"external_id" binding:"required"`
	SKU                string    `json:"sku" binding:"required"`
	Name               string    `json:"name" binding:"required"`
	QuantityText       string    `json:"quantity_text" binding:"required"`
	PackSize           string    `json:"pack_size" binding:"required"`
	BrandName          *string   `json:"brand_name"`
	GSTRate            *float64  `json:"gst_rate"`
	Images             *[]string `json:"images"`
	DosageInstructions *string   `json:"dosage_instructions"`
	UsageDetails       *string   `json:"usage_details"`
}

// WebhookOrderItem represents a line item in order.created webhook
type WebhookOrderItem struct {
	Product   WebhookProduct `json:"product" binding:"required"`
	Variant   WebhookVariant `json:"variant" binding:"required"`
	Quantity  int64          `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64        `json:"unit_price" binding:"required,gt=0"`
}

// WebhookFPO represents FPO information in webhook payloads
type WebhookFPO struct {
	FPOID           string         `json:"fpo_id" binding:"required"`
	DeliveryAddress WebhookAddress `json:"delivery_address" binding:"required"`
}

// WebhookOrder represents order information in webhook payloads
type WebhookOrder struct {
	ExternalOrderID      string  `json:"external_order_id" binding:"required"`
	OrderDate            *string `json:"order_date"`
	ExpectedDeliveryDate string  `json:"expected_delivery_date" binding:"required"`
	TotalAmount          float64 `json:"total_amount"`
	Currency             string  `json:"currency"`
}

// ========================================
// 1. ORDER.CREATED Webhook
// ========================================

// OrderCreatedWebhook represents the complete order.created webhook payload
type OrderCreatedWebhook struct {
	EventType    string              `json:"event_type" binding:"required"`
	EventID      string              `json:"event_id" binding:"required"`
	Timestamp    string              `json:"timestamp" binding:"required"`
	Order        WebhookOrder        `json:"order" binding:"required"`
	FPO          WebhookFPO          `json:"fpo" binding:"required"`
	Collaborator WebhookCollaborator `json:"collaborator" binding:"required"`
	Items        []WebhookOrderItem  `json:"items" binding:"required,min=1,dive"`
}

// ========================================
// 2. ORDER.CONFIRMED Webhook
// ========================================

// OrderConfirmedWebhook represents the order.confirmed webhook payload
type OrderConfirmedWebhook struct {
	EventType            string  `json:"event_type" binding:"required"`
	EventID              string  `json:"event_id" binding:"required"`
	Timestamp            string  `json:"timestamp" binding:"required"`
	ExternalOrderID      string  `json:"external_order_id" binding:"required"`
	ConfirmedDate        *string `json:"confirmed_date"`
	ExpectedDeliveryDate *string `json:"expected_delivery_date"`
}

// ========================================
// 3. ORDER.SHIPPED Webhook
// ========================================

// OrderShippedWebhook represents the order.shipped webhook payload
type OrderShippedWebhook struct {
	EventType            string  `json:"event_type" binding:"required"`
	EventID              string  `json:"event_id" binding:"required"`
	Timestamp            string  `json:"timestamp" binding:"required"`
	ExternalOrderID      string  `json:"external_order_id" binding:"required"`
	ShippedDate          *string `json:"shipped_date"`
	TrackingNumber       *string `json:"tracking_number"`
	Carrier              *string `json:"carrier"`
	ExpectedDeliveryDate *string `json:"expected_delivery_date"`
}

// ========================================
// 4. ORDER.DELIVERED Webhook (Most Complex)
// ========================================

// WebhookDeliveryItem represents a delivered item with batch details
type WebhookDeliveryItem struct {
	ExternalProductID string  `json:"external_product_id" binding:"required"`
	ExternalVariantID string  `json:"external_variant_id" binding:"required"`
	ReceivedQuantity  int64   `json:"received_quantity" binding:"required,gt=0"`
	AcceptedQuantity  int64   `json:"accepted_quantity" binding:"required,gte=0"`
	RejectedQuantity  int64   `json:"rejected_quantity" binding:"gte=0"`
	RejectionReason   *string `json:"rejection_reason"`
	BatchNumber       string  `json:"batch_number" binding:"required"`
	ExpiryDate        string  `json:"expiry_date" binding:"required"`
	ManufacturingDate *string `json:"manufacturing_date"`
	CostPrice         float64 `json:"cost_price" binding:"required,gt=0"`
}

// OrderDeliveredWebhook represents the order.delivered webhook payload
type OrderDeliveredWebhook struct {
	EventType       string                `json:"event_type" binding:"required"`
	EventID         string                `json:"event_id" binding:"required"`
	Timestamp       string                `json:"timestamp" binding:"required"`
	ExternalOrderID string                `json:"external_order_id" binding:"required"`
	DeliveryDate    *string               `json:"delivery_date"`
	InvoiceNumber   *string               `json:"invoice_number"`
	GRNNumber       *string               `json:"grn_number"`
	Items           []WebhookDeliveryItem `json:"items" binding:"required,min=1,dive"`
}

// ========================================
// 5. ORDER.PAYMENT Webhook
// ========================================

// OrderPaymentWebhook represents the order.payment webhook payload
type OrderPaymentWebhook struct {
	EventType       string  `json:"event_type" binding:"required"`
	EventID         string  `json:"event_id" binding:"required"`
	Timestamp       string  `json:"timestamp" binding:"required"`
	ExternalOrderID string  `json:"external_order_id" binding:"required"`
	PaidAmount      float64 `json:"paid_amount" binding:"required,gt=0"`
	PaymentStatus   string  `json:"payment_status" binding:"required"`
	PaymentMethod   *string `json:"payment_method"`
	PaymentDate     *string `json:"payment_date"`
	TransactionID   *string `json:"transaction_id"`
	Remarks         *string `json:"remarks"`
}

// ========================================
// Webhook Response DTOs
// ========================================

// WebhookSuccessResponse represents a successful webhook response
type WebhookSuccessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// WebhookErrorResponse represents a webhook error response
type WebhookErrorResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// ========================================
// Helper Methods
// ========================================

// ParseTimestamp parses ISO 8601 timestamp string to time.Time
func ParseTimestamp(timestampStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timestampStr)
}

// ParseDate parses YYYY-MM-DD date string to time.Time
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// NewWebhookSuccessResponse creates a success response
func NewWebhookSuccessResponse(message string) *WebhookSuccessResponse {
	return &WebhookSuccessResponse{
		Status:  "success",
		Message: message,
	}
}

// NewWebhookErrorResponse creates an error response
func NewWebhookErrorResponse(message string, errors []string) *WebhookErrorResponse {
	return &WebhookErrorResponse{
		Status:  "error",
		Message: message,
		Errors:  errors,
	}
}
