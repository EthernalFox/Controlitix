export { useEditorStore } from "./model/editorStore";

export { FRAME_PRESETS } from "./model/types";
export type {
  ConnectionStatus,
  EditorTool,
  FramePreset,
  Point,
  SaveStatus
} from "./model/types";

export { fitToFrame, screenToCanvas, zoomAtPoint } from "./lib/canvasUtils";
export { useEditorShortcuts } from "./lib/keyboardShortcuts";
export { computeSnapTargets } from "./lib/snap";
export type { SnapGuide, SnapRect } from "./lib/snap";
export { useThemeColor } from "./lib/useThemeColor";

export { EditorCanvas } from "./ui/EditorCanvas";
