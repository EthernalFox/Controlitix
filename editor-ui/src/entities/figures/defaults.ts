import type { ShapeType } from "@entities/shapes";

export const DEFAULT_FIGURE_PARAMS: Record<ShapeType, Record<string, unknown>> = {
  rect: {
    width: 120,
    height: 80,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  circle: {
    radius: 50,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  line: { points: [0, 0, 120, 0], stroke: "#212529", strokeWidth: 2, visible: true },
  text: {
    text: "Text",
    fontSize: 16,
    fill: "#212529",
    fontFamily: "Inter",
    width: 100,
    visible: true
  },
  image: { width: 160, height: 120, visible: true },
  ellipse: {
    radiusX: 60,
    radiusY: 40,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  wedge: {
    radius: 50,
    angle: 60,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  ring: {
    innerRadius: 30,
    outerRadius: 50,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  arc: {
    innerRadius: 30,
    outerRadius: 50,
    angle: 90,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  tag: {
    pointerDirection: "down",
    pointerWidth: 10,
    pointerHeight: 10,
    cornerRadius: 4,
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  },
  path: {
    data: "M 0 0 L 50 50 L 100 0 Z",
    fill: "#228BE6",
    stroke: "#1971C2",
    strokeWidth: 1,
    visible: true
  }
};
