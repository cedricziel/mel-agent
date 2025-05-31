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
			api.NewStringParameter("connectionId", "Baserow Connection", true).
				WithGroup("Settings").
				WithDescription("ID of the Baserow connection"),
			api.NewEnumParameter("operation", "Operation", []string{"list_rows", "get_row", "create_row", "update_row", "delete_row", "list_databases", "list_tables"}, true).
				WithDefault("list_rows").
				WithGroup("Operation").
				WithDescription("Baserow operation to perform"),
			api.NewStringParameter("databaseId", "Database", false).
				WithGroup("Configuration").
				WithDescription("Database to work with").
				WithVisibilityCondition("operation!='list_databases'"),
			api.NewStringParameter("tableId", "Table", false).
				WithGroup("Configuration").
				WithDescription("Table to work with").
				WithVisibilityCondition("operation!='list_databases' && operation!='list_tables'"),
			api.NewNumberParameter("rowId", "Row ID", false).
				WithGroup("Configuration").
				WithDescription("Row ID for get/update/delete operations").
				WithVisibilityCondition("operation=='get_row' || operation=='update_row' || operation=='delete_row'"),
			api.NewObjectParameter("rowData", "Row Data", false).
				WithGroup("Data").
				WithDescription("Row data for create/update operations").
				WithVisibilityCondition("operation=='create_row' || operation=='update_row'"),
			api.NewNumberParameter("page", "Page", false).
				WithDefault(1).
				WithGroup("Pagination").
				WithDescription("Page number for list operations").
				WithVisibilityCondition("operation=='list_rows'"),
			api.NewNumberParameter("size", "Page Size", false).
				WithDefault(100).
				WithGroup("Pagination").
				WithDescription("Number of rows per page").
				WithVisibilityCondition("operation=='list_rows'"),
			api.NewStringParameter("search", "Search", false).
				WithGroup("Filtering").
				WithDescription("Search term to filter rows").
				WithVisibilityCondition("operation=='list_rows'"),
			api.NewStringParameter("orderBy", "Order By", false).
				WithGroup("Sorting").
				WithDescription("Field to order by (prefix with - for descending)").
				WithVisibilityCondition("operation=='list_rows'"),
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
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
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
	
	req.Header.Set("Authorization", "Token "+c.Token)
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
	resp, err := c.makeRequest("GET", "/api/applications/", nil)
	if err != nil {
		return nil, err
	}
	
	var databases []Database
	if err := json.Unmarshal(resp, &databases); err != nil {
		return nil, fmt.Errorf("failed to parse databases response: %w", err)
	}
	
	return databases, nil
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
	// Get connection ID
	connID, ok := node.Data["connectionId"].(string)
	if !ok || connID == "" {
		return nil, api.NewNodeError(node.ID, node.Type, "connectionId is required")
	}

	// Load connection configuration
	var secretJSON, configJSON []byte
	err := db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, connID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("connection %s not found", connID))
		}
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to load connection: %v", err))
	}

	// Parse connection secret and config
	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("invalid connection secret: %v", err))
	}

	if conn.BaseURL == "" || conn.Token == "" {
		return nil, api.NewNodeError(node.ID, node.Type, "baseURL and token are required in connection")
	}

	// Create Baserow client
	client := NewBaserowClient(conn.BaseURL, conn.Token)

	// Get operation
	operation, _ := node.Data["operation"].(string)
	if operation == "" {
		operation = "list_rows"
	}

	var result interface{}

	switch operation {
	case "list_databases":
		databases, err := client.ListDatabases()
		if err != nil {
			return nil, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("failed to list databases: %v", err))
		}
		result = map[string]interface{}{
			"databases": databases,
			"count":     len(databases),
		}

	case "list_tables":
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

	case "list_rows":
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

	case "get_row":
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

	case "create_row":
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

	case "update_row":
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

	case "delete_row":
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
	switch parameterName {
	case "databaseId":
		return d.getDatabaseOptions(dependencies)
	case "tableId":
		return d.getTableOptions(dependencies)
	default:
		return nil, fmt.Errorf("parameter %s does not support dynamic options", parameterName)
	}
}

// getDatabaseOptions loads databases for a Baserow connection
func (d baserowDefinition) getDatabaseOptions(dependencies map[string]interface{}) ([]api.OptionChoice, error) {
	connectionID, ok := dependencies["connectionId"].(string)
	if !ok || connectionID == "" {
		return nil, fmt.Errorf("connectionId is required")
	}

	// Load connection configuration (same as in ExecuteEnvelope)
	var secretJSON, configJSON []byte
	err := db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, connectionID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connection %s not found", connectionID)
		}
		return nil, fmt.Errorf("failed to load connection: %v", err)
	}

	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, fmt.Errorf("invalid connection secret: %v", err)
	}

	if conn.BaseURL == "" || conn.Token == "" {
		return nil, fmt.Errorf("baseURL and token are required in connection")
	}

	// Create client and fetch databases
	client := NewBaserowClient(conn.BaseURL, conn.Token)
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
	connectionID, ok := dependencies["connectionId"].(string)
	if !ok || connectionID == "" {
		return nil, fmt.Errorf("connectionId is required")
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
	err = db.DB.QueryRow(`SELECT secret, config FROM connections WHERE id = $1`, connectionID).Scan(&secretJSON, &configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connection %s not found", connectionID)
		}
		return nil, fmt.Errorf("failed to load connection: %v", err)
	}

	var conn BaserowConnection
	if err := json.Unmarshal(secretJSON, &conn); err != nil {
		return nil, fmt.Errorf("invalid connection secret: %v", err)
	}

	if conn.BaseURL == "" || conn.Token == "" {
		return nil, fmt.Errorf("baseURL and token are required in connection")
	}

	// Create client and fetch tables
	client := NewBaserowClient(conn.BaseURL, conn.Token)
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