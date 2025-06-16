import { useState } from 'react';

/**
 * Custom hook to manage DataViewer state including search and expansion
 * @param {boolean} searchable - Whether search functionality is enabled
 * @returns {Object} State and handlers for DataViewer
 */
export default function useDataViewerState(searchable = true) {
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

  const shouldShowInSearch = (key, value, path) => {
    if (!searchable || !searchTerm) return true;
    const searchLower = searchTerm.toLowerCase();

    // Search in key name
    if (key.toLowerCase().includes(searchLower)) return true;

    // Search in string values
    if (typeof value === 'string' && value.toLowerCase().includes(searchLower))
      return true;

    // Search in path
    if (path.toLowerCase().includes(searchLower)) return true;

    return false;
  };

  const copyToClipboard = (value) => {
    navigator.clipboard.writeText(JSON.stringify(value, null, 2));
  };

  return {
    searchTerm,
    setSearchTerm,
    expandedPaths,
    toggleExpanded,
    shouldShowInSearch,
    copyToClipboard,
    searchable,
  };
}
