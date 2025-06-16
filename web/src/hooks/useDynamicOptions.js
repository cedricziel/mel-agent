import { useState, useEffect } from 'react';

/**
 * Custom hook to manage dynamic options loading for node parameters
 * @param {Object} nodeDef - Node definition containing parameters
 * @param {Object} currentFormData - Current form data for building query parameters
 * @returns {Object} Dynamic options state and loading functions
 */
export default function useDynamicOptions(nodeDef, currentFormData) {
  const [dynamicOptions, setDynamicOptions] = useState({});
  const [loadingOptions, setLoadingOptions] = useState({});

  // Function to load dynamic options for a specific parameter
  const loadDynamicOptionsForParam = async (paramName) => {
    const param = nodeDef?.parameters?.find((p) => p.name === paramName);
    if (!param?.dynamicOptions) return;

    try {
      // Set loading state for this specific parameter
      setLoadingOptions((prev) => ({ ...prev, [paramName]: true }));

      // Build query parameters from current form data
      const queryParams = new URLSearchParams();
      Object.entries(currentFormData || {}).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          queryParams.append(key, value);
        }
      });

      const url = `/api/node-types/${nodeDef.type}/parameters/${paramName}/options?${queryParams}`;
      console.log(`Loading dynamic options for ${paramName}:`, url);

      const response = await fetch(url);

      if (response.ok) {
        const data = await response.json();
        console.log(`Dynamic options for ${paramName}:`, data);
        setDynamicOptions((prev) => ({
          ...prev,
          [paramName]: data.options || [],
        }));
      } else {
        console.log(
          `No dynamic options for ${paramName}:`,
          response.status,
          await response.text()
        );
        setDynamicOptions((prev) => ({
          ...prev,
          [paramName]: [],
        }));
      }
    } catch (error) {
      console.log(`Error loading dynamic options for ${paramName}:`, error);
      setDynamicOptions((prev) => ({
        ...prev,
        [paramName]: [],
      }));
    } finally {
      // Clear loading state for this specific parameter
      setLoadingOptions((prev) => ({ ...prev, [paramName]: false }));
    }
  };

  // Load initial dynamic options when credential changes
  useEffect(() => {
    if (!nodeDef || !currentFormData?.credentialId) return;

    // Only load database options initially
    loadDynamicOptionsForParam('databaseId');
  }, [nodeDef, currentFormData?.credentialId]);

  // Load table options when database changes
  useEffect(() => {
    if (!nodeDef || !currentFormData?.databaseId) {
      // Clear table options when no database is selected
      setDynamicOptions((prev) => ({ ...prev, tableId: [] }));
      return;
    }

    loadDynamicOptionsForParam('tableId');
  }, [nodeDef, currentFormData?.databaseId]);

  // Clear all dynamic options when nodeDef changes
  useEffect(() => {
    setDynamicOptions({});
    setLoadingOptions({});
  }, [nodeDef]);

  return {
    dynamicOptions,
    loadingOptions,
    loadDynamicOptionsForParam,
  };
}
