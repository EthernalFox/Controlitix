import { Rect, Text } from "react-konva";

import type { FramePreset } from "../model/types";

interface FrameGuideProps {
  frame: FramePreset;
}

export const FrameGuide = ({ frame }: FrameGuideProps) => {
  return (
    <>
      <Rect x={0} y={0} width={frame.width} height={frame.height} fill="#FFFFFF" />
      <Rect
        x={0}
        y={0}
        width={frame.width}
        height={frame.height}
        stroke="#DEE2E6"
        dash={[8, 4]}
        strokeWidth={1}
      />
      <Text
        x={0}
        y={-24}
        text={`${frame.width} × ${frame.height}`}
        fontSize={12}
        fill="#868E96"
      />
    </>
  );
};
