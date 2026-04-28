import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

const toRadians = (degrees: number): number => (degrees * Math.PI) / 180;

export const WedgeShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawWedge = (graphics: Graphics) => {
    graphics.clear();
    graphics.cacheAsBitmap = cacheAsBitmap;

    if (!resolved.visible) {
      return;
    }

    const startAngle = toRadians(resolved.rotation);
    const endAngle = startAngle + toRadians(resolved.angle);

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity);
    graphics.moveTo(resolved.x, resolved.y);
    graphics.arc(resolved.x, resolved.y, resolved.outerRadius, startAngle, endAngle);
    graphics.lineTo(resolved.x, resolved.y);
    graphics.endFill();
  };

  return <pixiGraphics draw={drawWedge} />;
};
