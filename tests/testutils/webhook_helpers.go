package testutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"kisanlink-erp/internal/database/models"
)

// ========================================
// HMAC Signature Generation
// ========================================

// GenerateWebhookSignature generates an HMAC-SHA256 signature for a webhook payload
// This matches the signature generation logic in WebhookSecurityService
func GenerateWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateWebhookSignatureWithPrefix generates a signature with "sha256=" prefix
func GenerateWebhookSignatureWithPrefix(payload []byte, secret string) string {
	return "sha256=" + GenerateWebhookSignature(payload, secret)
}

// ========================================
// Webhook Payload Fixtures
// ========================================

// FixtureOrderCreatedWebhook creates a test order.created webhook payload
func FixtureOrderCreatedWebhook(externalOrderID string) *models.OrderCreatedWebhook {
	eventID := fmt.Sprintf("evt_%s_%d", externalOrderID, time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)
	expectedDelivery := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	description := "Premium organic rice"
	return &models.OrderCreatedWebhook{
		EventType: "order.created",
		EventID:   eventID,
		Timestamp: timestamp,
		Order: models.WebhookOrder{
			ExternalOrderID:      externalOrderID,
			ExpectedDeliveryDate: expectedDelivery,
			TotalAmount:          10500.00,
			Currency:             "INR",
		},
		FPO: models.WebhookFPO{
			FPOID: "FPO001",
			DeliveryAddress: models.WebhookAddress{
				AddressID:    "ADDR_12345678",
				Type:         "delivery",
				AddressLine1: "123 Main Street",
				City:         "Mumbai",
				State:        "Maharashtra",
				PostalCode:   "400001",
				Country:      "India",
			},
		},
		Collaborator: models.WebhookCollaborator{
			ExternalID:    "COLLAB001",
			CompanyName:   "ABC Suppliers",
			ContactPerson: "John Doe",
			ContactNumber: "9876543210",
			GSTNumber:     "27AABCU9603R1ZX",
			BankAccountNo: ptrString("1234567890"),
			BankIFSC:      ptrString("HDFC0001234"),
		},
		Items: []models.WebhookOrderItem{
			{
				Product: models.WebhookProduct{
					ExternalID:  "PROD001",
					Name:        "Organic Rice",
					Description: &description,
					Category:    "Grains",
					Unit:        "kg",
				},
				Variant: models.WebhookVariant{
					ExternalID:   "VAR001",
					SKU:          "RICE-1KG",
					Name:         "Organic Rice 1kg",
					QuantityText: "1kg",
					PackSize:     "1kg",
				},
				Quantity:  100,
				UnitPrice: 105.00,
			},
		},
	}
}

// FixtureOrderConfirmedWebhook creates a test order.confirmed webhook payload
func FixtureOrderConfirmedWebhook(externalOrderID string) *models.OrderConfirmedWebhook {
	eventID := fmt.Sprintf("evt_%s_confirmed_%d", externalOrderID, time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)
	confirmedDate := time.Now().Format(time.RFC3339)
	expectedDelivery := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	return &models.OrderConfirmedWebhook{
		EventType:            "order.confirmed",
		EventID:              eventID,
		Timestamp:            timestamp,
		ExternalOrderID:      externalOrderID,
		ConfirmedDate:        &confirmedDate,
		ExpectedDeliveryDate: &expectedDelivery,
	}
}

// FixtureOrderShippedWebhook creates a test order.shipped webhook payload
func FixtureOrderShippedWebhook(externalOrderID string) *models.OrderShippedWebhook {
	eventID := fmt.Sprintf("evt_%s_shipped_%d", externalOrderID, time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)
	shippedDate := time.Now().Format(time.RFC3339)
	expectedDelivery := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
	trackingNumber := "TRACK123456"
	carrier := "BlueDart"

	return &models.OrderShippedWebhook{
		EventType:            "order.shipped",
		EventID:              eventID,
		Timestamp:            timestamp,
		ExternalOrderID:      externalOrderID,
		ShippedDate:          &shippedDate,
		TrackingNumber:       &trackingNumber,
		Carrier:              &carrier,
		ExpectedDeliveryDate: &expectedDelivery,
	}
}

// FixtureOrderDeliveredWebhook creates a test order.delivered webhook payload
func FixtureOrderDeliveredWebhook(externalOrderID string) *models.OrderDeliveredWebhook {
	eventID := fmt.Sprintf("evt_%s_delivered_%d", externalOrderID, time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)
	deliveryDate := time.Now().Format(time.RFC3339)
	invoiceNumber := "INV-2025-001"
	grnNumber := "GRN-2025-001"
	expiryDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02") // 1 year from now

	return &models.OrderDeliveredWebhook{
		EventType:       "order.delivered",
		EventID:         eventID,
		Timestamp:       timestamp,
		ExternalOrderID: externalOrderID,
		DeliveryDate:    &deliveryDate,
		InvoiceNumber:   &invoiceNumber,
		GRNNumber:       &grnNumber,
		Items: []models.WebhookDeliveryItem{
			{
				ExternalProductID: "PROD001",
				ExternalVariantID: "VAR001",
				ReceivedQuantity:  100,
				AcceptedQuantity:  95,
				RejectedQuantity:  5,
				BatchNumber:       "BATCH-2025-001",
				ExpiryDate:        expiryDate,
				CostPrice:         105.00,
			},
		},
	}
}

// FixtureOrderPaymentWebhook creates a test order.payment webhook payload
func FixtureOrderPaymentWebhook(externalOrderID string) *models.OrderPaymentWebhook {
	eventID := fmt.Sprintf("evt_%s_payment_%d", externalOrderID, time.Now().Unix())
	timestamp := time.Now().Format(time.RFC3339)
	paymentDate := time.Now().Format(time.RFC3339)
	paymentMethod := "bank_transfer"
	transactionID := "TXN-12345678"
	remarks := "Payment received successfully"

	return &models.OrderPaymentWebhook{
		EventType:       "order.payment",
		EventID:         eventID,
		Timestamp:       timestamp,
		ExternalOrderID: externalOrderID,
		PaidAmount:      10500.00,
		PaymentStatus:   "paid",
		PaymentMethod:   &paymentMethod,
		PaymentDate:     &paymentDate,
		TransactionID:   &transactionID,
		Remarks:         &remarks,
	}
}

// ========================================
// Webhook JSON Marshaling Helpers
// ========================================

// MarshalWebhook marshals a webhook payload to JSON bytes
func MarshalWebhook(webhook interface{}) ([]byte, error) {
	return json.Marshal(webhook)
}

// MustMarshalWebhook marshals a webhook payload to JSON bytes, panics on error
func MustMarshalWebhook(webhook interface{}) []byte {
	data, err := json.Marshal(webhook)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal webhook: %v", err))
	}
	return data
}

// ========================================
// Webhook Header Helpers
// ========================================

// WebhookTestHeaders contains all required webhook headers
type WebhookTestHeaders struct {
	Signature string
	EventID   string
	Timestamp string
}

// GenerateWebhookHeaders generates all required webhook headers with HMAC signature
func GenerateWebhookHeaders(payload []byte, secret string) *WebhookTestHeaders {
	signature := GenerateWebhookSignature(payload, secret)
	eventID := fmt.Sprintf("test_event_%d", time.Now().Unix())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return &WebhookTestHeaders{
		Signature: signature,
		EventID:   eventID,
		Timestamp: timestamp,
	}
}

// GenerateWebhookHeadersWithEventID generates webhook headers with a specific event ID
func GenerateWebhookHeadersWithEventID(payload []byte, secret, eventID string) *WebhookTestHeaders {
	signature := GenerateWebhookSignature(payload, secret)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return &WebhookTestHeaders{
		Signature: signature,
		EventID:   eventID,
		Timestamp: timestamp,
	}
}

// GenerateExpiredWebhookHeaders generates webhook headers with expired timestamp (> 5 minutes old)
func GenerateExpiredWebhookHeaders(payload []byte, secret string) *WebhookTestHeaders {
	signature := GenerateWebhookSignature(payload, secret)
	eventID := fmt.Sprintf("test_event_%d", time.Now().Unix())
	// Timestamp 6 minutes ago (expired)
	timestamp := fmt.Sprintf("%d", time.Now().Add(-6*time.Minute).Unix())

	return &WebhookTestHeaders{
		Signature: signature,
		EventID:   eventID,
		Timestamp: timestamp,
	}
}

// GenerateFutureWebhookHeaders generates webhook headers with future timestamp
func GenerateFutureWebhookHeaders(payload []byte, secret string) *WebhookTestHeaders {
	signature := GenerateWebhookSignature(payload, secret)
	eventID := fmt.Sprintf("test_event_%d", time.Now().Unix())
	// Timestamp 2 minutes in future (beyond 1 minute tolerance)
	timestamp := fmt.Sprintf("%d", time.Now().Add(2*time.Minute).Unix())

	return &WebhookTestHeaders{
		Signature: signature,
		EventID:   eventID,
		Timestamp: timestamp,
	}
}

// GenerateInvalidSignatureHeaders generates webhook headers with invalid signature
func GenerateInvalidSignatureHeaders(payload []byte) *WebhookTestHeaders {
	eventID := fmt.Sprintf("test_event_%d", time.Now().Unix())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return &WebhookTestHeaders{
		Signature: "invalid_signature_12345",
		EventID:   eventID,
		Timestamp: timestamp,
	}
}

// ========================================
// Timestamp Helpers
// ========================================

// CurrentUnixTimestamp returns current time as Unix timestamp string
func CurrentUnixTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// PastUnixTimestamp returns past time as Unix timestamp string (minutes ago)
func PastUnixTimestamp(minutesAgo int) string {
	return fmt.Sprintf("%d", time.Now().Add(-time.Duration(minutesAgo)*time.Minute).Unix())
}

// FutureUnixTimestamp returns future time as Unix timestamp string (minutes from now)
func FutureUnixTimestamp(minutesFromNow int) string {
	return fmt.Sprintf("%d", time.Now().Add(time.Duration(minutesFromNow)*time.Minute).Unix())
}

// ========================================
// Helper Functions
// ========================================

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}
