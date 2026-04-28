import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const RingShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawRing = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    const ringWidth = Math.max(1, resolved.outerRadius - resolved.innerRadius);
    const ringRadius = resolved.innerRadius + ringWidth / 2;

    graphics.lineStyle(ringWidth, resolved.fillColor, alpha * resolved.opacity);
    graphics.drawCircle(resolved.x, resolved.y, ringRadius);

    if (resolved.strokeWidth > 0) {
      graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha * resolved.opacity);
      graphics.drawCircle(resolved.x, resolved.y, resolved.outerRadius);
      if (resolved.innerRadius > 0) {
        graphics.drawCircle(resolved.x, resolved.y, resolved.innerRadius);
      }
    }
  };

  return <pixiGraphics draw={drawRing} />;
};
