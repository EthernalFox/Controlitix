import type { Graphics } from "pixi.js";

import type { ShapeProps } from "./types";

export const TagShape = ({ resolved, alpha, cacheAsBitmap }: ShapeProps) => {
  const drawTag = (graphics: Graphics) => {
    graphics.cacheAsBitmap = cacheAsBitmap;
    graphics.clear();

    if (!resolved.visible) {
      return;
    }

    graphics.lineStyle(resolved.strokeWidth, resolved.strokeColor, alpha);
    graphics.beginFill(resolved.fillColor, alpha * resolved.opacity);
    graphics.drawRoundedRect(resolved.x, resolved.y, resolved.width, resolved.height, 4);
    graphics.endFill();
  };

  if (!resolved.visible) {
    return null;
  }

  return (
    <pixiContainer>
      <pixiGraphics draw={drawTag} />
      <pixiText
        text={resolved.text}
        x={resolved.x + 6}
        y={resolved.y + Math.max(0, (resolved.height - resolved.fontSize) / 2)}
        alpha={alpha * resolved.opacity}
        style={{
          fill: resolved.strokeColor,
          fontSize: resolved.fontSize,
          fontFamily: resolved.fontFamily
        }}
      />
    </pixiContainer>
  );
};
