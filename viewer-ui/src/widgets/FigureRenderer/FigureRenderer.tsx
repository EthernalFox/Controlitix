import { useTick } from "@pixi/react";
import { useEffect, useMemo, useRef, useState } from "react";

import type {
  DiagramFigure,
  DiagramSnapshotPoint,
  DiagramTagMeta
} from "@/features/diagram";
import { loadTexture } from "@/shared/libs/pixi/setup";

import { applyDynamics } from "./lib/applyDynamics";
import { parseParams } from "./lib/parseParams";
import { ArcShape } from "./shapes/ArcShape";
import { CircleShape } from "./shapes/CircleShape";
import { EllipseShape } from "./shapes/EllipseShape";
import { ImageShape } from "./shapes/ImageShape";
import { LineShape } from "./shapes/LineShape";
import { PathShape } from "./shapes/PathShape";
import { RectShape } from "./shapes/RectShape";
import { RingShape } from "./shapes/RingShape";
import { TagShape } from "./shapes/TagShape";
import { TextShape } from "./shapes/TextShape";
import { WedgeShape } from "./shapes/WedgeShape";

interface FigureRendererProps {
  figure: DiagramFigure;
  value: DiagramSnapshotPoint | undefined;
  enableStaticBitmapCache: boolean;
}

const BLINK_HALF_PERIOD_MS = 400;

const supportsBitmapCaching = (type: DiagramFigure["type"]): boolean => {
  return type !== "image" && type !== "text" && type !== "tag";
};

const resolveUnitSymbol = (tag: DiagramTagMeta | null | undefined): string | undefined => {
  const symbol = tag?.unit?.symbol?.trim();
  if (!symbol) {
    return undefined;
  }

  return symbol;
};

export const FigureRenderer = ({
  figure,
  value,
  enableStaticBitmapCache
}: FigureRendererProps) => {
  const params = useMemo(() => parseParams(figure.params), [figure.params]);

  const hasTagBinding = Boolean(figure.tagId);
  const unitSymbol = resolveUnitSymbol(figure.tag);
  const resolved = useMemo(() => {
    return applyDynamics(params, value, unitSymbol, hasTagBinding);
  }, [hasTagBinding, params, unitSymbol, value]);

  const isDynamic = Boolean(params.dynamics && hasTagBinding);
  const cacheAsBitmap =
    enableStaticBitmapCache && !isDynamic && supportsBitmapCaching(figure.type);

  const [texture, setTexture] = useState<Awaited<ReturnType<typeof loadTexture>>>(null);

  useEffect(() => {
    if (figure.type !== "image") {
      setTexture(null);
      return;
    }

    let isCancelled = false;

    void loadTexture(resolved.src).then((loadedTexture) => {
      if (isCancelled) {
        return;
      }

      setTexture(loadedTexture);
    });

    return () => {
      isCancelled = true;
    };
  }, [figure.type, resolved.src]);

  const [blinkAlpha, setBlinkAlpha] = useState(1);
  const blinkElapsedRef = useRef(0);

  useEffect(() => {
    if (!resolved.blink) {
      setBlinkAlpha(1);
      blinkElapsedRef.current = 0;
    }
  }, [resolved.blink]);

  useTick((ticker) => {
    if (!resolved.blink) {
      return;
    }

    blinkElapsedRef.current += ticker.deltaMS;

    if (blinkElapsedRef.current < BLINK_HALF_PERIOD_MS) {
      return;
    }

    blinkElapsedRef.current = 0;
    setBlinkAlpha((current) => (current === 1 ? 0.4 : 1));
  });

  if (!resolved.visible) {
    return null;
  }

  switch (figure.type) {
    case "rect":
      return (
        <RectShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={cacheAsBitmap} />
      );
    case "circle":
      return (
        <CircleShape
          resolved={resolved}
          alpha={blinkAlpha}
          cacheAsBitmap={cacheAsBitmap}
        />
      );
    case "ellipse":
      return (
        <EllipseShape
          resolved={resolved}
          alpha={blinkAlpha}
          cacheAsBitmap={cacheAsBitmap}
        />
      );
    case "wedge":
      return (
        <WedgeShape
          resolved={resolved}
          alpha={blinkAlpha}
          cacheAsBitmap={cacheAsBitmap}
        />
      );
    case "line":
      return (
        <LineShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={cacheAsBitmap} />
      );
    case "image":
      return (
        <ImageShape
          resolved={resolved}
          alpha={blinkAlpha}
          cacheAsBitmap={false}
          texture={texture}
        />
      );
    case "text":
      return (
        <TextShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={false} />
      );
    case "ring":
      return (
        <RingShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={cacheAsBitmap} />
      );
    case "arc":
      return (
        <ArcShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={cacheAsBitmap} />
      );
    case "tag":
      return (
        <TagShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={false} />
      );
    case "path":
      return (
        <PathShape resolved={resolved} alpha={blinkAlpha} cacheAsBitmap={cacheAsBitmap} />
      );
    default:
      return null;
  }
};
