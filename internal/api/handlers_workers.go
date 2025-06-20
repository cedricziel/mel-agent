package api

import (
	"context"
	"time"
)

// ListWorkers retrieves all workers
func (h *OpenAPIHandlers) ListWorkers(ctx context.Context, request ListWorkersRequestObject) (ListWorkersResponseObject, error) {
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, hostname, status, last_heartbeat, max_concurrent_steps, started_at FROM workflow_workers ORDER BY started_at DESC")
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkers500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var worker Worker
		var id, hostname, status string
		var lastHeartbeat, startedAt time.Time
		var maxConcurrentSteps int

		err := rows.Scan(&id, &hostname, &status, &lastHeartbeat, &maxConcurrentSteps, &startedAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListWorkers500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		worker.Id = &id
		worker.Name = &hostname // Use hostname as name
		worker.Status = func() *WorkerStatus {
			if status == "idle" {
				s := WorkerStatusInactive // Map "idle" to "inactive"
				return &s
			} else if status == "active" {
				s := WorkerStatusActive
				return &s
			}
			return nil
		}()
		worker.LastHeartbeat = &lastHeartbeat
		worker.Concurrency = &maxConcurrentSteps
		worker.RegisteredAt = &startedAt

		workers = append(workers, worker)
	}

	return ListWorkers200JSONResponse(workers), nil
}

// RegisterWorker registers a new worker
func (h *OpenAPIHandlers) RegisterWorker(ctx context.Context, request RegisterWorkerRequestObject) (RegisterWorkerResponseObject, error) {
	now := time.Now()

	hostname := request.Body.Id
	if request.Body.Name != nil {
		hostname = *request.Body.Name
	}

	concurrency := 5
	if request.Body.Concurrency != nil {
		concurrency = *request.Body.Concurrency
	}

	// Insert worker into database
	_, err := h.db.ExecContext(ctx,
		"INSERT INTO workflow_workers (id, hostname, status, last_heartbeat, max_concurrent_steps, started_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET hostname = $2, status = $3, last_heartbeat = $4, max_concurrent_steps = $5",
		request.Body.Id, hostname, "active", now, concurrency, now)
	if err != nil {
		errorMsg := "failed to register worker"
		message := err.Error()
		return RegisterWorker500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	status := WorkerStatusActive
	worker := Worker{
		Id:            &request.Body.Id,
		Name:          &hostname,
		Status:        &status,
		LastHeartbeat: &now,
		Concurrency:   &concurrency,
		RegisteredAt:  &now,
	}

	return RegisterWorker201JSONResponse(worker), nil
}

// UnregisterWorker removes a worker
func (h *OpenAPIHandlers) UnregisterWorker(ctx context.Context, request UnregisterWorkerRequestObject) (UnregisterWorkerResponseObject, error) {
	result, err := h.db.ExecContext(ctx, "DELETE FROM workflow_workers WHERE id = $1", request.Id)
	if err != nil {
		errorMsg := "failed to unregister worker"
		message := err.Error()
		return UnregisterWorker500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check unregister result"
		message := err.Error()
		return UnregisterWorker500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Worker not found"
		return UnregisterWorker404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return UnregisterWorker204Response{}, nil
}

// UpdateWorkerHeartbeat updates a worker's heartbeat
func (h *OpenAPIHandlers) UpdateWorkerHeartbeat(ctx context.Context, request UpdateWorkerHeartbeatRequestObject) (UpdateWorkerHeartbeatResponseObject, error) {
	now := time.Now()

	result, err := h.db.ExecContext(ctx,
		"UPDATE workflow_workers SET last_heartbeat = $1, status = 'active' WHERE id = $2",
		now, request.Id)
	if err != nil {
		errorMsg := "failed to update heartbeat"
		message := err.Error()
		return UpdateWorkerHeartbeat500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check heartbeat update result"
		message := err.Error()
		return UpdateWorkerHeartbeat500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Worker not found"
		return UpdateWorkerHeartbeat404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return UpdateWorkerHeartbeat200Response{}, nil
}

// ClaimWork allows a worker to claim work items
func (h *OpenAPIHandlers) ClaimWork(ctx context.Context, request ClaimWorkRequestObject) (ClaimWorkResponseObject, error) {
	// For now, return empty array as work queue implementation would be complex
	// In a real implementation, this would:
	// 1. Query for available work items
	// 2. Mark them as claimed by this worker
	// 3. Return the work items

	var workItems []WorkItem
	return ClaimWork200JSONResponse(workItems), nil
}

// CompleteWork marks a work item as completed
func (h *OpenAPIHandlers) CompleteWork(ctx context.Context, request CompleteWorkRequestObject) (CompleteWorkResponseObject, error) {
	// For now, just return success
	// In a real implementation, this would:
	// 1. Verify the work item belongs to this worker
	// 2. Update the work item status
	// 3. Store the result/error
	// 4. Potentially trigger follow-up actions

	return CompleteWork200Response{}, nil
}
