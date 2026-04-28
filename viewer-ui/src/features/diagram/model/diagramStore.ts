import { create } from "zustand";

import type {
  DiagramDetails,
  DiagramListItem,
  DiagramObjectItem,
  DiagramSnapshotPoint
} from "@/features/diagram/model/diagramApi";

export interface DiagramViewportState {
  x: number;
  y: number;
  scale: number;
}

interface DiagramState {
  objects: DiagramObjectItem[];
  diagramsByObject: Record<string, DiagramListItem[] | undefined>;
  diagram: DiagramDetails | null;
  valuesByTag: Record<string, DiagramSnapshotPoint | undefined>;
  missingTagIds: string[];
  snapshotTs: string | null;
  viewport: DiagramViewportState;
  isLoading: boolean;
  error: string | null;

  setObjects: (objects: DiagramObjectItem[]) => void;
  setObjectDiagrams: (objectId: string, diagrams: DiagramListItem[]) => void;
  setDiagram: (diagram: DiagramDetails | null) => void;
  setSnapshot: (snapshotTs: string, values: DiagramSnapshotPoint[], missingTagIds: string[]) => void;
  setValue: (point: DiagramSnapshotPoint) => void;
  setViewport: (viewport: DiagramViewportState) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  resetDiagramState: () => void;
}

const defaultViewport: DiagramViewportState = {
  x: 0,
  y: 0,
  scale: 1
};

export const useDiagramStore = create<DiagramState>((set) => ({
  objects: [],
  diagramsByObject: {},
  diagram: null,
  valuesByTag: {},
  missingTagIds: [],
  snapshotTs: null,
  viewport: defaultViewport,
  isLoading: false,
  error: null,

  setObjects: (objects) => {
    set({ objects });
  },

  setObjectDiagrams: (objectId, diagrams) => {
    set((state) => ({
      diagramsByObject: {
        ...state.diagramsByObject,
        [objectId]: diagrams
      }
    }));
  },

  setDiagram: (diagram) => {
    set({
      diagram,
      valuesByTag: {},
      missingTagIds: [],
      snapshotTs: null,
      viewport: defaultViewport,
      error: null
    });
  },

  setSnapshot: (snapshotTs, values, missingTagIds) => {
    const valuesByTag = values.reduce<Record<string, DiagramSnapshotPoint>>((acc, point) => {
      acc[point.tagId] = point;
      return acc;
    }, {});

    set({
      snapshotTs,
      valuesByTag,
      missingTagIds
    });
  },

  setValue: (point) => {
    set((state) => ({
      valuesByTag: {
        ...state.valuesByTag,
        [point.tagId]: point
      }
    }));
  },

  setViewport: (viewport) => {
    set({ viewport });
  },

  setLoading: (isLoading) => {
    set({ isLoading });
  },

  setError: (error) => {
    set({ error });
  },

  resetDiagramState: () => {
    set({
      diagram: null,
      valuesByTag: {},
      missingTagIds: [],
      snapshotTs: null,
      viewport: defaultViewport,
      isLoading: false,
      error: null
    });
  }
}));
