import { ArcConfig } from 'konva/lib/shapes/Arc';
import { CircleConfig } from 'konva/lib/shapes/Circle';
import { EllipseConfig } from 'konva/lib/shapes/Ellipse';
import { ImageConfig } from 'konva/lib/shapes/Image';
import { TagConfig } from 'konva/lib/shapes/Label';
import { LineConfig } from 'konva/lib/shapes/Line';
import { PathConfig } from 'konva/lib/shapes/Path';
import { RectConfig } from 'konva/lib/shapes/Rect';
import { RingConfig } from 'konva/lib/shapes/Ring';
import { TextConfig } from 'konva/lib/shapes/Text';
import { WedgeConfig } from 'konva/lib/shapes/Wedge';

type Shape<T, U extends string> = T & { type: U; id: string };

export type ShapeConfig =
  | Shape<RectConfig, 'rect'>
  | Shape<CircleConfig, 'circle'>
  | Shape<EllipseConfig, 'ellipse'>
  | Shape<WedgeConfig, 'wedge'>
  | Shape<LineConfig, 'line'>
  | Shape<ImageConfig, 'image'>
  | Shape<TextConfig, 'text'>
  | Shape<RingConfig, 'ring'>
  | Shape<ArcConfig, 'arc'>
  | Shape<TagConfig, 'tag'>
  | Shape<PathConfig, 'path'>;

export type ExtractedShapeConfig<T extends ShapeConfig['type']> = Extract<
  ShapeConfig,
  { type: T }
>;

export interface shapeStore {
  shapes: ShapeConfig[];
  selectedShapes: number[];
  addShape: <T extends ShapeConfig['type']>(
    shape: ExtractedShapeConfig<T>
  ) => void;
}
