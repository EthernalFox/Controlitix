import type { DiagramSnapshotPoint } from "@/features/diagram";
import { hexToPixiColor } from "@/shared/libs/pixi/colors";
import type { Quality } from "@/shared/modules/charts";

import type { ParsedFigureParams } from "./parseParams";

export interface ResolvedFigureProps extends ParsedFigureParams {
  fillColor: number;
  strokeColor: number;
  text: string;
  visible: boolean;
  blink: boolean;
  quality: Quality;
}

const formatNumericValue = (format: string, value: number): string => {
  const normalizedFormat = format.trim();
  const floatMatch = /^%\.?([0-9]*)f$/i.exec(normalizedFormat);
  if (!floatMatch) {
    return String(value);
  }

  const precision = floatMatch[1] === "" ? 0 : Number.parseInt(floatMatch[1], 10);
  if (!Number.isFinite(precision) || precision < 0) {
    return String(value);
  }

  return value.toFixed(precision);
};

const resolveQuality = (value: DiagramSnapshotPoint | undefined): Quality => {
  if (!value) {
    return "bad";
  }

  return value.q;
};

const resolveText = (
  params: ParsedFigureParams,
  value: DiagramSnapshotPoint | undefined,
  unitSymbol: string | undefined,
  hasTagBinding: boolean
): string => {
  if (!hasTagBinding || !params.dynamics?.textFromValue) {
    return params.text;
  }

  if (!value || value.v === null || value.v === undefined) {
    return "—";
  }

  const baseText = params.dynamics.textFromValue.format
    ? formatNumericValue(params.dynamics.textFromValue.format, value.v)
    : String(value.v);

  if (!params.dynamics.textFromValue.appendUnit || !unitSymbol) {
    return baseText;
  }

  return `${baseText} ${unitSymbol}`;
};

export const applyDynamics = (
  params: ParsedFigureParams,
  value: DiagramSnapshotPoint | undefined,
  unitSymbol: string | undefined,
  hasTagBinding: boolean
): ResolvedFigureProps => {
  const quality = resolveQuality(value);

  if (!params.dynamics || !hasTagBinding) {
    return {
      ...params,
      text: params.text,
      fillColor: hexToPixiColor(params.fill),
      strokeColor: hexToPixiColor(params.stroke),
      visible: params.visible,
      blink: false,
      quality
    };
  }

  const fillByQuality = params.dynamics.fillByQuality;
  const resolvedFillHex = (fillByQuality && fillByQuality[quality]) || params.fill;

  const visibilityByQuality = params.dynamics.visibilityByQuality;
  const visible =
    params.visible &&
    (!visibilityByQuality || visibilityByQuality.length === 0
      ? true
      : visibilityByQuality.includes(quality));

  const blink = Boolean(params.dynamics.blinkOnAlarm) && (quality === "hi" || quality === "hihi");

  return {
    ...params,
    fill: resolvedFillHex,
    text: resolveText(params, value, unitSymbol, hasTagBinding),
    fillColor: hexToPixiColor(resolvedFillHex),
    strokeColor: hexToPixiColor(params.stroke),
    visible,
    blink,
    quality
  };
};
