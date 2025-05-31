package baserow

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/api"
)

type baserowDefinition struct{}

func (baserowDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "baserow",
		Label:    "Baserow",
		Icon:     "🗃️",
		Category: "Database",
		Parameters: []api.ParameterDefinition{
			api.NewCredentialParameter("credentialId", "Baserow Credential", "baserow_jwt", true).
				WithGroup("Connection").
				WithDescription("Select your Baserow credential"),
			api.NewEnumParameter("resource", "Resource", []string{"databases", "tables", "rows"}, true).
				WithDefault("rows").
				WithGroup("Resource").
				WithDescription("What type of resource to work with"),
			api.NewEnumParameter("operation", "Operation", []string{"list", "get", "create", "update", "delete"}, true).
				WithDefault("list").
				WithGroup("Resource").
				WithDescription("Operation to perform on the resource"),
			// Dynamic fields - these get populated based on resource/operation
			api.NewStringParameter("databaseId", "Database", false).
				WithGroup("Target").
				WithDescription("Database to work with").
				WithVisibilityCondition("resource=='tables' || resource=='rows'").
				WithDynamicOptions(),
			api.NewStringParameter("tableId", "Table", false).
				WithGroup("Target").
				WithDescription("Table to work with").
				WithVisibilityCondition("resource=='rows'").
				WithDynamicOptions(),
			// Resource-specific parameters
			api.NewNumberParameter("rowId", "Row ID", false).
				WithGroup("Parameters").
				WithDescription("Row ID for get/update/delete operations").
				WithVisibilityCondition("resource=='rows' && (operation=='get' || operation=='update' || operation=='delete')"),
			api.NewObjectParameter("rowData", "Row Data", false).
				WithGroup("Parameters").
				WithDescription("Row data for create/update operations").
				WithVisibilityCondition("resource=='rows' && (operation=='create' || operation=='update')"),
			api.NewNumberParameter("page", "Page", false).
				WithDefault(1).
				WithGroup("Parameters").
				WithDescription("Page number for list operations").
				WithVisibilityCondition("operation=='list'"),
			api.NewNumberParameter("size", "Page Size", false).
				WithDefault(100).
				WithGroup("Parameters").
				WithDescription("Number of rows per page").
				WithVisibilityCondition("operation=='list'"),
			api.NewStringParameter("search", "Search", false).
				WithGroup("Parameters").
				WithDescription("Search term to filter rows").
				WithVisibilityCondition("resource=='rows' && operation=='list'"),
			api.NewStringParameter("orderBy", "Order By", false).
				WithGroup("Parameters").
				WithDescription("Field to order by (prefix with - for descending)").
				WithVisibilityCondition("resource=='rows' && operation=='list'"),
		},
	}
}

// BaserowClient handles Baserow API interactions
type BaserowClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// BaserowConnection represents the connection configuration
type BaserowConnection struct {
	BaseURL  string `json:"baseUrl"`
	Token    string `json:"token,omitempty"`    // For token-based auth
	Username string `json:"username,omitempty"` // For JWT auth
	Password string `json:"password,omitempty"` // For JWT auth
}

// Database represents a Baserow database
type Database struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// Table represents a Baserow table
type Table struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Order        int    `json:"order"`
	DatabaseID   int    `json:"database_id"`
}

// Row represents a Baserow row
type Row struct {
	ID    int                    `json:"id"`
	Order string                 `json:"order"`
	Data  map[string]interface{} `json:"-"` // Will be populated with field data
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Count    int           `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []interface{} `json:"results"`
}

// JWTResponse represents a JWT authentication response
type JWTResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// NewBaserowClient creates a new Baserow API client
func NewBaserowClient(baseURL, token string) *BaserowClient {
	return &BaserowClient{
		BaseURL: baseURL,
		Token:   token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewBaserowClientWithJWT creates a new Baserow API client using JWT authentication
func NewBaserowClientWithJWT(baseURL, username, password string) (*BaserowClient, error) {
	client := &BaserowClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Authenticate to get JWT token
	token, err := client.authenticateJWT(username, password)
	if err != nil {
		return nil, fmt.Errorf("JWT authentication failed: %w", err)
	}

	client.Token = token
	return client, nil
}

// authenticateJWT performs JWT authentication with Baserow
func (c *BaserowClient) authenticateJWT(username, password string) (string, error) {
	authPayload := map[string]string{
		"email":    username, // Baserow uses "email" not "username"
		"password": password,
	}

	jsonBody, err := json.Marshal(authPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/user/token-auth/", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jwtResp struct {
		AccessToken string `json:"access_token"` // /api/user/token-auth/ now returns "access_token"
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwtResp); err != nil {
		return "", fmt.Errorf("failed to decode JWT response: %w", err)
	}

	return jwtResp.AccessToken, nil
}

// makeRequest performs an HTTP request to the Baserow API
func (c *BaserowClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := c.BaseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header - always use JWT format
	if c.Token != "" {
		req.Header.Set("Authorization", "JWT "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ListDatabases retrieves all databases accessible to the user
func (c *BaserowClient) ListDatabases() ([]Database, error) {
	// First get applications (workspaces in Baserow)
	resp, err := c.makeRequest("GET", "/api/applications/", nil)
	if err != nil {
		return nil, err
	}

	// Parse applications response
	var applications []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Order int    `json:"order"`
	}

	if err := json.Unmarshal(resp, &applications); err != nil {
		return nil, fmt.Errorf("failed to parse applications response: %w", err)
	}

	// Collect all databases from all applications
	var allDatabases []Database

	for _, app := range applications {
		// Only process database applications (skip other types like forms, etc.)
		if app.Type == "database" {
			// In Baserow, database applications ARE the databases
			database := Database{
				ID:    app.ID,
				Name:  app.Name,
				Order: app.Order,
			}
			allDatabases = append(allDatabases, database)
		}
	}

	return allDatabases, nil
}

// ListTables retrieves all tables in a database
func (c *BaserowClient) ListTables(databaseID int) ([]Table, error) {
	endpoint := fmt.Sprintf("/api/database/tables/database/%d/", databaseID)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var tables []Table
	if err := json.Unmarshal(resp, &tables); err != nil {
		return nil, fmt.Errorf("failed to parse tables response: %w", err)
	}

	return tables, nil
}

// ListRows retrieves rows from a table with optional filtering and pagination
func (c *BaserowClient) ListRows(tableID int, page, size int, search, orderBy string) (*ListResponse, error) {
	endpoint := fmt.Sprintf("/api/database/rows/table/%d/", tableID)

	// Add query parameters
	params := fmt.Sprintf("?page=%d&size=%d", page, size)
	if search != "" {
		params += "&search=" + search
	}
	if orderBy != "" {
		params += "&order_by=" + orderBy
	}

	resp, err := c.makeRequest("GET", endpoint+params, nil)
	if err != nil {
		return nil, err
	}

	var listResp ListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse rows response: %w", err)
	}

	return &listResp, nil
}

// GetRow retrieves a specific row by ID
func (c *BaserowClient) GetRow(tableID, rowID int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/api/database/rows/table/%d/%d/", tableID, rowID)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var row map[string]interface{}
	if err := json.Unmarshal(resp, &row); err != nil {
		return nil, fmt.Errorf("failed to parse row response: %w", err)
	}

	return row, nil
}

// CreateRow creates a new row in a table
func (c *BaserowClient) CreateRow(tableID int, rowData map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/api/database/rows/table/%d/", tableID)
	resp, err := c.makeRequest("POST", endpoint, rowData)
	if err != nil {
		return nil, err
	}

	var row map[string]interface{}
	if err := json.Unmarshal(resp, &row); err != nil {
		return nil, fmt.Errorf("failed to parse created row response: %w", err)
	}

	return row, nil
}

// UpdateRow updates an existing row
func (c *BaserowClient) UpdateRow(tableID, rowID int, rowData map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/api/database/rows/table/%d/%d/", tableID, rowID)
	resp, err := c.makeRequest("PATCH", endpoint, rowData)
	if err != nil {
		return nil, err
	}

	var row map[string]interface{}
	if err := json.Unmarshal(resp, &row); err != nil {
		return nil, fmt.Errorf("failed to parse updated row response: %w", err)
	}

	return row, nil
}

// DeleteRow deletes a row by ID
func (c *BaserowClient) DeleteRow(tableID, rowID int) error {
	endpoint := fmt.Sprintf("/api/database/rows/table/%d/%d/", tableID, rowID)
	_, err := c.makeRequest("DELETE", endpoint, nil)
	return err
}

// ExecuteEnvelope performs Baserow operations using envelopes
func (d baserowDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// Get credential ID
	credID, ok := node.Data["credentialId"].(string)
	if !ok || credID == "" {
		return nil, api.NewNodeError(node.ID, node.Type, "credentialId is required")
	}

	// Load credential configuration
	var secretJSON, configJSON []byte
	err := db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, credID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("credential %s not found", credID))
		}
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to load credential: %v", err))
	}

	// Parse connection secret and config
	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("invalid connection secret: %v", err))
	}

	if conn.BaseURL == "" {
		return nil, api.NewNodeError(node.ID, node.Type, "baseURL is required in connection")
	}

	// Create client with fresh JWT authentication
	if conn.Username == "" || conn.Password == "" {
		return nil, api.NewNodeError(node.ID, node.Type, "username and password are required for JWT authentication")
	}

	client, err2 := NewBaserowClientWithJWT(conn.BaseURL, conn.Username, conn.Password)
	if err2 != nil {
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to authenticate with Baserow: %v", err2))
	}

	// Get resource and operation
	resource, _ := node.Data["resource"].(string)
	if resource == "" {
		resource = "rows"
	}

	operation, _ := node.Data["operation"].(string)
	if operation == "" {
		operation = "list"
	}

	var result interface{}

	// Handle operations based on resource type
	resourceOperation := resource + "_" + operation

	switch resourceOperation {
	case "databases_list":
		databases, err := client.ListDatabases()
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to list databases: %v", err))
		}
		result = map[string]interface{}{
			"databases": databases,
			"count":     len(databases),
		}

	case "tables_list":
		databaseID, err := getIntParameter(node.Data, "databaseId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "databaseId is required for list_tables operation")
		}

		tables, err := client.ListTables(databaseID)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to list tables: %v", err))
		}
		result = map[string]interface{}{
			"tables":     tables,
			"count":      len(tables),
			"database_id": databaseID,
		}

	case "rows_list":
		tableID, err := getIntParameter(node.Data, "tableId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "tableId is required for list_rows operation")
		}

		page, _ := getIntParameter(node.Data, "page")
		if page <= 0 {
			page = 1
		}

		size, _ := getIntParameter(node.Data, "size")
		if size <= 0 {
			size = 100
		}

		search, _ := node.Data["search"].(string)
		orderBy, _ := node.Data["orderBy"].(string)

		listResp, err := client.ListRows(tableID, page, size, search, orderBy)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to list rows: %v", err))
		}
		result = listResp

	case "rows_get":
		tableID, err := getIntParameter(node.Data, "tableId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "tableId is required for get_row operation")
		}

		rowID, err := getIntParameter(node.Data, "rowId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "rowId is required for get_row operation")
		}

		row, err := client.GetRow(tableID, rowID)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to get row: %v", err))
		}
		result = row

	case "rows_create":
		tableID, err := getIntParameter(node.Data, "tableId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "tableId is required for create_row operation")
		}

		rowData, ok := node.Data["rowData"].(map[string]interface{})
		if !ok {
			return nil, api.NewNodeError(node.ID, node.Type, "rowData is required for create_row operation")
		}

		row, err := client.CreateRow(tableID, rowData)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to create row: %v", err))
		}
		result = row

	case "rows_update":
		tableID, err := getIntParameter(node.Data, "tableId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "tableId is required for update_row operation")
		}

		rowID, err := getIntParameter(node.Data, "rowId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "rowId is required for update_row operation")
		}

		rowData, ok := node.Data["rowData"].(map[string]interface{})
		if !ok {
			return nil, api.NewNodeError(node.ID, node.Type, "rowData is required for update_row operation")
		}

		row, err := client.UpdateRow(tableID, rowID, rowData)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to update row: %v", err))
		}
		result = row

	case "rows_delete":
		tableID, err := getIntParameter(node.Data, "tableId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "tableId is required for delete_row operation")
		}

		rowID, err := getIntParameter(node.Data, "rowId")
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, "rowId is required for delete_row operation")
		}

		err = client.DeleteRow(tableID, rowID)
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to delete row: %v", err))
		}
		result = map[string]interface{}{
			"deleted":  true,
			"row_id":   rowID,
			"table_id": tableID,
		}

	default:
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("unsupported operation: %s", operation))
	}

	// Create result envelope
	resultEnvelope := envelope.Clone()
	resultEnvelope.Trace = envelope.Trace.Next(node.ID)
	resultEnvelope.Data = result
	resultEnvelope.DataType = "object"

	return resultEnvelope, nil
}

// getIntParameter safely extracts an integer parameter from node data
func getIntParameter(data map[string]interface{}, key string) (int, error) {
	value, exists := data[key]
	if !exists {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("parameter %s is not a valid integer", key)
	}
}

func (baserowDefinition) Initialize(mel api.Mel) error {
	return nil
}

// GetDynamicOptions implements the DynamicOptionsProvider interface
func (d baserowDefinition) GetDynamicOptions(ctx api.ExecutionContext, parameterName string, dependencies map[string]interface{}) ([]api.OptionChoice, error) {
	// Check if credential is provided
	credentialID, ok := dependencies["credentialId"].(string)
	if !ok || credentialID == "" {
		return []api.OptionChoice{}, nil // Return empty options instead of error
	}

	switch parameterName {
	case "databaseId":
		// Always load databases when credential is available - users need to see them to select
		return d.getDatabaseOptions(dependencies)
	case "tableId":
		// Load tables whenever a database is selected - users need to see them to select
		databaseID, _ := dependencies["databaseId"].(string)
		if databaseID != "" {
			return d.getTableOptions(dependencies)
		}
		return []api.OptionChoice{}, nil // Return empty options when no database selected
	default:
		return nil, fmt.Errorf("parameter %s does not support dynamic options", parameterName)
	}
}

// getDatabaseOptions loads databases for a Baserow credential
func (d baserowDefinition) getDatabaseOptions(dependencies map[string]interface{}) ([]api.OptionChoice, error) {
	credentialID, ok := dependencies["credentialId"].(string)
	if !ok || credentialID == "" {
		return nil, fmt.Errorf("credentialId is required")
	}

	// Load credential configuration
	var secretJSON, configJSON []byte
	err := db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, credentialID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential %s not found", credentialID)
		}
		return nil, fmt.Errorf("failed to load connection: %v", err)
	}

	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, fmt.Errorf("invalid connection secret: %v", err)
	}

	if conn.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required in connection")
	}

	// Create client with fresh JWT authentication
	if conn.Username == "" || conn.Password == "" {
		return nil, fmt.Errorf("username and password are required for JWT authentication")
	}

	client, err2 := NewBaserowClientWithJWT(conn.BaseURL, conn.Username, conn.Password)
	if err2 != nil {
		return nil, fmt.Errorf("failed to authenticate with Baserow: %v", err2)
	}

	databases, err := client.ListDatabases()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %v", err)
	}

	// Convert to OptionChoice
	options := make([]api.OptionChoice, len(databases))
	for i, db := range databases {
		options[i] = api.OptionChoice{
			Value: fmt.Sprintf("%d", db.ID),
			Label: db.Name,
		}
	}

	return options, nil
}

// getTableOptions loads tables for a Baserow database
func (d baserowDefinition) getTableOptions(dependencies map[string]interface{}) ([]api.OptionChoice, error) {
	credentialID, ok := dependencies["credentialId"].(string)
	if !ok || credentialID == "" {
		return nil, fmt.Errorf("credentialId is required")
	}

	databaseIDStr, ok := dependencies["databaseId"].(string)
	if !ok || databaseIDStr == "" {
		return nil, fmt.Errorf("databaseId is required")
	}

	databaseID, err := strconv.Atoi(databaseIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid databaseId: %v", err)
	}

	// Load connection configuration
	var secretJSON, configJSON []byte
	err = db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, credentialID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential %s not found", credentialID)
		}
		return nil, fmt.Errorf("failed to load connection: %v", err)
	}

	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, fmt.Errorf("invalid connection secret: %v", err)
	}

	if conn.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required in connection")
	}

	// Create client with fresh JWT authentication
	if conn.Username == "" || conn.Password == "" {
		return nil, fmt.Errorf("username and password are required for JWT authentication")
	}

	client, err2 := NewBaserowClientWithJWT(conn.BaseURL, conn.Username, conn.Password)
	if err2 != nil {
		return nil, fmt.Errorf("failed to authenticate with Baserow: %v", err2)
	}

	tables, err := client.ListTables(databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}

	// Convert to OptionChoice
	options := make([]api.OptionChoice, len(tables))
	for i, table := range tables {
		options[i] = api.OptionChoice{
			Value: fmt.Sprintf("%d", table.ID),
			Label: table.Name,
		}
	}

	return options, nil
}

func init() {
	api.RegisterNodeDefinition(baserowDefinition{})
}

// assert that baserowDefinition implements both interfaces
var _ api.NodeDefinition = (*baserowDefinition)(nil)
var _ api.DynamicOptionsProvider = (*baserowDefinition)(nil)
