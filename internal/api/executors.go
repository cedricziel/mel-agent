package api

import (
   "bytes"
   "encoding/json"
   "errors"
   "io/ioutil"
   "net/http"
   "time"
)

// Node represents a workflow node with its configuration.
type Node struct {
   ID   string                 `json:"id"`
   Type string                 `json:"type"`
   Data map[string]interface{} `json:"data"`
}

// NodeExecutor defines execution logic for a node type.
type NodeExecutor interface {
   Execute(agentID string, node Node, input interface{}) (interface{}, error)
}


// DefaultExecutor performs no-op: returns input unchanged.
type DefaultExecutor struct{}

func (DefaultExecutor) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
   return input, nil
}

// IfExecutor filters items based on a boolean condition.
type IfExecutor struct{}

func (IfExecutor) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
   condRaw, ok := node.Data["condition"]
   if !ok {
       return input, nil
   }
   condStr, ok := condRaw.(string)
   if !ok {
       return input, nil
   }
   // simplistic: only "true" (case-insensitive) passes
   if condStr == "true" || condStr == "True" {
       return input, nil
   }
   // item is filtered out: return nil without error
   return nil, nil
}

// HTTPRequestExecutor performs an HTTP call as configured in node.Data.
type HTTPRequestExecutor struct{}

func (HTTPRequestExecutor) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
   urlRaw, ok := node.Data["url"].(string)
   if !ok || urlRaw == "" {
       return nil, errors.New("http: missing url parameter")
   }
   method := "GET"
   if m, ok := node.Data["method"].(string); ok && m != "" {
       method = m
   }
   var bodyReader *bytes.Reader
   if b, ok := node.Data["body"].(string); ok && b != "" {
       bodyReader = bytes.NewReader([]byte(b))
   } else {
       bodyReader = bytes.NewReader(nil)
   }
   req, err := http.NewRequest(method, urlRaw, bodyReader)
   if err != nil {
       return nil, err
   }
   // TODO: support headers from node.Data["headers"]
   resp, err := http.DefaultClient.Do(req)
   if err != nil {
       return nil, err
   }
   defer resp.Body.Close()
   respBody, err := ioutil.ReadAll(resp.Body)
   if err != nil {
       return nil, err
   }
   var result interface{}
   if err := json.Unmarshal(respBody, &result); err != nil {
       // return raw string on JSON unmarshal error
       return string(respBody), nil
   }
   return result, nil
}

// TimerExecutor produces a timestamp payload for timer triggers.
type TimerExecutor struct{}

func (TimerExecutor) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
   now := map[string]interface{}{"now": time.Now().UTC().Format(time.RFC3339)}
   // optionally include interval if specified (in seconds)
   if iv, ok := node.Data["interval"].(float64); ok {
       now["interval"] = iv
   }
   return now, nil
}

// ScheduleExecutor emits a scheduled event with current timestamp and cron spec.
type ScheduleExecutor struct{}

func (ScheduleExecutor) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
   cronSpec, _ := node.Data["cron"].(string)
   return map[string]interface{}{
       "now": time.Now().UTC().Format(time.RFC3339),
       "cron": cronSpec,
   }, nil
}
