import { create } from "zustand";

import { FRAME_PRESETS } from "./types";
import type { EditorTool, FramePreset, SaveStatus } from "./types";

interface EditorState {
  zoom: number;
  panX: number;
  panY: number;
  activeTool: EditorTool;
  selectedFigureIds: string[];
  frame: FramePreset;
  saveStatus: SaveStatus;
  cursorX: number;
  cursorY: number;
  setZoom: (zoom: number) => void;
  setPan: (x: number, y: number) => void;
  setActiveTool: (tool: EditorTool) => void;
  selectFigure: (id: string) => void;
  clearSelection: () => void;
  setFrame: (preset: FramePreset) => void;
  setSaveStatus: (status: SaveStatus) => void;
  setCursor: (x: number, y: number) => void;
  reset: () => void;
}

export const useEditorStore = create<EditorState>((set) => ({
  zoom: 1,
  panX: 0,
  panY: 0,
  activeTool: "select",
  selectedFigureIds: [],
  frame: FRAME_PRESETS[0],
  saveStatus: "saved",
  cursorX: 0,
  cursorY: 0,

  setZoom: (zoom) => set({ zoom: Math.min(5, Math.max(0.1, zoom)) }),
  setPan: (panX, panY) => set({ panX, panY }),
  setActiveTool: (activeTool) => set({ activeTool }),
  selectFigure: (id) => set({ selectedFigureIds: [id] }),
  clearSelection: () => set({ selectedFigureIds: [] }),
  setFrame: (frame) => set({ frame }),
  setSaveStatus: (saveStatus) => set({ saveStatus }),
  setCursor: (cursorX, cursorY) => set({ cursorX, cursorY }),
  reset: () =>
    set({
      zoom: 1,
      panX: 0,
      panY: 0,
      activeTool: "select",
      selectedFigureIds: [],
      frame: FRAME_PRESETS[0],
      saveStatus: "saved",
      cursorX: 0,
      cursorY: 0
    })
}));
