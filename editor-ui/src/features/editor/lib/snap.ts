export interface SnapRect {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface SnapGuide {
  orientation: "vertical" | "horizontal";
  points: [number, number, number, number];
}

export interface SnapComputation {
  guides: SnapGuide[];
  deltaX: number;
  deltaY: number;
}

const EDGE_MARGIN = 24;

const centerX = (rect: SnapRect) => rect.x + rect.width / 2;
const centerY = (rect: SnapRect) => rect.y + rect.height / 2;

const getXPoints = (rect: SnapRect) => [rect.x, centerX(rect), rect.x + rect.width];
const getYPoints = (rect: SnapRect) => [rect.y, centerY(rect), rect.y + rect.height];

const distanceBetweenCenters = (a: SnapRect, b: SnapRect) => {
  const dx = centerX(a) - centerX(b);
  const dy = centerY(a) - centerY(b);
  return Math.sqrt(dx * dx + dy * dy);
};

interface AxisMatch {
  delta: number;
  line: number;
  targetRect: SnapRect;
}

const computeAxisMatch = (
  activePoints: number[],
  candidateRects: SnapRect[],
  threshold: number,
  axis: "x" | "y"
): AxisMatch | null => {
  let bestMatch: AxisMatch | null = null;

  for (const rect of candidateRects) {
    const candidatePoints = axis === "x" ? getXPoints(rect) : getYPoints(rect);

    for (const sourcePoint of activePoints) {
      for (const targetPoint of candidatePoints) {
        const delta = targetPoint - sourcePoint;
        if (Math.abs(delta) > threshold) {
          continue;
        }

        if (!bestMatch || Math.abs(delta) < Math.abs(bestMatch.delta)) {
          bestMatch = {
            delta,
            line: targetPoint,
            targetRect: rect
          };
        }
      }
    }
  }

  return bestMatch;
};

export const computeSnapTargets = (
  rects: SnapRect[],
  draggingId: string,
  threshold: number,
  radius: number = 200
): SnapComputation => {
  const draggingRect = rects.find((rect) => rect.id === draggingId);
  if (!draggingRect) {
    return { guides: [], deltaX: 0, deltaY: 0 };
  }

  const candidates = rects
    .filter((rect) => rect.id !== draggingId)
    .filter((rect) => distanceBetweenCenters(draggingRect, rect) <= radius);

  if (candidates.length === 0) {
    return { guides: [], deltaX: 0, deltaY: 0 };
  }

  const xMatch = computeAxisMatch(getXPoints(draggingRect), candidates, threshold, "x");
  const yMatch = computeAxisMatch(getYPoints(draggingRect), candidates, threshold, "y");

  const nextRect: SnapRect = {
    ...draggingRect,
    x: draggingRect.x + (xMatch?.delta ?? 0),
    y: draggingRect.y + (yMatch?.delta ?? 0)
  };

  const guides: SnapGuide[] = [];

  if (xMatch) {
    const y1 = Math.min(nextRect.y, xMatch.targetRect.y) - EDGE_MARGIN;
    const y2 =
      Math.max(nextRect.y + nextRect.height, xMatch.targetRect.y + xMatch.targetRect.height) +
      EDGE_MARGIN;

    guides.push({
      orientation: "vertical",
      points: [xMatch.line, y1, xMatch.line, y2]
    });
  }

  if (yMatch) {
    const x1 = Math.min(nextRect.x, yMatch.targetRect.x) - EDGE_MARGIN;
    const x2 =
      Math.max(nextRect.x + nextRect.width, yMatch.targetRect.x + yMatch.targetRect.width) +
      EDGE_MARGIN;

    guides.push({
      orientation: "horizontal",
      points: [x1, yMatch.line, x2, yMatch.line]
    });
  }

  return {
    guides,
    deltaX: xMatch?.delta ?? 0,
    deltaY: yMatch?.delta ?? 0
  };
};
