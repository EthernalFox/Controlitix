import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

const toRadians = (degrees: number): number => (degrees * Math.PI) / 180;

export const ArcShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawArc = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    const startAngle = toRadians(resolved.rotation);
    const endAngle = startAngle + toRadians(resolved.angle);

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha * resolved.opacity);
    graphics.arc(resolved.x, resolved.y, resolved.radius, startAngle, endAngle);
  };

  return <pixiGraphics draw={drawArc} />;
};
