import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const CircleShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawCircle = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity);
    graphics.drawCircle(resolved.x, resolved.y, resolved.radius);
    graphics.endFill();
  };

  return <pixiGraphics draw={drawCircle} />;
};
