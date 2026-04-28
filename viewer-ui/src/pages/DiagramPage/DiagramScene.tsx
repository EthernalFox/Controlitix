import type { Graphics } from "pixi.js";

import type { DiagramDetails, DiagramSnapshotPoint } from "@/features/diagram";
import { hexToPixiColor } from "@/shared/libs/pixi/colors";
import { FigureRenderer } from "@/widgets/FigureRenderer";

interface DiagramSceneProps {
  diagram: DiagramDetails;
  viewport: {
    x: number;
    y: number;
    scale: number;
  };
  valuesByTag: Record<string, DiagramSnapshotPoint | undefined>;
}

export const DiagramScene = ({
  diagram,
  viewport,
  valuesByTag
}: DiagramSceneProps) => {
  const drawCanvasBackground = (graphics: Graphics) => {
    graphics.clear();
    graphics.beginFill(hexToPixiColor(diagram.canvas.background));
    graphics.drawRect(0, 0, diagram.canvas.width, diagram.canvas.height);
    graphics.endFill();
  };

  const enableStaticBitmapCache = diagram.figures.length >= 200;

  return (
    <pixiContainer>
      <pixiContainer
        x={viewport.x}
        y={viewport.y}
        scale={{ x: viewport.scale, y: viewport.scale }}
      >
        <pixiGraphics draw={drawCanvasBackground} />

        {diagram.figures.map((figure) => {
          const value = figure.tagId ? valuesByTag[figure.tagId] : undefined;

          return (
            <FigureRenderer
              key={figure.id}
              figure={figure}
              value={value}
              enableStaticBitmapCache={enableStaticBitmapCache}
            />
          );
        })}
      </pixiContainer>
    </pixiContainer>
  );
};
