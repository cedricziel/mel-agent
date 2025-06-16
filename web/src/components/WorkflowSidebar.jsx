import React from 'react';
import NodeDetailsPanel from './NodeDetailsPanel';
import ChatAssistant from './ChatAssistant';

/**
 * WorkflowSidebar component - handles the right sidebar in the workflow builder
 * Contains executions panel, node details panel, and chat assistant
 */
function WorkflowSidebar({
  isVisible,
  sidebarTab,
  viewMode,
  selectedNode,
  selectedNodeDef,
  selectedExecution,
  executions,
  loadingExecutions,
  isLiveMode,
  isDraft,
  agentId,
  triggersMap,
  onExecutionSelect,
  onNodeChange,
  onNodeExecute,
  onChatClose,
  onAddNode,
  onConnectNodes,
  onGetWorkflow,
}) {
  if (!isVisible) {
    return null;
  }

  return (
    <div
      data-testid="workflow-sidebar"
      className="w-80 bg-white border-l shadow-lg h-screen overflow-y-auto"
    >
      {/* Executions panel */}
      {viewMode === 'executions' && (
        <div className="h-full flex flex-col">
          <div className="p-4 border-b">
            <h3 className="text-lg font-semibold">Executions</h3>
            {loadingExecutions && (
              <div className="text-sm text-gray-500">Loading executions...</div>
            )}
          </div>
          <div className="flex-1 overflow-y-auto">
            {executions.length === 0 && !loadingExecutions ? (
              <div className="p-4 text-center text-gray-500">
                No executions found. Run your workflow to see execution history.
              </div>
            ) : (
              <div className="p-2">
                {executions.map((execution) => (
                  <div
                    key={execution.id}
                    onClick={() => onExecutionSelect(execution)}
                    className={`p-3 border rounded mb-2 cursor-pointer hover:bg-gray-50 ${
                      selectedExecution?.id === execution.id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">
                        {new Date(execution.created_at).toLocaleString()}
                      </span>
                      <span className="px-2 py-1 text-xs bg-green-100 text-green-800 rounded">
                        Completed
                      </span>
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      ID: {execution.id.slice(0, 8)}...
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Node Details Panel */}
      {sidebarTab === 'details' &&
        selectedNode &&
        selectedNodeDef &&
        viewMode === 'editor' && (
          <NodeDetailsPanel
            node={selectedNode}
            nodeDef={selectedNodeDef}
            readOnly={isLiveMode}
            onChange={onNodeChange}
            onExecute={onNodeExecute}
            publicUrl={
              selectedNodeDef.type === 'webhook' && triggersMap[selectedNode.id]
                ? `${window.location.origin}/api/webhooks/${selectedNode.type}/${triggersMap[selectedNode.id].id}`
                : undefined
            }
          />
        )}

      {/* Chat Assistant */}
      {sidebarTab === 'chat' && (
        <ChatAssistant
          inline
          agentId={agentId}
          onAddNode={onAddNode}
          onConnectNodes={onConnectNodes}
          onGetWorkflow={onGetWorkflow}
          onClose={onChatClose}
        />
      )}
    </div>
  );
}

export default WorkflowSidebar;
