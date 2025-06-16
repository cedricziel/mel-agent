/**
 * Component to render arrays as tables with consistent object keys
 * @param {Array} array - Array of objects to render as table
 * @param {string} path - Current path for debugging
 * @param {Function} getTypeColor - Function to get color for value types
 * @param {Function} formatValue - Function to format values for display
 * @returns {JSX.Element|null} Table component or null if not applicable
 */
export default function DataViewerTable({
  array,
  path,
  getTypeColor,
  formatValue,
}) {
  if (!Array.isArray(array) || array.length === 0) return null;

  // Check if it's an array of objects with consistent keys
  const firstItem = array[0];
  if (typeof firstItem !== 'object' || firstItem === null) return null;

  const allKeys = new Set();
  array.forEach((item) => {
    if (typeof item === 'object' && item !== null) {
      Object.keys(item).forEach((key) => allKeys.add(key));
    }
  });

  if (allKeys.size === 0) return null;

  return (
    <div className="mt-2 border rounded overflow-hidden">
      <table className="w-full text-xs">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-2 py-1 text-left border-r">#</th>
            {Array.from(allKeys).map((key) => (
              <th key={key} className="px-2 py-1 text-left border-r">
                {key}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {array.map((item, index) => (
            <tr key={index} className="border-t hover:bg-gray-50">
              <td className="px-2 py-1 border-r font-mono text-gray-500">
                {index}
              </td>
              {Array.from(allKeys).map((key) => (
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
}
