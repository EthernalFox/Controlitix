import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const EllipseShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawEllipse = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity);
    graphics.drawEllipse(resolved.x, resolved.y, resolved.radiusX, resolved.radiusY);
    graphics.endFill();
  };

  return <pixiGraphics draw={drawEllipse} />;
};
