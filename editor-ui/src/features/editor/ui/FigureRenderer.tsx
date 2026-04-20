import type Konva from "konva";
import type { KonvaEventObject } from "konva/lib/Node";

import type { Figure } from "@entities/figures";

import { ShapeRenderer } from "./ShapeRenderer";

interface FigureRendererProps {
  figure: Figure;
  isSelected: boolean;
  onSelect: (id: string) => void;
  onDragEnd: (id: string, x: number, y: number) => void;
  onTransformEnd: (id: string, params: Record<string, unknown>) => void;
}

const toFiniteNumber = (value: unknown, fallback: number) => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return fallback;
};

const buildTransformedParams = (
  figure: Figure,
  node: Konva.Node
): Record<string, unknown> => {
  const currentParams = figure.params;
  const nodeWithSize = node as unknown as {
    width: () => number;
    height: () => number;
    points?: () => number[];
  };

  const scaleX = node.scaleX();
  const scaleY = node.scaleY();
  const nextParams: Record<string, unknown> = {
    ...currentParams,
    x: node.x(),
    y: node.y(),
    rotation: node.rotation()
  };

  if ("width" in currentParams) {
    nextParams.width = Math.max(1, nodeWithSize.width() * scaleX);
  }

  if ("height" in currentParams) {
    nextParams.height = Math.max(1, nodeWithSize.height() * scaleY);
  }

  if ("radius" in currentParams) {
    nextParams.radius = Math.max(
      1,
      toFiniteNumber(currentParams.radius, 0) * Math.max(Math.abs(scaleX), Math.abs(scaleY))
    );
  }

  if ("radiusX" in currentParams) {
    nextParams.radiusX = Math.max(1, toFiniteNumber(currentParams.radiusX, 0) * Math.abs(scaleX));
  }

  if ("radiusY" in currentParams) {
    nextParams.radiusY = Math.max(1, toFiniteNumber(currentParams.radiusY, 0) * Math.abs(scaleY));
  }

  if ("innerRadius" in currentParams) {
    nextParams.innerRadius = Math.max(
      1,
      toFiniteNumber(currentParams.innerRadius, 0) * Math.max(Math.abs(scaleX), Math.abs(scaleY))
    );
  }

  if ("outerRadius" in currentParams) {
    nextParams.outerRadius = Math.max(
      1,
      toFiniteNumber(currentParams.outerRadius, 0) * Math.max(Math.abs(scaleX), Math.abs(scaleY))
    );
  }

  if (Array.isArray(currentParams.points)) {
    const rawPoints = currentParams.points as number[];
    nextParams.points = rawPoints.map((point, index) =>
      index % 2 === 0 ? point * scaleX : point * scaleY
    );
  }

  return nextParams;
};

export const FigureRenderer = ({
  figure,
  isSelected,
  onSelect,
  onDragEnd,
  onTransformEnd
}: FigureRendererProps) => {
  const shapeParams = figure.params ?? {};
  const visible = shapeParams.visible !== false;

  return (
    <ShapeRenderer
      id={`figure-${figure.id}`}
      type={figure.type}
      {...(shapeParams as Record<string, unknown>)}
      visible={visible}
      draggable
      listening
      onClick={() => onSelect(figure.id)}
      onTap={() => onSelect(figure.id)}
      onDragEnd={(event: KonvaEventObject<DragEvent>) =>
        onDragEnd(figure.id, event.target.x(), event.target.y())
      }
      onTransformEnd={(event: KonvaEventObject<Event>) => {
        const node = event.target;
        const nextParams = buildTransformedParams(figure, node);

        node.scaleX(1);
        node.scaleY(1);

        onTransformEnd(figure.id, nextParams);
      }}
      shadowColor={isSelected ? "rgba(34, 139, 230, 0.45)" : undefined}
      shadowBlur={isSelected ? 8 : 0}
      shadowOpacity={isSelected ? 0.6 : 0}
    />
  );
};
