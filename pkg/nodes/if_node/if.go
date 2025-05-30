package if_node

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type ifDefinition struct{}

func (ifDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:      "if",
		Label:     "If",
		Icon:      "❓",
		Category:  "Control",
		Branching: true,
		Parameters: []api.ParameterDefinition{
			api.NewArrayParameter("conditions", "Conditions", true).
				WithGroup("Logic").
				WithDescription("Array of condition objects with expression and branch properties").
				WithDefault([]interface{}{
					map[string]interface{}{
						"expression": "true",
						"branch":     "default",
					},
				}).
				WithItemSchema(
					api.NewStringParameter("expression", "Expression", true).
						WithDescription("Boolean expression to evaluate (e.g., 'input.value > 10')"),
					api.NewStringParameter("branch", "Branch", true).
						WithDescription("Branch name to return if condition matches"),
				),
			api.NewBooleanParameter("hasElse", "Has Else Branch", false).
				WithDefault(false).
				WithGroup("Logic").
				WithDescription("Whether to include an else branch for unmatched conditions"),
		},
	}
}
// ConditionSpec represents a single if/else-if condition
type ConditionSpec struct {
	Expression string `json:"expression"`
	Branch     string `json:"branch"`
}

// IfResult represents the result of if node execution with branch information
type IfResult struct {
	Input       interface{} `json:"input"`
	Branch      string      `json:"branch"`
	Matched     bool        `json:"matched"`
	Expression  string      `json:"expression,omitempty"`
}

func (ifDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	// Parse conditions from array parameter
	conditionsRaw, ok := node.Data["conditions"]
	if !ok {
		return &IfResult{Input: input, Branch: "else", Matched: false}, nil
	}

	conditionsArray, ok := conditionsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("if: conditions must be an array")
	}

	conditions := make([]ConditionSpec, 0, len(conditionsArray))
	for i, condRaw := range conditionsArray {
		condMap, ok := condRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("if: condition %d must be an object", i)
		}

		expression, ok := condMap["expression"].(string)
		if !ok {
			return nil, fmt.Errorf("if: condition %d missing expression string", i)
		}

		branch, ok := condMap["branch"].(string)
		if !ok {
			return nil, fmt.Errorf("if: condition %d missing branch string", i)
		}

		conditions = append(conditions, ConditionSpec{
			Expression: expression,
			Branch:     branch,
		})
	}

	hasElse, _ := node.Data["hasElse"].(bool)

	// Evaluate conditions in order
	for _, condition := range conditions {
		matched, err := evaluateExpression(condition.Expression, input, ctx)
		if err != nil {
			return nil, fmt.Errorf("if: error evaluating expression '%s': %w", condition.Expression, err)
		}

		if matched {
			return &IfResult{
				Input:      input,
				Branch:     condition.Branch,
				Matched:    true,
				Expression: condition.Expression,
			}, nil
		}
	}

	// No conditions matched
	if hasElse {
		return &IfResult{Input: input, Branch: "else", Matched: false}, nil
	}

	// No else branch, return original input with no match indication
	return &IfResult{Input: input, Branch: "", Matched: false}, nil
}

// evaluateExpression evaluates a simple boolean expression against input data
func evaluateExpression(expression string, input interface{}, ctx api.ExecutionContext) (bool, error) {
	// Handle literal boolean values
	expression = strings.TrimSpace(expression)
	if expression == "true" {
		return true, nil
	}
	if expression == "false" {
		return false, nil
	}

	// Convert input to map for property access
	var data map[string]interface{}
	switch v := input.(type) {
	case map[string]interface{}:
		// Wrap the input map under "input" key and also add direct access
		data = make(map[string]interface{})
		data["input"] = v
		// Also add direct access to properties for backward compatibility
		for k, val := range v {
			data[k] = val
		}
	case *IfResult:
		// Handle chained if nodes
		data = map[string]interface{}{
			"input":      v.Input,
			"branch":     v.Branch,
			"matched":    v.Matched,
			"expression": v.Expression,
		}
	default:
		// Wrap primitive values
		data = map[string]interface{}{"input": input}
	}

	// Add context variables
	if ctx.Variables != nil {
		for k, v := range ctx.Variables {
			data[k] = v
		}
	}

	// Simple expression evaluation for common patterns
	return evaluateSimpleExpression(expression, data)
}

// evaluateSimpleExpression handles basic comparison operations
func evaluateSimpleExpression(expr string, data map[string]interface{}) (bool, error) {
	expr = strings.TrimSpace(expr)
	
	// Handle simple equality checks: input.field == "value"
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid == expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		return compareValues(leftVal, rightVal, "==")
	}
	
	// Handle inequality checks: input.field != "value"
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid != expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		result, err := compareValues(leftVal, rightVal, "==")
		return !result, err
	}
	
	// Handle greater than: input.field > value
	if strings.Contains(expr, ">") && !strings.Contains(expr, ">=") {
		parts := strings.SplitN(expr, ">", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid > expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		return compareValues(leftVal, rightVal, ">")
	}
	
	// Handle greater than or equal: input.field >= value
	if strings.Contains(expr, ">=") {
		parts := strings.SplitN(expr, ">=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid >= expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		return compareValues(leftVal, rightVal, ">=")
	}
	
	// Handle less than: input.field < value
	if strings.Contains(expr, "<") && !strings.Contains(expr, "<=") {
		parts := strings.SplitN(expr, "<", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid < expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		return compareValues(leftVal, rightVal, "<")
	}
	
	// Handle less than or equal: input.field <= value
	if strings.Contains(expr, "<=") {
		parts := strings.SplitN(expr, "<=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid <= expression: %s", expr)
		}
		
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		leftVal, err := getValue(left, data)
		if err != nil {
			return false, err
		}
		
		rightVal, err := parseValue(right)
		if err != nil {
			return false, err
		}
		
		return compareValues(leftVal, rightVal, "<=")
	}
	
	// Handle simple property existence/truthiness check
	val, err := getValue(expr, data)
	if err != nil {
		return false, err
	}
	
	return isTruthy(val), nil
}

// getValue extracts a value from data using dot notation (e.g., "input.field")
func getValue(path string, data map[string]interface{}) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part], nil
		}
		
		// Navigate deeper
		next, ok := current[part]
		if !ok {
			return nil, fmt.Errorf("property '%s' not found in path '%s'", part, path)
		}
		
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot navigate through non-object at '%s' in path '%s'", part, path)
		}
		
		current = nextMap
	}
	
	return nil, fmt.Errorf("empty path")
}

// parseValue parses a string value, handling quotes, numbers, and booleans
func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)
	
	// Handle quoted strings
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1], nil
	}
	
	// Handle booleans
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	
	// Handle numbers
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val, nil
	}
	
	// Default to string
	return s, nil
}

// compareValues compares two values based on the operator
func compareValues(left, right interface{}, op string) (bool, error) {
	switch op {
	case "==":
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
		
	case ">", ">=", "<", "<=":
		leftNum, leftOk := toNumber(left)
		rightNum, rightOk := toNumber(right)
		
		if !leftOk || !rightOk {
			return false, fmt.Errorf("cannot compare non-numeric values with %s", op)
		}
		
		switch op {
		case ">":
			return leftNum > rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}
	
	return false, fmt.Errorf("unsupported operator: %s", op)
}

// toNumber converts a value to float64 if possible
func toNumber(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// isTruthy determines if a value is considered true
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case float64:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	default:
		return true // Non-nil objects are truthy
	}
}

func (ifDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(ifDefinition{})
}

// assert that ifDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*ifDefinition)(nil)
