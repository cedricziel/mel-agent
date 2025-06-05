import { useState } from 'react';

// Structured data viewer component inspired by n8n
export default function DataViewer({ data, title = "Data", searchable = true }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [expandedPaths, setExpandedPaths] = useState(new Set(['root']));

  const toggleExpanded = (path) => {
    const newExpanded = new Set(expandedPaths);
    if (newExpanded.has(path)) {
      newExpanded.delete(path);
    } else {
      newExpanded.add(path);
    }
    setExpandedPaths(newExpanded);
  };

  const copyToClipboard = (value) => {
    navigator.clipboard.writeText(JSON.stringify(value, null, 2));
  };

  const getTypeColor = (type) => {
    switch (type) {
      case 'string': return 'text-green-600';
      case 'number': return 'text-blue-600';
      case 'boolean': return 'text-purple-600';
      case 'null': return 'text-gray-500';
      case 'undefined': return 'text-gray-400';
      default: return 'text-gray-800';
    }
  };

  const getTypeIcon = (type) => {
    switch (type) {
      case 'string': return '"abc"';
      case 'number': return '123';
      case 'boolean': return 'T/F';
      case 'array': return '[]';
      case 'object': return '{}';
      case 'null': return 'null';
      default: return '?';
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

  const shouldShowInSearch = (key, value, path) => {
    if (!searchTerm) return true;
    const searchLower = searchTerm.toLowerCase();
    
    // Search in key name
    if (key.toLowerCase().includes(searchLower)) return true;
    
    // Search in string values
    if (typeof value === 'string' && value.toLowerCase().includes(searchLower)) return true;
    
    // Search in path
    if (path.toLowerCase().includes(searchLower)) return true;
    
    return false;
  };

  const renderArrayTable = (array, path) => {
    if (!Array.isArray(array) || array.length === 0) return null;
    
    // Check if it's an array of objects with consistent keys
    const firstItem = array[0];
    if (typeof firstItem !== 'object' || firstItem === null) return null;
    
    const allKeys = new Set();
    array.forEach(item => {
      if (typeof item === 'object' && item !== null) {
        Object.keys(item).forEach(key => allKeys.add(key));
      }
    });
    
    if (allKeys.size === 0) return null;
    
    return (
      <div className="mt-2 border rounded overflow-hidden">
        <table className="w-full text-xs">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-2 py-1 text-left border-r">#</th>
              {Array.from(allKeys).map(key => (
                <th key={key} className="px-2 py-1 text-left border-r">{key}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {array.map((item, index) => (
              <tr key={index} className="border-t hover:bg-gray-50">
                <td className="px-2 py-1 border-r font-mono text-gray-500">{index}</td>
                {Array.from(allKeys).map(key => (
                  <td key={key} className="px-2 py-1 border-r">
                    <span className={getTypeColor(typeof item[key])}>
                      {formatValue(item[key])}
                    </span>
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const renderValue = (key, value, path = 'root', level = 0) => {
    const currentPath = `${path}.${key}`;
    const isExpanded = expandedPaths.has(currentPath);
    const valueType = Array.isArray(value) ? 'array' : typeof value;
    
    if (!shouldShowInSearch(key, value, currentPath)) {
      return null;
    }

    if (valueType === 'object' && value !== null) {
      const entries = Object.entries(value);
      return (
        <div key={currentPath} className="mb-1">
          <div 
            className="flex items-center gap-2 hover:bg-gray-50 p-1 rounded cursor-pointer"
            style={{ paddingLeft: `${level * 12 + 4}px` }}
            onClick={() => toggleExpanded(currentPath)}
          >
            <span className="text-gray-400 w-3">
              {entries.length > 0 ? (isExpanded ? '▼' : '▶') : '○'}
            </span>
            <span className="font-medium text-gray-700">{key}</span>
            <span className="text-xs bg-gray-100 px-1 rounded">{getTypeIcon(valueType)}</span>
            <span className="text-xs text-gray-500">({entries.length} keys)</span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                copyToClipboard(value);
              }}
              className="ml-auto text-xs text-gray-400 hover:text-gray-600 px-1"
              title="Copy to clipboard"
            >
              📋
            </button>
          </div>
          {isExpanded && (
            <div className="ml-4">
              {entries.map(([subKey, subValue]) => 
                renderValue(subKey, subValue, currentPath, level + 1)
              )}
            </div>
          )}
        </div>
      );
    }

    if (valueType === 'array') {
      const tableView = renderArrayTable(value, currentPath);
      return (
        <div key={currentPath} className="mb-1">
          <div 
            className="flex items-center gap-2 hover:bg-gray-50 p-1 rounded cursor-pointer"
            style={{ paddingLeft: `${level * 12 + 4}px` }}
            onClick={() => toggleExpanded(currentPath)}
          >
            <span className="text-gray-400 w-3">
              {value.length > 0 ? (isExpanded ? '▼' : '▶') : '○'}
            </span>
            <span className="font-medium text-gray-700">{key}</span>
            <span className="text-xs bg-gray-100 px-1 rounded">{getTypeIcon(valueType)}</span>
            <span className="text-xs text-gray-500">({value.length} items)</span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                copyToClipboard(value);
              }}
              className="ml-auto text-xs text-gray-400 hover:text-gray-600 px-1"
              title="Copy to clipboard"
            >
              📋
            </button>
          </div>
          {isExpanded && (
            <div className="ml-4">
              {tableView ? (
                tableView
              ) : (
                value.map((item, index) => 
                  renderValue(`[${index}]`, item, currentPath, level + 1)
                )
              )}
            </div>
          )}
        </div>
      );
    }

    // Primitive values
    return (
      <div key={currentPath} className="mb-1">
        <div 
          className="flex items-center gap-2 hover:bg-gray-50 p-1 rounded"
          style={{ paddingLeft: `${level * 12 + 4}px` }}
        >
          <span className="text-gray-400 w-3">○</span>
          <span className="font-medium text-gray-700">{key}</span>
          <span className="text-xs bg-gray-100 px-1 rounded">{getTypeIcon(valueType)}</span>
          <span className={`flex-1 ${getTypeColor(valueType)}`}>
            {formatValue(value)}
          </span>
          <button
            onClick={() => copyToClipboard(value)}
            className="text-xs text-gray-400 hover:text-gray-600 px-1"
            title="Copy to clipboard"
          >
            📋
          </button>
        </div>
      </div>
    );
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
          <span className="text-xs bg-gray-100 px-1 rounded">{getTypeIcon(typeof data)}</span>
        </div>
        <div className="bg-gray-50 p-2 rounded">
          <span className={getTypeColor(typeof data)}>
            {formatValue(data)}
          </span>
        </div>
      </div>
    );
  }

  const rootEntries = Object.entries(data);

  return (
    <div className="p-3">
      <div className="flex items-center gap-2 mb-3">
        <h4 className="font-medium text-gray-700">{title}</h4>
        <span className="text-xs bg-gray-100 px-1 rounded">{getTypeIcon('object')}</span>
        <span className="text-xs text-gray-500">({rootEntries.length} keys)</span>
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
          rootEntries.map(([key, value]) => renderValue(key, value, 'root', 0))
        )}
      </div>
    </div>
  );
}