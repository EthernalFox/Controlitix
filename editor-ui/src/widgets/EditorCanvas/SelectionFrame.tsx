import type Konva from "konva";
import { useEffect, useMemo, useRef, useState } from "react";
import { Rect, Transformer } from "react-konva";

interface SelectionFrameProps {
  stageRef: React.RefObject<Konva.Stage | null>;
  selectedFigureIds: string[];
  borderColor: string;
  multiFill: string;
  version: string;
}

interface Bounds {
  x: number;
  y: number;
  width: number;
  height: number;
}

const getBounds = (nodes: Konva.Node[]): Bounds | null => {
  if (nodes.length === 0) {
    return null;
  }

  const rects = nodes.map((node) => node.getClientRect({ skipShadow: true }));
  const minX = Math.min(...rects.map((rect) => rect.x));
  const minY = Math.min(...rects.map((rect) => rect.y));
  const maxX = Math.max(...rects.map((rect) => rect.x + rect.width));
  const maxY = Math.max(...rects.map((rect) => rect.y + rect.height));

  return {
    x: minX,
    y: minY,
    width: maxX - minX,
    height: maxY - minY
  };
};

export const SelectionFrame = ({
  stageRef,
  selectedFigureIds,
  borderColor,
  multiFill,
  version
}: SelectionFrameProps) => {
  const transformerRef = useRef<Konva.Transformer | null>(null);
  const [multiBounds, setMultiBounds] = useState<Bounds | null>(null);

  const nodes = useMemo(() => {
    const stage = stageRef.current;
    if (!stage) {
      return [];
    }

    return selectedFigureIds
      .map((id) => stage.findOne(`#figure-${id}`))
      .filter((node): node is Konva.Node => Boolean(node));
  }, [selectedFigureIds, stageRef, version]);

  useEffect(() => {
    const transformer = transformerRef.current;

    if (!transformer) {
      return;
    }

    if (nodes.length === 1) {
      transformer.nodes(nodes);
      setMultiBounds(null);
      transformer.getLayer()?.batchDraw();
      return;
    }

    transformer.nodes([]);
    setMultiBounds(nodes.length >= 2 ? getBounds(nodes) : null);
    transformer.getLayer()?.batchDraw();
  }, [nodes]);

  return (
    <>
      {multiBounds && (
        <Rect
          x={multiBounds.x}
          y={multiBounds.y}
          width={multiBounds.width}
          height={multiBounds.height}
          stroke={borderColor}
          strokeWidth={1.5}
          dash={[4, 3]}
          fill={multiFill}
          listening={false}
        />
      )}

      <Transformer
        ref={transformerRef}
        rotateEnabled
        enabledAnchors={[
          "top-left",
          "top-center",
          "top-right",
          "middle-right",
          "bottom-right",
          "bottom-center",
          "bottom-left",
          "middle-left"
        ]}
        borderEnabled
        borderStroke={borderColor}
        borderStrokeWidth={2}
        borderDash={[6, 4]}
        anchorSize={8}
        anchorStroke={borderColor}
        anchorStrokeWidth={1.5}
        anchorFill="#FFFFFF"
        anchorCornerRadius={1}
        rotateAnchorOffset={24}
      />
    </>
  );
};
