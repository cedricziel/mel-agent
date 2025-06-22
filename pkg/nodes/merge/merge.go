package merge

import (
	"encoding/json"
	"fmt"

	"dario.cat/mergo"

	api "github.com/cedricziel/mel-agent/pkg/api"
)

// mergeDefinition provides the built-in "Merge" node.
type mergeDefinition struct{}

// Meta returns metadata for the Merge node.
func (mergeDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "merge",
		Label:    "Merge",
		Icon:     "🔀",
		Category: "Control",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("strategy", "Strategy", []string{"concat", "union", "deep", "intersection"}, true).
				WithDefault("concat").
				WithGroup("Settings").
				WithDescription("Merge strategy"),
		},
	}
}

// ExecuteEnvelope merges array or object items according to the selected strategy.
//
// The input envelope is expected to contain an array of items. When the items
// themselves are arrays they will be flattened. When they are objects, a map
// merge will be performed. The following strategies are supported:
//   - "concat": simply appends all elements in order.
//   - "union":  deduplicates array elements or merges object keys.
//   - "deep":   recursively merges nested maps using last-value wins semantics.
//   - "intersection": keeps only values present across all items.
func (d mergeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	strategy, _ := node.Data["strategy"].(string)
	if strategy == "" {
		strategy = "concat"
	}
	valid := map[string]bool{"concat": true, "union": true, "deep": true, "intersection": true}
	if !valid[strategy] {
		err := fmt.Errorf("invalid merge strategy %q", strategy)
		envelope.AddError(node.ID, "invalid strategy", err)
		return envelope, err
	}

	dataSlice, ok := envelope.Data.([]interface{})
	if !ok {
		// Nothing to merge - just passthrough
		result := envelope.Clone()
		result.Trace = envelope.Trace.Next(node.ID)
		return result, nil
	}

	// Detect the type of the first non-nil item to decide how to merge
	var first interface{}
	for _, v := range dataSlice {
		if v != nil {
			first = v
			break
		}
	}

	var resultData interface{}

	switch first.(type) {
	case map[string]interface{}:
		switch strategy {
		case "deep":
			merged := make(map[string]interface{})
			for _, item := range dataSlice {
				if m, ok := item.(map[string]interface{}); ok {
					if err := mergo.Merge(&merged, m, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
						envelope.AddError(node.ID, "deep merge failed", err)
						return envelope, err
					}
				}
			}
			resultData = merged
		case "intersection":
			resultData = intersectMaps(dataSlice)
		default: // concat or union (union handles maps like override)
			merged := make(map[string]interface{})
			for _, item := range dataSlice {
				if m, ok := item.(map[string]interface{}); ok {
					for k, v := range m {
						merged[k] = v
					}
				}
			}
			resultData = merged
		}
	default:
		switch strategy {
		case "intersection":
			resultData = intersectValues(dataSlice)
		default: // concat or union or deep
			var merged []interface{}
			seen := make(map[string]bool)
			for _, item := range dataSlice {
				if arr, ok := item.([]interface{}); ok {
					for _, sub := range arr {
						if strategy == "union" {
							key := dedupKey(sub)
							if seen[key] {
								continue
							}
							seen[key] = true
						}
						merged = append(merged, sub)
					}
					continue
				}

				if strategy == "union" {
					key := dedupKey(item)
					if seen[key] {
						continue
					}
					seen[key] = true
				}
				merged = append(merged, item)
			}
			resultData = merged
		}
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = resultData
	result.DataType = inferDataType(resultData)
	return result, nil
}

func (mergeDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(mergeDefinition{})
}

// assert that mergeDefinition implements the interface
var _ api.NodeDefinition = (*mergeDefinition)(nil)

// inferDataType replicates core.inferDataType without exporting it
func inferDataType(data interface{}) string {
	if data == nil {
		return "null"
	}
	switch data.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// intersectMaps returns a map containing only keys present in all maps.
func intersectMaps(items []interface{}) map[string]interface{} {
	if len(items) == 0 {
		return map[string]interface{}{}
	}
	if len(items) == 1 {
		m, _ := items[0].(map[string]interface{})
		return m
	}

	result := map[string]interface{}{}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range first {
		result[k] = v
	}

	for _, item := range items[1:] {
		m, ok := item.(map[string]interface{})
		if !ok {
			return map[string]interface{}{}
		}
		for key, existing := range result {
			val, ok := m[key]
			if !ok {
				delete(result, key)
				continue
			}

			switch ev := existing.(type) {
			case map[string]interface{}:
				if mv, ok := val.(map[string]interface{}); ok {
					result[key] = intersectMaps([]interface{}{ev, mv})
				} else {
					result[key] = val
				}
			case []interface{}:
				if mv, ok := val.([]interface{}); ok {
					result[key] = intersectValues([]interface{}{ev, mv})
				} else {
					result[key] = val
				}
			default:
				result[key] = val
			}
		}
	}

	return result
}

// intersectValues returns slice elements that occur in every array/item.
func intersectValues(items []interface{}) []interface{} {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		if arr, ok := items[0].([]interface{}); ok {
			return arr
		}
		return []interface{}{items[0]}
	}

	counts := map[string]struct {
		val   interface{}
		count int
	}{}
	total := 0

	for _, item := range items {
		total++
		seen := map[string]bool{}
		if arr, ok := item.([]interface{}); ok {
			for _, v := range arr {
				key := dedupKey(v)
				if !seen[key] {
					seen[key] = true
					c := counts[key]
					if c.count == 0 {
						c.val = v
					}
					c.count++
					counts[key] = c
				}
			}
		} else {
			key := dedupKey(item)
			if !seen[key] {
				seen[key] = true
				c := counts[key]
				if c.count == 0 {
					c.val = item
				}
				c.count++
				counts[key] = c
			}
		}
	}

	var result []interface{}
	for _, c := range counts {
		if c.count == total {
			result = append(result, c.val)
		}
	}
	return result
}

// dedupKey returns a stable string key for arbitrary values used in sets.
func dedupKey(v interface{}) string {
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%T:%#v", v, v)
}
