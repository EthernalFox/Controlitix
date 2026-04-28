import type { Quality } from "@/shared/modules/charts";

export interface Point2D {
  x: number;
  y: number;
}

export interface FigureDynamics {
  fillByQuality?: Partial<Record<Quality, string>>;
  textFromValue?: {
    format?: string;
    appendUnit?: boolean;
  };
  visibilityByQuality?: Quality[];
  blinkOnAlarm?: boolean;
}

export interface ParsedFigureParams {
  x: number;
  y: number;
  width: number;
  height: number;
  radius: number;
  radiusX: number;
  radiusY: number;
  innerRadius: number;
  outerRadius: number;
  angle: number;
  rotation: number;
  points: Point2D[];
  src: string;
  text: string;
  fontSize: number;
  fontFamily: string;
  fill: string;
  stroke: string;
  strokeWidth: number;
  opacity: number;
  visible: boolean;
  dynamics?: FigureDynamics;
}

const toNumber = (value: unknown, fallback: number): number => {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }

  return parsed;
};

const toString = (value: unknown, fallback: string): string => {
  if (typeof value !== "string") {
    return fallback;
  }

  const normalized = value.trim();
  return normalized || fallback;
};

const toBoolean = (value: unknown, fallback: boolean): boolean => {
  if (typeof value !== "boolean") {
    return fallback;
  }

  return value;
};

const parsePoint = (value: unknown): Point2D | null => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }

  const point = value as Record<string, unknown>;

  return {
    x: toNumber(point.x, 0),
    y: toNumber(point.y, 0)
  };
};

const parsePoints = (value: unknown): Point2D[] => {
  if (!Array.isArray(value)) {
    return [];
  }

  return value
    .map(parsePoint)
    .filter((point): point is Point2D => point !== null);
};

const parseDynamics = (value: unknown): FigureDynamics | undefined => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return undefined;
  }

  const rawDynamics = value as Record<string, unknown>;
  const fillByQualityRaw = rawDynamics.fillByQuality;

  const fillByQuality =
    fillByQualityRaw && typeof fillByQualityRaw === "object" && !Array.isArray(fillByQualityRaw)
      ? (fillByQualityRaw as Partial<Record<Quality, string>>)
      : undefined;

  const textFromValueRaw = rawDynamics.textFromValue;
  const textFromValue =
    textFromValueRaw &&
    typeof textFromValueRaw === "object" &&
    !Array.isArray(textFromValueRaw)
      ? {
          format: toString((textFromValueRaw as Record<string, unknown>).format, ""),
          appendUnit: toBoolean((textFromValueRaw as Record<string, unknown>).appendUnit, false)
        }
      : undefined;

  const visibilityByQualityRaw = rawDynamics.visibilityByQuality;
  const visibilityByQuality = Array.isArray(visibilityByQualityRaw)
    ? (visibilityByQualityRaw.filter((entry) => typeof entry === "string") as Quality[])
    : undefined;

  const blinkOnAlarm = toBoolean(rawDynamics.blinkOnAlarm, false);

  return {
    fillByQuality,
    textFromValue,
    visibilityByQuality,
    blinkOnAlarm
  };
};

export const parseParams = (params: Record<string, unknown>): ParsedFigureParams => {
  return {
    x: toNumber(params.x, 0),
    y: toNumber(params.y, 0),
    width: Math.max(0, toNumber(params.width, 80)),
    height: Math.max(0, toNumber(params.height, 40)),
    radius: Math.max(0, toNumber(params.radius, 30)),
    radiusX: Math.max(0, toNumber(params.radiusX, 40)),
    radiusY: Math.max(0, toNumber(params.radiusY, 20)),
    innerRadius: Math.max(0, toNumber(params.innerRadius, 18)),
    outerRadius: Math.max(0, toNumber(params.outerRadius, 30)),
    angle: toNumber(params.angle, 270),
    rotation: toNumber(params.rotation, 0),
    points: parsePoints(params.points),
    src: toString(params.src, ""),
    text: toString(params.text, ""),
    fontSize: Math.max(1, toNumber(params.fontSize, 14)),
    fontFamily: toString(params.fontFamily, "Inter"),
    fill: toString(params.fill, "#FFFFFF"),
    stroke: toString(params.stroke, "#000000"),
    strokeWidth: Math.max(0, toNumber(params.strokeWidth, 1)),
    opacity: Math.max(0, Math.min(1, toNumber(params.opacity, 1))),
    visible: toBoolean(params.visible, true),
    dynamics: parseDynamics(params.dynamics)
  };
};
