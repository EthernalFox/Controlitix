import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const PathShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawPath = (graphics: Graphics) => {
    graphics.cacheAsBitmap = cacheAsBitmap;
    graphics.clear();

    if (!resolved.visible || resolved.points.length < 2) {
      return;
    }

    const [firstPoint, ...restPoints] = resolved.points;

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha * resolved.opacity);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity * 0.15);
    graphics.moveTo(firstPoint.x, firstPoint.y);

    restPoints.forEach((point) => {
      graphics.lineTo(point.x, point.y);
    });

    graphics.endFill();
  };

  return <pixiGraphics draw={drawPath} />;
};
