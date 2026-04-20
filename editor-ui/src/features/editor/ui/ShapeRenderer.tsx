import type { ArcConfig } from "konva/lib/shapes/Arc";
import type { EllipseConfig } from "konva/lib/shapes/Ellipse";
import type { ImageConfig } from "konva/lib/shapes/Image";
import type { RingConfig } from "konva/lib/shapes/Ring";
import type { WedgeConfig } from "konva/lib/shapes/Wedge";
import {
  Arc,
  Circle,
  Ellipse,
  Image as KonvaImage,
  Line,
  Path,
  Rect,
  Ring,
  Tag,
  Text,
  Wedge
} from "react-konva";

import type { ShapeType } from "@entities/shapes";

interface ShapeRendererProps {
  type: ShapeType;
  [key: string]: unknown;
}

export const ShapeRenderer = ({ type, ...config }: ShapeRendererProps) => {
  switch (type) {
    case "rect":
      return <Rect {...config} />;
    case "circle":
      return <Circle {...config} />;
    case "ellipse":
      return <Ellipse {...(config as EllipseConfig)} />;
    case "wedge":
      return <Wedge {...(config as WedgeConfig)} />;
    case "line":
      return <Line {...config} />;
    case "image":
      return <KonvaImage {...(config as ImageConfig)} />;
    case "text":
      return <Text {...config} />;
    case "ring":
      return <Ring {...(config as RingConfig)} />;
    case "arc":
      return <Arc {...(config as ArcConfig)} />;
    case "tag":
      return <Tag {...config} />;
    case "path":
      return <Path {...config} />;
    default:
      return null;
  }
};
