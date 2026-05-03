export type EditorTool = "select" | "rect" | "circle" | "line" | "text" | "image";

export interface Point {
  x: number;
  y: number;
}

export interface FramePreset {
  label: string;
  width: number;
  height: number;
}

export const FRAME_PRESETS: FramePreset[] = [
  { label: "Full HD", width: 1920, height: 1080 },
  { label: "QHD", width: 2560, height: 1440 },
  { label: "4K UHD", width: 3840, height: 2160 },
  { label: "HD", width: 1280, height: 720 },
  { label: "XGA", width: 1024, height: 768 }
];

export type SaveStatus = "idle" | "saving" | "saved" | "error";

export type ConnectionStatus = "online" | "offline";
