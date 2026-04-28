import type { ImageShapeProps } from "./types";

export const ImageShape = ({ resolved, alpha, texture }: ImageShapeProps) => {
  if (!resolved.visible || !texture) {
    return null;
  }

  return (
    <pixiSprite
      texture={texture}
      x={resolved.x}
      y={resolved.y}
      width={resolved.width}
      height={resolved.height}
      alpha={alpha * resolved.opacity}
    />
  );
};
