import useDataViewerState from '../hooks/useDataViewerState';
import DataViewerItem from './DataViewerItem';

// Structured data viewer component inspired by n8n
export default function DataViewer({
  data,
  title = 'Data',
  searchable = true,
}) {
  const {
    searchTerm,
    setSearchTerm,
    expandedPaths,
    toggleExpanded,
    shouldShowInSearch,
    copyToClipboard,
  } = useDataViewerState(searchable);

  const getTypeColor = (type) => {
    switch (type) {
      case 'string':
        return 'text-green-600';
      case 'number':
        return 'text-blue-600';
      case 'boolean':
        return 'text-purple-600';
      case 'null':
        return 'text-gray-500';
      case 'undefined':
        return 'text-gray-400';
      default:
        return 'text-gray-800';
    }
  };

  const getTypeIcon = (type) => {
    switch (type) {
      case 'string':
        return '"abc"';
      case 'number':
        return '123';
      case 'boolean':
        return 'T/F';
      case 'array':
        return '[]';
      case 'object':
        return '{}';
      case 'null':
        return 'null';
      default:
        return '?';
    }
  };

  const formatValue = (value) => {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (typeof value === 'string') {
      // Check if it's a date string
      if (value.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)) {
        try {
          const date = new Date(value);
          return `"${value}" (${date.toLocaleString()})`;
        } catch {
          return `"${value}"`;
        }
      }
      return `"${value}"`;
    }
    if (typeof value === 'boolean') return value.toString();
    if (typeof value === 'number') return value.toString();
    return JSON.stringify(value);
  };

  if (!data || (typeof data === 'object' && Object.keys(data).length === 0)) {
    return (
      <div className="p-3 text-center text-gray-500 text-sm">
        No data available
      </div>
    );
  }

  // Handle non-object root data
  if (typeof data !== 'object') {
    return (
      <div className="p-3">
        <div className="flex items-center gap-2 mb-2">
          <h4 className="font-medium text-gray-700">{title}</h4>
          <span className="text-xs bg-gray-100 px-1 rounded">
            {getTypeIcon(typeof data)}
          </span>
        </div>
        <div className="bg-gray-50 p-2 rounded">
          <span className={getTypeColor(typeof data)}>{formatValue(data)}</span>
        </div>
      </div>
    );
  }

  const rootEntries = Object.entries(data);

  return (
    <div className="p-3">
      <div className="flex items-center gap-2 mb-3">
        <h4 className="font-medium text-gray-700">{title}</h4>
        <span className="text-xs bg-gray-100 px-1 rounded">
          {getTypeIcon('object')}
        </span>
        <span className="text-xs text-gray-500">
          ({rootEntries.length} keys)
        </span>
        <button
          onClick={() => copyToClipboard(data)}
          className="ml-auto text-xs text-gray-400 hover:text-gray-600 px-1"
          title="Copy all to clipboard"
        >
          📋 Copy All
        </button>
      </div>

      {searchable && (
        <div className="mb-3">
          <input
            type="text"
            placeholder="Search data..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full text-xs border rounded px-2 py-1 bg-white"
          />
        </div>
      )}

      <div className="space-y-1 max-h-96 overflow-y-auto">
        {rootEntries.length === 0 ? (
          <div className="text-center text-gray-500 text-sm py-4">
            Empty object
          </div>
        ) : (
          rootEntries.map(([key, value]) => (
            <DataViewerItem
              key={key}
              itemKey={key}
              value={value}
              path="root"
              level={0}
              isExpanded={expandedPaths.has(`root.${key}`)}
              toggleExpanded={toggleExpanded}
              copyToClipboard={copyToClipboard}
              getTypeColor={getTypeColor}
              getTypeIcon={getTypeIcon}
              formatValue={formatValue}
              shouldShowInSearch={shouldShowInSearch}
            />
          ))
        )}
      </div>
    </div>
  );
}
