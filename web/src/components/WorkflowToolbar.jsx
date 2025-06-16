import React from 'react';

/**
 * WorkflowToolbar component - handles the top toolbar in the workflow builder
 * Contains view mode toggle, action buttons, and status indicators
 */
function WorkflowToolbar({
  isDraft,
  isDirty,
  isSaving,
  lastSaved,
  saveError,
  viewMode,
  testing,
  isLiveMode,
  sidebarTab,
  onViewModeChange,
  onAddNode,
  onSave,
  onTestRun,
  onAutoLayout,
  onToggleSidebar,
  onToggleLiveMode,
}) {
  return (
    <div className="absolute top-4 left-4 right-4 flex justify-between items-center">
      <div className="flex gap-2 items-center">
        {/* Draft/Auto-save status */}
        <div className="flex items-center gap-3 mr-4">
          {/* Draft vs Production indicator */}
          <div
            className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
              isDraft
                ? 'bg-yellow-100 text-yellow-800'
                : 'bg-green-100 text-green-800'
            }`}
          >
            <div
              className={`w-2 h-2 rounded-full ${
                isDraft ? 'bg-yellow-400' : 'bg-green-400'
              }`}
            ></div>
            {isDraft ? 'Draft' : 'Deployed'}
          </div>

          {/* Auto-save status */}
          {isDraft && (
            <div className="flex items-center gap-1 text-xs text-gray-600">
              {isSaving ? (
                <>
                  <div className="animate-spin w-3 h-3 border border-blue-500 border-t-transparent rounded-full"></div>
                  Saving...
                </>
              ) : saveError ? (
                <span className="text-red-600">Save failed</span>
              ) : lastSaved ? (
                <span>Saved {new Date(lastSaved).toLocaleTimeString()}</span>
              ) : (
                <span>Auto-save enabled</span>
              )}
            </div>
          )}
        </div>

        {/* View mode toggle switch */}
        <div className="flex bg-gray-100 rounded-lg p-1 mr-4">
          <button
            onClick={() => onViewModeChange('editor')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              viewMode === 'editor'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Editor
          </button>
          <button
            onClick={() => onViewModeChange('executions')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              viewMode === 'executions'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Executions
          </button>
        </div>

        {/* Action buttons */}
        <button
          onClick={onAddNode}
          disabled={viewMode === 'executions'}
          className={`px-4 py-2 rounded ${
            viewMode === 'executions'
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-blue-500 text-white'
          }`}
        >
          + Add Node
        </button>

        <button
          onClick={onSave}
          disabled={(!isDirty && !isDraft) || viewMode === 'executions'}
          className={`px-4 py-2 rounded ${
            (isDirty || isDraft) && viewMode === 'editor'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-300 text-gray-500'
          }`}
        >
          {isDraft ? 'Deploy' : 'Save'}
        </button>

        <button
          onClick={onTestRun}
          disabled={testing || viewMode === 'executions'}
          className={`px-4 py-2 rounded ${
            viewMode === 'executions'
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-green-500 text-white'
          }`}
        >
          {testing ? 'Running...' : 'Test Run'}
        </button>

        <button
          onClick={onAutoLayout}
          disabled={viewMode === 'executions'}
          className={`px-4 py-2 rounded ${
            viewMode === 'executions'
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-purple-500 text-white'
          }`}
        >
          Auto Layout
        </button>
      </div>

      {/* Right side buttons */}
      <div className="flex gap-2">
        <button
          onClick={() => onToggleSidebar('chat')}
          className={`px-4 py-2 rounded ${
            sidebarTab === 'chat' ? 'bg-blue-500 text-white' : 'bg-gray-300'
          }`}
        >
          💬 Chat
        </button>
        <button
          onClick={onToggleLiveMode}
          className={`px-4 py-2 rounded ${
            isLiveMode ? 'bg-orange-500 text-white' : 'bg-gray-300'
          }`}
        >
          {isLiveMode ? 'Live Mode' : 'Edit Mode'}
        </button>
      </div>
    </div>
  );
}

export default WorkflowToolbar;
