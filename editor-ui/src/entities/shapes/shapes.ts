import type { ShapeConfig } from "./types";

export type ShapeType = ShapeConfig["type"];

export interface ShapeDefinition {
  type: ShapeType;
  label: string;
}

export const shapes: ShapeDefinition[] = [
  { type: "rect", label: "Rectangle" },
  { type: "circle", label: "Circle" },
  { type: "ellipse", label: "Ellipse" },
  { type: "wedge", label: "Wedge" },
  { type: "line", label: "Line" },
  { type: "image", label: "Image" },
  { type: "text", label: "Text" },
  { type: "ring", label: "Ring" },
  { type: "arc", label: "Arc" },
  { type: "tag", label: "Tag" },
  { type: "path", label: "Path" }
];
