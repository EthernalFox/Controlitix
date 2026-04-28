import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const RectShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawRect = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity);
    graphics.drawRoundedRect(resolved.x, resolved.y, resolved.width, resolved.height, 4);
    graphics.endFill();
  };

  return <pixiGraphics draw={drawRect} />;
};
