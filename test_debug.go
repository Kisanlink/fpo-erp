package main

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"kisanlink-erp/tests/integration"
	"kisanlink-erp/tests/testutils"

	"github.com/stretchr/testify/mock"
)

func TestDebugWebhook(t *testing.T) {
	ctx, cleanup := integration.SetupWebhookIntegration(t)
	defer cleanup()

	webhook := testutils.FixtureOrderCreatedWebhook("ORDER-DEBUG")
	payload := testutils.MustMarshalWebhook(webhook)

	// First request
	headers1 := testutils.GenerateWebhookHeadersWithEventID(payload, "test-webhook-secret", "evt_001")
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000001", nil).
		Once()

	req1 := httptest.NewRequest("POST", "/api/v1/webhooks/ecommerce/order/created", bytes.NewReader(payload))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Webhook-Signature", headers1.Signature)
	req1.Header.Set("X-Event-ID", headers1.EventID)
	req1.Header.Set("X-Timestamp", headers1.Timestamp)

	w1 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w1, req1)

	fmt.Printf("First Request Status: %d\n", w1.Code)
	fmt.Printf("First Request Body: %s\n", w1.Body.String())

	// Second request
	headers2 := testutils.GenerateWebhookHeadersWithEventID(payload, "test-webhook-secret", "evt_002")
	ctx.MockWebhookService.On("ProcessOrderCreated", mock.Anything, mock.AnythingOfType("*models.OrderCreatedWebhook")).
		Return("PORD00000002", nil).
		Once()

	req2 := httptest.NewRequest("POST", "/api/v1/webhooks/ecommerce/order/created", bytes.NewReader(payload))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Webhook-Signature", headers2.Signature)
	req2.Header.Set("X-Event-ID", headers2.EventID)
	req2.Header.Set("X-Timestamp", headers2.Timestamp)

	w2 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w2, req2)

	fmt.Printf("Second Request Status: %d\n", w2.Code)
	fmt.Printf("Second Request Body: %s\n", w2.Body.String())
}
