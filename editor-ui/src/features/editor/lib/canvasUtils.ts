import type { FramePreset } from "../model/types";

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

export function fitToFrame(
  containerWidth: number,
  containerHeight: number,
  frame: FramePreset,
  padding: number = 40
): { zoom: number; panX: number; panY: number } {
  const safeContainerWidth = Math.max(1, containerWidth - padding * 2);
  const safeContainerHeight = Math.max(1, containerHeight - padding * 2);
  const zoom = clamp(
    Math.min(safeContainerWidth / frame.width, safeContainerHeight / frame.height),
    0.1,
    5
  );

  return {
    zoom,
    panX: (containerWidth - frame.width * zoom) / 2,
    panY: (containerHeight - frame.height * zoom) / 2
  };
}

export function zoomAtPoint(
  currentZoom: number,
  currentPan: { x: number; y: number },
  pointerPos: { x: number; y: number },
  delta: number,
  minZoom: number = 0.1,
  maxZoom: number = 5
): { zoom: number; panX: number; panY: number } {
  const scaleBy = 1.05;
  const nextZoom = clamp(
    delta > 0 ? currentZoom / scaleBy : currentZoom * scaleBy,
    minZoom,
    maxZoom
  );

  const canvasPoint = {
    x: (pointerPos.x - currentPan.x) / currentZoom,
    y: (pointerPos.y - currentPan.y) / currentZoom
  };

  return {
    zoom: nextZoom,
    panX: pointerPos.x - canvasPoint.x * nextZoom,
    panY: pointerPos.y - canvasPoint.y * nextZoom
  };
}

export function screenToCanvas(
  screenX: number,
  screenY: number,
  zoom: number,
  pan: { x: number; y: number }
): { x: number; y: number } {
  return {
    x: (screenX - pan.x) / zoom,
    y: (screenY - pan.y) / zoom
  };
}
