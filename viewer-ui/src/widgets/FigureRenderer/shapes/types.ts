import type { Texture } from "pixi.js";

import type { ResolvedFigureProps } from "@/widgets/FigureRenderer/lib/applyDynamics";

export interface ShapeProps {
  resolved: ResolvedFigureProps;
  alpha: number;
  cacheAsBitmap: boolean;
}

export interface ImageShapeProps extends ShapeProps {
  texture: Texture | null;
}
