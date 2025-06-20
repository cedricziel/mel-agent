package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ListIntegrations lists available integrations
func (h *OpenAPIHandlers) ListIntegrations(ctx context.Context, request ListIntegrationsRequestObject) (ListIntegrationsResponseObject, error) {
	// Get integrations from database
	rows, err := h.db.QueryContext(ctx, "SELECT id, name, description, type, status, created_at, updated_at FROM integrations ORDER BY name")
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListIntegrations500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var integrations []Integration
	for rows.Next() {
		var integration Integration
		var id, name, description, integrationType, status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &name, &description, &integrationType, &status, &createdAt, &updatedAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListIntegrations500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		integrationUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListIntegrations500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		integration.Id = &integrationUUID
		integration.Name = &name
		integration.Description = &description
		integration.Type = &integrationType
		integration.Status = func() *IntegrationStatus { s := IntegrationStatus(status); return &s }()
		integration.CreatedAt = &createdAt
		integration.UpdatedAt = &updatedAt

		integrations = append(integrations, integration)
	}

	return ListIntegrations200JSONResponse(integrations), nil
}
