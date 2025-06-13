// Draft API client for auto-persistence workflow management

import { useState, useEffect, useCallback } from 'react';

const API_BASE = '/api';

export class DraftAPI {
  static async getDraft(agentId) {
    try {
      const response = await fetch(`${API_BASE}/agents/${agentId}/draft`);
      if (!response.ok) {
        throw new Error(`Failed to get draft: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error getting draft:', error);
      throw error;
    }
  }

  static async updateDraft(agentId, draftData) {
    try {
      const response = await fetch(`${API_BASE}/agents/${agentId}/draft`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(draftData),
      });

      if (!response.ok) {
        throw new Error(`Failed to update draft: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error updating draft:', error);
      throw error;
    }
  }

  static async testDraftNode(agentId, nodeId, testData = {}) {
    try {
      const response = await fetch(
        `${API_BASE}/agents/${agentId}/draft/nodes/${nodeId}/test`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            node_id: nodeId,
            test_data: testData,
          }),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to test node: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error testing draft node:', error);
      throw error;
    }
  }

  static async deployVersion(agentId, version, notes = '') {
    try {
      const response = await fetch(`${API_BASE}/agents/${agentId}/deploy`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          version: version,
          notes: notes,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to deploy version: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error deploying version:', error);
      throw error;
    }
  }
}

// Auto-save utilities
export class AutoSaver {
  constructor(agentId, onSave = null, onError = null) {
    this.agentId = agentId;
    this.onSave = onSave;
    this.onError = onError;
    this.saveTimeout = null;
    this.isSaving = false;
    this.pendingChanges = null;
    this.saveDelay = 1000; // 1 second delay for auto-save
  }

  // Schedule an auto-save with debouncing
  scheduleSave(draftData) {
    this.pendingChanges = draftData;

    // Clear existing timeout
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
    }

    // Schedule new save
    this.saveTimeout = setTimeout(() => {
      this.performSave();
    }, this.saveDelay);
  }

  // Perform the actual save
  async performSave() {
    if (this.isSaving || !this.pendingChanges) {
      return;
    }

    this.isSaving = true;

    try {
      const result = await DraftAPI.updateDraft(
        this.agentId,
        this.pendingChanges
      );
      this.pendingChanges = null;

      if (this.onSave) {
        this.onSave(result);
      }

      console.log('✅ Draft auto-saved at', new Date().toLocaleTimeString());
    } catch (error) {
      console.error('❌ Auto-save failed:', error);

      if (this.onError) {
        this.onError(error);
      }
    } finally {
      this.isSaving = false;
    }
  }

  // Force immediate save
  async saveNow() {
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
      this.saveTimeout = null;
    }

    await this.performSave();
  }

  // Cleanup
  destroy() {
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
    }
  }
}

// Hook for React components
export function useAutoSaver(agentId) {
  const [autoSaver, setAutoSaver] = useState(null);
  const [isSaving, setIsSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);
  const [saveError, setSaveError] = useState(null);

  useEffect(() => {
    if (!agentId) return;

    const saver = new AutoSaver(
      agentId,
      (result) => {
        setIsSaving(false);
        setLastSaved(new Date());
        setSaveError(null);
      },
      (error) => {
        setIsSaving(false);
        setSaveError(error);
      }
    );

    setAutoSaver(saver);

    return () => {
      saver.destroy();
    };
  }, [agentId]);

  const scheduleSave = useCallback(
    (draftData) => {
      if (autoSaver) {
        setIsSaving(true);
        setSaveError(null);
        autoSaver.scheduleSave(draftData);
      }
    },
    [autoSaver]
  );

  const saveNow = useCallback(async () => {
    if (autoSaver) {
      setIsSaving(true);
      setSaveError(null);
      await autoSaver.saveNow();
    }
  }, [autoSaver]);

  return {
    scheduleSave,
    saveNow,
    isSaving,
    lastSaved,
    saveError,
  };
}
