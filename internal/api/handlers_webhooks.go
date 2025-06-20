package api

import (
	"context"
	"database/sql"
	"encoding/json"
)

// HandleWebhook processes incoming webhook requests
func (h *OpenAPIHandlers) HandleWebhook(ctx context.Context, request HandleWebhookRequestObject) (HandleWebhookResponseObject, error) {
	// Verify webhook token exists and get associated workflow/trigger
	var triggerID, workflowID string
	err := h.db.QueryRowContext(ctx,
		"SELECT t.id, t.workflow_id FROM triggers t WHERE t.type = 'webhook' AND t.config->>'token' = $1 AND t.enabled = true",
		request.Token).Scan(&triggerID, &workflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Webhook not found or disabled"
			return HandleWebhook404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return HandleWebhook500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// In a real implementation, this would:
	// 1. Validate the webhook payload based on trigger configuration
	// 2. Transform the webhook data if needed
	// 3. Trigger workflow execution via the engine
	// 4. Return appropriate response

	// For now, just log the webhook received and return success
	payloadJson, _ := json.Marshal(request.Body)

	// Store webhook event (optional)
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO webhook_events (trigger_id, payload, created_at) VALUES ($1, $2, NOW())",
		triggerID, payloadJson)
	if err != nil {
		// Log error but don't fail the webhook
		// In production, you might want to queue this for retry
	}

	// TODO: Trigger workflow execution
	// err = h.engine.TriggerWorkflow(ctx, workflowID, request.Body)
	// if err != nil {
	//     return HandleWebhook500JSONResponse{...}
	// }

	return HandleWebhook200Response{}, nil
}
