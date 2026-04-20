import type Konva from "konva";
import { useEffect, useRef } from "react";
import { Transformer } from "react-konva";

interface SelectionLayerProps {
  stageRef: React.RefObject<Konva.Stage | null>;
  selectedFigureIds: string[];
}

export const SelectionLayer = ({ stageRef, selectedFigureIds }: SelectionLayerProps) => {
  const transformerRef = useRef<Konva.Transformer | null>(null);

  useEffect(() => {
    const transformer = transformerRef.current;
    const stage = stageRef.current;

    if (!transformer || !stage) {
      return;
    }

    const nodes = selectedFigureIds
      .map((id) => stage.findOne(`#figure-${id}`))
      .filter((node): node is Konva.Node => Boolean(node));

    transformer.nodes(nodes);
    transformer.getLayer()?.batchDraw();
  }, [selectedFigureIds, stageRef]);

  return (
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
    />
  );
};
