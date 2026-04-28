import type { ShapeProps } from "./types";

export const TextShape = ({ resolved, alpha }: ShapeProps) => {
  if (!resolved.visible) {
    return null;
  }

  return (
    <pixiText
      text={resolved.text}
      x={resolved.x}
      y={resolved.y}
      alpha={alpha * resolved.opacity}
      style={{
        fill: resolved.fillColor,
        fontSize: resolved.fontSize,
        fontFamily: resolved.fontFamily
      }}
    />
  );
};
