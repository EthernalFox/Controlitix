import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const LineShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawLine = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible || resolved.points.length < 2) {
      return;
    }

    const [firstPoint, ...restPoints] = resolved.points;

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha * resolved.opacity);
    graphics.moveTo(firstPoint.x, firstPoint.y);

    restPoints.forEach((point) => {
      graphics.lineTo(point.x, point.y);
    });
  };

  return <pixiGraphics draw={drawLine} />;
};
