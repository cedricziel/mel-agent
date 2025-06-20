import { useState, useEffect } from 'react';
import { credentialsApi } from '../api/client';

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
        const response = await credentialsApi.listCredentials({
          credential_type: param.credentialType,
        });
        return { paramName: param.name, credentials: response.data };
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
