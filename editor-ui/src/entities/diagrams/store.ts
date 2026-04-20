import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { diagramsApi } from "./api";
import type { CreateDiagramPayload, Diagram } from "./types";

interface DiagramsState {
  diagrams: Diagram[];
  isLoading: boolean;
  error: string | null;
  fetchDiagrams: (objectId: string) => Promise<void>;
  createDiagram: (objectId: string, payload: CreateDiagramPayload) => Promise<Diagram>;
  deleteDiagram: (id: string) => Promise<void>;
  reset: () => void;
}

const toErrorMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Request failed";
};

export const useDiagramsStore = create<DiagramsState>((set) => ({
  diagrams: [],
  isLoading: false,
  error: null,

  fetchDiagrams: async (objectId) => {
    set({ isLoading: true, error: null });

    try {
      const response = await diagramsApi.listByObject(objectId);
      set({ diagrams: response.items, isLoading: false });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  createDiagram: async (objectId, payload) => {
    set({ error: null });
    const diagram = await diagramsApi.create(objectId, payload);

    set((state) => ({
      diagrams: [diagram, ...state.diagrams]
    }));

    return diagram;
  },

  deleteDiagram: async (id) => {
    set({ error: null });
    await diagramsApi.delete(id);

    set((state) => ({
      diagrams: state.diagrams.filter((diagram) => diagram.id !== id)
    }));
  },

  reset: () => {
    set({
      diagrams: [],
      isLoading: false,
      error: null
    });
  }
}));

