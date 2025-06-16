import { useState, useCallback } from 'react';

export function useValidation() {
  const [validationErrors, setValidationErrors] = useState({});

  const validateWorkflow = useCallback((nodes, edges) => {
    const errorsMap = {};

    // Validate HTTP request nodes
    nodes.forEach((n) => {
      if (n.type === 'http_request') {
        const url = n.data.url || '';
        const method = n.data.method || '';
        if (!url.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Node "${n.data.label || n.id}" is missing a URL`
          );
        }
        if (!method.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Node "${n.data.label || n.id}" is missing a method`
          );
        }
      }
    });

    // Validate async webhook flows
    const nodeMap = Object.fromEntries(nodes.map((n) => [n.id, n]));
    nodes.forEach((n) => {
      if (n.type === 'webhook' && n.data.mode === 'async') {
        const visited = new Set();
        const queue = [n.id];
        let found = false;
        while (queue.length && !found) {
          const curr = queue.shift();
          visited.add(curr);
          for (const e of edges) {
            if (e.source === curr) {
              const tgt = e.target;
              if (visited.has(tgt)) continue;
              const child = nodeMap[tgt];
              if (child) {
                if (child.type === 'http_response') {
                  found = true;
                  break;
                }
                queue.push(tgt);
              }
            }
          }
        }
        if (!found) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Async Webhook node "${n.data.label || n.id}" must be followed by a Webhook Response node`
          );
        }
      }
    });

    setValidationErrors(errorsMap);
    return Object.keys(errorsMap).length === 0;
  }, []);

  const clearValidationErrors = useCallback(() => {
    setValidationErrors({});
  }, []);

  const getNodeValidationErrors = useCallback(
    (nodeId) => {
      return validationErrors[nodeId] || [];
    },
    [validationErrors]
  );

  const hasValidationErrors = useCallback(() => {
    return Object.keys(validationErrors).length > 0;
  }, [validationErrors]);

  return {
    validationErrors,
    validateWorkflow,
    clearValidationErrors,
    getNodeValidationErrors,
    hasValidationErrors,
  };
}
