import { create } from "zustand";

import { FRAME_PRESETS } from "./types";
import type {
  ConnectionStatus,
  EditorTool,
  FramePreset,
  Point,
  SaveStatus
} from "./types";

interface DraggingState {
  ids: string[];
  from: Point;
}

interface EditorState {
  zoom: number;
  panX: number;
  panY: number;
  activeTool: EditorTool;
  selectedFigureIds: string[];
  selection: { ids: string[] };
  dragging: DraggingState | null;
  gridEnabled: boolean;
  snapEnabled: boolean;
  frame: FramePreset;
  saveStatus: SaveStatus;
  connectionStatus: ConnectionStatus;
  cursor: Point | null;
  setZoom: (zoom: number) => void;
  setPan: (x: number, y: number) => void;
  setActiveTool: (tool: EditorTool) => void;
  setSelection: (ids: string[]) => void;
  selectFigure: (id: string, additive?: boolean) => void;
  clearSelection: () => void;
  setDragging: (dragging: DraggingState | null) => void;
  setGridEnabled: (enabled: boolean) => void;
  setSnapEnabled: (enabled: boolean) => void;
  setFrame: (preset: FramePreset) => void;
  setSaveStatus: (status: SaveStatus) => void;
  setConnectionStatus: (status: ConnectionStatus) => void;
  setCursor: (cursor: Point | null) => void;
  reset: () => void;
}

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

const dedupeIds = (ids: string[]) => [...new Set(ids)];

const isOnline = () => {
  if (typeof window === "undefined") {
    return "online" as const;
  }

  return window.navigator.onLine ? "online" : "offline";
};

export const useEditorStore = create<EditorState>((set) => ({
  zoom: 1,
  panX: 0,
  panY: 0,
  activeTool: "select",
  selectedFigureIds: [],
  selection: { ids: [] },
  dragging: null,
  gridEnabled: true,
  snapEnabled: true,
  frame: FRAME_PRESETS[0],
  saveStatus: "idle",
  connectionStatus: isOnline(),
  cursor: null,

  setZoom: (zoom) => set({ zoom: clamp(zoom, 0.1, 5) }),
  setPan: (panX, panY) => set({ panX, panY }),
  setActiveTool: (activeTool) => set({ activeTool }),
  setSelection: (ids) => {
    const nextIds = dedupeIds(ids);
    set({ selectedFigureIds: nextIds, selection: { ids: nextIds } });
  },
  selectFigure: (id, additive = false) =>
    set((state) => {
      if (!additive) {
        return { selectedFigureIds: [id], selection: { ids: [id] } };
      }

      const hasId = state.selection.ids.includes(id);
      const nextIds = hasId
        ? state.selection.ids.filter((selectedId) => selectedId !== id)
        : [...state.selection.ids, id];

      return {
        selectedFigureIds: nextIds,
        selection: { ids: nextIds }
      };
    }),
  clearSelection: () => set({ selectedFigureIds: [], selection: { ids: [] } }),
  setDragging: (dragging) => set({ dragging }),
  setGridEnabled: (gridEnabled) => set({ gridEnabled }),
  setSnapEnabled: (snapEnabled) => set({ snapEnabled }),
  setFrame: (frame) => set({ frame }),
  setSaveStatus: (saveStatus) => set({ saveStatus }),
  setConnectionStatus: (connectionStatus) => set({ connectionStatus }),
  setCursor: (cursor) => set({ cursor }),
  reset: () =>
    set({
      zoom: 1,
      panX: 0,
      panY: 0,
      activeTool: "select",
      selectedFigureIds: [],
      selection: { ids: [] },
      dragging: null,
      gridEnabled: true,
      snapEnabled: true,
      frame: FRAME_PRESETS[0],
      saveStatus: "idle",
      connectionStatus: isOnline(),
      cursor: null
    })
}));
