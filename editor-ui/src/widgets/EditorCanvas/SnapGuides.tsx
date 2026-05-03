import { Line, Layer } from "react-konva";

import type { SnapGuide } from "@features/editor";

interface SnapGuidesProps {
  guides: SnapGuide[];
  stroke: string;
}

export const SnapGuides = ({ guides, stroke }: SnapGuidesProps) => {
  if (guides.length === 0) {
    return null;
  }

  return (
    <Layer listening={false}>
      {guides.map((guide, index) => (
        <Line
          key={`${guide.orientation}-${index}`}
          points={guide.points}
          stroke={stroke}
          strokeWidth={1}
          dash={[4, 4]}
          listening={false}
        />
      ))}
    </Layer>
  );
};
