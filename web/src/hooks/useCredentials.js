import { useState, useEffect } from 'react';

/**
 * Custom hook to manage credential loading for node parameters
 * @param {Object} nodeDef - Node definition containing parameters
 * @returns {Object} Credentials state and loading status
 */
export default function useCredentials(nodeDef) {
  const [credentials, setCredentials] = useState({});
  const [loadingCredentials, setLoadingCredentials] = useState(false);

  useEffect(() => {
    if (!nodeDef || !nodeDef.parameters) {
      setCredentials({});
      return;
    }

    const credentialParams = nodeDef.parameters.filter(
      (p) => p.type === 'credential' || p.parameterType === 'credential'
    );

    if (credentialParams.length === 0) {
      setCredentials({});
      return;
    }

    setLoadingCredentials(true);

    // Load credentials for each credential parameter
    const promises = credentialParams.map(async (param) => {
      try {
        const url = param.credentialType
          ? `/api/credentials?credential_type=${param.credentialType}`
          : '/api/credentials';
        const response = await fetch(url);
        const data = await response.json();
        return { paramName: param.name, credentials: data };
      } catch (error) {
        console.error(`Failed to load credentials for ${param.name}:`, error);
        return { paramName: param.name, credentials: [] };
      }
    });

    Promise.all(promises)
      .then((results) => {
        const credentialsMap = {};
        results.forEach(({ paramName, credentials }) => {
          credentialsMap[paramName] = credentials;
        });
        setCredentials(credentialsMap);
      })
      .finally(() => {
        setLoadingCredentials(false);
      });
  }, [nodeDef]);

  return {
    credentials,
    loadingCredentials,
  };
}
