import { useState } from 'react';

const CONFIG_OPTIONS = {
  model: [
    {
      id: 'openai',
      type: 'openai_model',
      label: 'OpenAI',
      description: 'GPT-4, GPT-3.5, and other OpenAI models',
      icon: '🤖',
      defaultData: {
        provider: 'openai',
        model: 'gpt-4',
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      id: 'anthropic',
      type: 'anthropic_model',
      label: 'Anthropic',
      description: 'Claude 3.5 Sonnet, Claude 3 Opus, and other Claude models',
      icon: '🧠',
      defaultData: {
        provider: 'anthropic',
        model: 'claude-3-5-sonnet-20241022',
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
  ],
  tools: [
    {
      id: 'workflow_tools',
      type: 'workflow_tools',
      label: 'Workflow Tools',
      description: 'Use other workflow nodes as tools for this agent',
      icon: '⚡',
      defaultData: {
        enabledTools: [],
        allowWorkflowNodes: true,
      },
    },
  ],
  memory: [
    {
      id: 'local_memory',
      type: 'local_memory',
      label: 'Local Memory',
      description: 'Store conversation context locally within this workflow',
      icon: '💾',
      defaultData: {
        memoryType: 'local',
        maxMessages: 100,
        enableSummarization: true,
      },
    },
  ],
};

export default function ConfigSelectionDialog({
  isOpen,
  configType,
  onClose,
  onSelect,
}) {
  const [selectedOption, setSelectedOption] = useState(null);

  if (!isOpen) return null;

  const options = CONFIG_OPTIONS[configType] || [];

  const handleSelect = () => {
    if (selectedOption) {
      onSelect(selectedOption);
      onClose();
      setSelectedOption(null);
    }
  };

  const handleClose = () => {
    onClose();
    setSelectedOption(null);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-96 max-h-96 overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-bold">
            Select {configType?.charAt(0).toUpperCase() + configType?.slice(1)}{' '}
            Configuration
          </h2>
          <button
            onClick={handleClose}
            className="text-gray-500 hover:text-gray-700"
          >
            ✕
          </button>
        </div>

        <div className="space-y-3">
          {options.map((option) => (
            <div
              key={option.id}
              onClick={() => setSelectedOption(option)}
              className={`p-3 border rounded-lg cursor-pointer hover:bg-gray-50 ${
                selectedOption?.id === option.id
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200'
              }`}
            >
              <div className="flex items-center gap-3">
                <span className="text-2xl">{option.icon}</span>
                <div className="flex-1">
                  <div className="font-medium">{option.label}</div>
                  <div className="text-sm text-gray-500">
                    {option.description}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="flex gap-2 mt-6">
          <button
            onClick={handleClose}
            className="flex-1 px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSelect}
            disabled={!selectedOption}
            className={`flex-1 px-4 py-2 rounded text-white ${
              selectedOption
                ? 'bg-blue-500 hover:bg-blue-600'
                : 'bg-gray-300 cursor-not-allowed'
            }`}
          >
            Add Configuration
          </button>
        </div>
      </div>
    </div>
  );
}
