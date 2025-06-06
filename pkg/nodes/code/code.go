package code

import (
	"context"
	"fmt"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// Runtime defines the interface for executing code in different languages
type Runtime interface {
	Execute(code string, context CodeExecutionContext) (interface{}, error)
	GetLanguage() string
	Initialize() error
	Cleanup() error
}

// CodeExecutionContext provides context for code execution
type CodeExecutionContext struct {
	// Input data from envelope
	Data      interface{}            `json:"data"`
	Variables map[string]interface{} `json:"variables"`
	
	// Node context
	NodeData map[string]interface{} `json:"nodeData"`
	NodeID   string                 `json:"nodeId"`
	AgentID  string                 `json:"agentId"`
	
	// Platform utilities
	Mel api.Mel `json:"-"`
}

// codeDefinition provides the built-in "Code" node
type codeDefinition struct {
	runtimes map[string]Runtime
}

// Meta returns metadata for the Code node
func (codeDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "code",
		Label:    "Code",
		Category: "Code",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("language", "Language", 
				[]string{"javascript", "python", "typescript"}, true).
				WithDefault("javascript").
				WithGroup("Settings").
				WithDescription("Programming language to use"),
			func() api.ParameterDefinition {
				param := api.NewStringParameter("code", "Code", true).
					WithDescription("Code to execute").
					WithGroup("Code")
				param.JSONSchema = &api.JSONSchema{
					Type:   "string",
					Format: "code", // Custom format for UI that adapts to language
				}
				return param
			}(),
			api.NewNumberParameter("timeout", "Timeout (seconds)", false).
				WithDefault(30).
				WithGroup("Settings").
				WithDescription("Maximum execution time"),
			api.NewBooleanParameter("strict_mode", "Strict Mode", false).
				WithDefault(true).
				WithGroup("Settings").
				WithDescription("Enable strict mode (where applicable)"),
		},
	}
}

// ExecuteEnvelope executes user-provided code in the specified language
func (d *codeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	code, ok := node.Data["code"].(string)
	if !ok || code == "" {
		return nil, api.NewNodeError(node.ID, "code", "code parameter is required")
	}
	
	language, ok := node.Data["language"].(string)
	if !ok || language == "" {
		language = "javascript" // Default language
	}
	
	timeout, _ := node.Data["timeout"].(float64)
	if timeout <= 0 {
		timeout = 30
	}
	
	// Get runtime for selected language
	runtime, exists := d.runtimes[language]
	if !exists {
		return nil, api.NewNodeError(node.ID, "code", "unsupported language: "+language)
	}
	
	// Prepare execution context
	execContext := CodeExecutionContext{
		Data:      envelope.Data,
		Variables: envelope.Variables,
		NodeData:  node.Data,
		NodeID:    node.ID,
		AgentID:   ctx.AgentID,
		Mel:       ctx.Mel,
	}
	
	// Execute with timeout
	resultChan := make(chan codeExecutionResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- codeExecutionResult{
					Error: fmt.Errorf("panic during execution: %v", r),
				}
			}
		}()
		
		result, err := runtime.Execute(code, execContext)
		resultChan <- codeExecutionResult{
			Value: result,
			Error: err,
		}
	}()
	
	// Wait for result or timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	
	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, api.NewNodeError(node.ID, "code", "execution error: "+result.Error.Error())
		}
		
		resultEnvelope := envelope.Clone()
		resultEnvelope.Trace = envelope.Trace.Next(node.ID)
		resultEnvelope.Data = result.Value
		
		return resultEnvelope, nil
		
	case <-timeoutCtx.Done():
		return nil, api.NewNodeError(node.ID, "code", "execution timeout exceeded")
	}
}

// Initialize sets up the code execution environment
func (d *codeDefinition) Initialize(mel api.Mel) error {
	// Initialize all runtimes if not already done
	if d.runtimes == nil {
		d.runtimes = make(map[string]Runtime)
	}
	
	// Add JavaScript runtime if not already present
	if _, exists := d.runtimes["javascript"]; !exists {
		jsRuntime := NewJavaScriptRuntime()
		if err := jsRuntime.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize JavaScript runtime: %w", err)
		}
		d.runtimes["javascript"] = jsRuntime
	}
	
	// Future: Add other language runtimes here
	// if _, exists := d.runtimes["python"]; !exists {
	//     pythonRuntime := NewPythonRuntime()
	//     d.runtimes["python"] = pythonRuntime
	// }
	
	return nil
}

// codeExecutionResult holds the result of code execution
type codeExecutionResult struct {
	Value interface{}
	Error error
}

// NewCodeDefinition creates a new code node definition
func NewCodeDefinition() api.NodeDefinition {
	return &codeDefinition{
		runtimes: make(map[string]Runtime),
	}
}

func init() {
	api.RegisterNodeDefinition(NewCodeDefinition())
}

// Compile-time interface check
var _ api.NodeDefinition = (*codeDefinition)(nil)