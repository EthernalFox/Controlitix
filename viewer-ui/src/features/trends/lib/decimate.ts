import type { TrendPoint } from "@/shared/modules/charts";

interface NumericPoint {
  index: number;
  x: number;
  y: number;
}

const toNumeric = (points: TrendPoint[]): NumericPoint[] => {
  const normalized: NumericPoint[] = [];

  points.forEach((point, index) => {
    if (point.v === null) {
      return;
    }

    normalized.push({
      index,
      x: Date.parse(point.ts),
      y: point.v
    });
  });

  return normalized;
};

const lttbIndices = (points: NumericPoint[], threshold: number): Set<number> => {
  if (points.length <= threshold || threshold < 3) {
    return new Set(points.map((point) => point.index));
  }

  const sampled: NumericPoint[] = [points[0]];
  const bucketSize = (points.length - 2) / (threshold - 2);
  let anchorIndex = 0;

  for (let bucket = 0; bucket < threshold - 2; bucket += 1) {
    const avgRangeStart = Math.floor((bucket + 1) * bucketSize) + 1;
    const avgRangeEnd = Math.min(
      Math.floor((bucket + 2) * bucketSize) + 1,
      points.length
    );

    let avgX = 0;
    let avgY = 0;
    const avgRangeLength = Math.max(1, avgRangeEnd - avgRangeStart);

    for (let i = avgRangeStart; i < avgRangeEnd; i += 1) {
      avgX += points[i].x;
      avgY += points[i].y;
    }

    avgX /= avgRangeLength;
    avgY /= avgRangeLength;

    const rangeStart = Math.floor(bucket * bucketSize) + 1;
    const rangeEnd = Math.min(
      Math.floor((bucket + 1) * bucketSize) + 1,
      points.length - 1
    );

    let maxArea = -1;
    let nextAnchorIndex = rangeStart;

    for (let i = rangeStart; i < rangeEnd; i += 1) {
      const area =
        Math.abs(
          (points[anchorIndex].x - avgX) * (points[i].y - points[anchorIndex].y) -
            (points[anchorIndex].x - points[i].x) * (avgY - points[anchorIndex].y)
        ) * 0.5;

      if (area > maxArea) {
        maxArea = area;
        nextAnchorIndex = i;
      }
    }

    sampled.push(points[nextAnchorIndex]);
    anchorIndex = nextAnchorIndex;
  }

  sampled.push(points[points.length - 1]);
  return new Set(sampled.map((point) => point.index));
};

export const decimateTrendPoints = (
  points: TrendPoint[],
  maxPoints = 500
): TrendPoint[] => {
  if (points.length <= maxPoints) {
    return points;
  }

  const numericPoints = toNumeric(points);
  const nullCount = points.length - numericPoints.length;
  const numericBudget = Math.max(3, maxPoints - nullCount);
  const keepIndices = lttbIndices(numericPoints, numericBudget);

  return points.filter((point, index) => {
    if (point.v === null) {
      return true;
    }

    return keepIndices.has(index);
  });
};
