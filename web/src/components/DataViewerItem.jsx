import DataViewerTable from './DataViewerTable';

/**
 * Component to render individual data items with proper type handling
 * @param {string} itemKey - The key name for this item
 * @param {*} value - The value to render
 * @param {string} path - Current path in the data structure
 * @param {number} level - Nesting level for indentation
 * @param {boolean} isExpanded - Whether this item is expanded
 * @param {Function} toggleExpanded - Function to toggle expansion
 * @param {Function} copyToClipboard - Function to copy value to clipboard
 * @param {Function} getTypeColor - Function to get color for value types
 * @param {Function} getTypeIcon - Function to get icon for value types
 * @param {Function} formatValue - Function to format values for display
 * @param {Function} shouldShowInSearch - Function to check if item should be visible in search
 * @returns {JSX.Element|null} Rendered item or null if filtered out
 */
export default function DataViewerItem({
  itemKey,
  value,
  path = 'root',
  level = 0,
  isExpanded,
  toggleExpanded,
  copyToClipboard,
  getTypeColor,
  getTypeIcon,
  formatValue,
  shouldShowInSearch,
}) {
  const currentPath = `${path}.${itemKey}`;
  const valueType = Array.isArray(value) ? 'array' : typeof value;

  if (!shouldShowInSearch(itemKey, value, currentPath)) {
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
          <span className="font-medium text-gray-700">{itemKey}</span>
          <span className="text-xs bg-gray-100 px-1 rounded">
            {getTypeIcon(valueType)}
          </span>
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
            {entries.map(([subKey, subValue]) => (
              <DataViewerItem
                key={`${currentPath}.${subKey}`}
                itemKey={subKey}
                value={subValue}
                path={currentPath}
                level={level + 1}
                isExpanded={isExpanded}
                toggleExpanded={toggleExpanded}
                copyToClipboard={copyToClipboard}
                getTypeColor={getTypeColor}
                getTypeIcon={getTypeIcon}
                formatValue={formatValue}
                shouldShowInSearch={shouldShowInSearch}
              />
            ))}
          </div>
        )}
      </div>
    );
  }

  if (valueType === 'array') {
    const tableView = (
      <DataViewerTable
        array={value}
        path={currentPath}
        getTypeColor={getTypeColor}
        formatValue={formatValue}
      />
    );

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
          <span className="font-medium text-gray-700">{itemKey}</span>
          <span className="text-xs bg-gray-100 px-1 rounded">
            {getTypeIcon(valueType)}
          </span>
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
            {tableView
              ? tableView
              : value.map((item, index) => (
                  <DataViewerItem
                    key={`${currentPath}[${index}]`}
                    itemKey={`[${index}]`}
                    value={item}
                    path={currentPath}
                    level={level + 1}
                    isExpanded={isExpanded}
                    toggleExpanded={toggleExpanded}
                    copyToClipboard={copyToClipboard}
                    getTypeColor={getTypeColor}
                    getTypeIcon={getTypeIcon}
                    formatValue={formatValue}
                    shouldShowInSearch={shouldShowInSearch}
                  />
                ))}
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
        <span className="font-medium text-gray-700">{itemKey}</span>
        <span className="text-xs bg-gray-100 px-1 rounded">
          {getTypeIcon(valueType)}
        </span>
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
}
