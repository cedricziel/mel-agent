// Draft API client for auto-persistence workflow management

import { useState, useEffect, useCallback } from 'react';
import { agentsApi } from './client';

export class DraftAPI {
  static async getDraft(agentId) {
    try {
      const response = await agentsApi.getAgentDraft(agentId);
      return response.data;
    } catch (error) {
      console.error('Error getting draft:', error);
      throw error;
    }
  }

  static async updateDraft(agentId, draftData) {
    try {
      const response = await agentsApi.updateAgentDraft(agentId, draftData);
      return response.data;
    } catch (error) {
      console.error('Error updating draft:', error);
      throw error;
    }
  }

  static async testDraftNode(agentId, nodeId, testData = {}) {
    try {
      const response = await agentsApi.testDraftNode(agentId, nodeId, {
        input: testData,
      });
      return response.data;
    } catch (error) {
      console.error('Error testing draft node:', error);
      throw error;
    }
  }

  static async deployVersion(agentId, versionId) {
    try {
      const response = await agentsApi.deployAgentVersion(agentId, {
        version_id: versionId,
      });
      return response.data;
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
      () => {
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
