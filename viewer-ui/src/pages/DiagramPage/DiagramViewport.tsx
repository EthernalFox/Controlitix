import { useCallback, useEffect, useRef } from "react";
import type {
  PointerEvent as ReactPointerEvent,
  RefObject,
  WheelEvent as ReactWheelEvent
} from "react";

import type { DiagramViewportState } from "@/features/diagram";

const MIN_SCALE = 0.1;
const MAX_SCALE = 8;
const PAN_VISIBLE_MARGIN = 64;

interface CanvasMetrics {
  width: number;
  height: number;
}

interface UseDiagramViewportOptions {
  containerRef: RefObject<HTMLDivElement | null>;
  canvas: CanvasMetrics;
  viewport: DiagramViewportState;
  onChange: (nextViewport: DiagramViewportState) => void;
}

export interface DiagramViewportHandlers {
  onWheel: (event: ReactWheelEvent<HTMLDivElement>) => void;
  onPointerDown: (event: ReactPointerEvent<HTMLDivElement>) => void;
  onPointerMove: (event: ReactPointerEvent<HTMLDivElement>) => void;
  onPointerUp: (event: ReactPointerEvent<HTMLDivElement>) => void;
  onPointerLeave: (event: ReactPointerEvent<HTMLDivElement>) => void;
  onContextMenu: (event: ReactPointerEvent<HTMLDivElement>) => void;
}

interface DiagramViewportControls {
  handlers: DiagramViewportHandlers;
  fitToScreen: () => void;
  resetToActualSize: () => void;
  zoomIn: () => void;
  zoomOut: () => void;
}

const clamp = (value: number, min: number, max: number): number => {
  if (value < min) {
    return min;
  }

  if (value > max) {
    return max;
  }

  return value;
};

const clampAxis = (
  value: number,
  scaledSize: number,
  viewportSize: number
): number => {
  const minValue = PAN_VISIBLE_MARGIN - scaledSize;
  const maxValue = viewportSize - PAN_VISIBLE_MARGIN;

  if (minValue > maxValue) {
    return (viewportSize - scaledSize) / 2;
  }

  return clamp(value, minValue, maxValue);
};

const clampScale = (scale: number): number => {
  return clamp(scale, MIN_SCALE, MAX_SCALE);
};

const clampViewport = (
  viewport: DiagramViewportState,
  viewportSize: CanvasMetrics,
  canvasSize: CanvasMetrics
): DiagramViewportState => {
  const scale = clampScale(viewport.scale);
  const scaledWidth = canvasSize.width * scale;
  const scaledHeight = canvasSize.height * scale;

  return {
    scale,
    x: clampAxis(viewport.x, scaledWidth, viewportSize.width),
    y: clampAxis(viewport.y, scaledHeight, viewportSize.height)
  };
};

const getContainerSize = (
  containerRef: RefObject<HTMLDivElement | null>
): CanvasMetrics | null => {
  const container = containerRef.current;
  if (!container) {
    return null;
  }

  const width = container.clientWidth;
  const height = container.clientHeight;

  if (width <= 0 || height <= 0) {
    return null;
  }

  return { width, height };
};

const buildFitViewport = (
  viewportSize: CanvasMetrics,
  canvasSize: CanvasMetrics
): DiagramViewportState => {
  const fitScale = clampScale(
    Math.min(viewportSize.width / canvasSize.width, viewportSize.height / canvasSize.height)
  );

  const scaledWidth = canvasSize.width * fitScale;
  const scaledHeight = canvasSize.height * fitScale;

  return clampViewport(
    {
      scale: fitScale,
      x: (viewportSize.width - scaledWidth) / 2,
      y: (viewportSize.height - scaledHeight) / 2
    },
    viewportSize,
    canvasSize
  );
};

export const useDiagramViewport = ({
  containerRef,
  canvas,
  viewport,
  onChange
}: UseDiagramViewportOptions): DiagramViewportControls => {
  const viewportRef = useRef(viewport);
  const isSpacePressedRef = useRef(false);
  const dragStateRef = useRef<{
    pointerId: number;
    lastX: number;
    lastY: number;
  } | null>(null);

  useEffect(() => {
    viewportRef.current = viewport;
  }, [viewport]);

  const applyViewport = useCallback(
    (nextViewport: DiagramViewportState) => {
      const viewportSize = getContainerSize(containerRef);
      if (!viewportSize) {
        onChange(nextViewport);
        return;
      }

      onChange(clampViewport(nextViewport, viewportSize, canvas));
    },
    [canvas, containerRef, onChange]
  );

  const fitToScreen = useCallback(() => {
    const viewportSize = getContainerSize(containerRef);
    if (!viewportSize) {
      return;
    }

    onChange(buildFitViewport(viewportSize, canvas));
  }, [canvas, containerRef, onChange]);

  const resetToActualSize = useCallback(() => {
    const viewportSize = getContainerSize(containerRef);
    if (!viewportSize) {
      return;
    }

    const scaledWidth = canvas.width;
    const scaledHeight = canvas.height;

    onChange(
      clampViewport(
        {
          scale: 1,
          x: (viewportSize.width - scaledWidth) / 2,
          y: (viewportSize.height - scaledHeight) / 2
        },
        viewportSize,
        canvas
      )
    );
  }, [canvas, containerRef, onChange]);

  const zoomAt = useCallback(
    (factor: number, anchorX: number, anchorY: number) => {
      const viewportSize = getContainerSize(containerRef);
      if (!viewportSize) {
        return;
      }

      const currentViewport = viewportRef.current;
      const nextScale = clampScale(currentViewport.scale * factor);
      if (nextScale === currentViewport.scale) {
        return;
      }

      const worldX = (anchorX - currentViewport.x) / currentViewport.scale;
      const worldY = (anchorY - currentViewport.y) / currentViewport.scale;

      applyViewport({
        scale: nextScale,
        x: anchorX - worldX * nextScale,
        y: anchorY - worldY * nextScale
      });
    },
    [applyViewport, containerRef]
  );

  const zoomIn = useCallback(() => {
    const viewportSize = getContainerSize(containerRef);
    if (!viewportSize) {
      return;
    }

    zoomAt(1.1, viewportSize.width / 2, viewportSize.height / 2);
  }, [containerRef, zoomAt]);

  const zoomOut = useCallback(() => {
    const viewportSize = getContainerSize(containerRef);
    if (!viewportSize) {
      return;
    }

    zoomAt(0.9, viewportSize.width / 2, viewportSize.height / 2);
  }, [containerRef, zoomAt]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.code === "Space") {
        isSpacePressedRef.current = true;
      }

      if (!event.ctrlKey) {
        return;
      }

      if (event.key === "0") {
        event.preventDefault();
        fitToScreen();
        return;
      }

      if (event.key === "1") {
        event.preventDefault();
        resetToActualSize();
        return;
      }

      if (event.key === "+" || event.key === "=") {
        event.preventDefault();
        zoomIn();
        return;
      }

      if (event.key === "-" || event.key === "_") {
        event.preventDefault();
        zoomOut();
      }
    };

    const handleKeyUp = (event: KeyboardEvent) => {
      if (event.code === "Space") {
        isSpacePressedRef.current = false;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [fitToScreen, resetToActualSize, zoomIn, zoomOut]);

  const onWheel = useCallback(
    (event: ReactWheelEvent<HTMLDivElement>) => {
      event.preventDefault();

      const rect = event.currentTarget.getBoundingClientRect();
      const anchorX = event.clientX - rect.left;
      const anchorY = event.clientY - rect.top;

      zoomAt(event.deltaY < 0 ? 1.1 : 0.9, anchorX, anchorY);
    },
    [zoomAt]
  );

  const onPointerDown = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    const canStartDrag =
      event.button === 1 ||
      event.button === 2 ||
      (event.button === 0 && isSpacePressedRef.current);

    if (!canStartDrag) {
      return;
    }

    event.preventDefault();

    dragStateRef.current = {
      pointerId: event.pointerId,
      lastX: event.clientX,
      lastY: event.clientY
    };

    event.currentTarget.setPointerCapture(event.pointerId);
  }, []);

  const onPointerMove = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const dragState = dragStateRef.current;
      if (!dragState || dragState.pointerId !== event.pointerId) {
        return;
      }

      const deltaX = event.clientX - dragState.lastX;
      const deltaY = event.clientY - dragState.lastY;

      dragStateRef.current = {
        ...dragState,
        lastX: event.clientX,
        lastY: event.clientY
      };

      const currentViewport = viewportRef.current;
      applyViewport({
        ...currentViewport,
        x: currentViewport.x + deltaX,
        y: currentViewport.y + deltaY
      });
    },
    [applyViewport]
  );

  const onPointerUp = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    const dragState = dragStateRef.current;
    if (!dragState || dragState.pointerId !== event.pointerId) {
      return;
    }

    dragStateRef.current = null;

    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  }, []);

  const onPointerLeave = useCallback(() => {
    dragStateRef.current = null;
  }, []);

  const onContextMenu = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    event.preventDefault();
  }, []);

  return {
    handlers: {
      onWheel,
      onPointerDown,
      onPointerMove,
      onPointerUp,
      onPointerLeave,
      onContextMenu
    },
    fitToScreen,
    resetToActualSize,
    zoomIn,
    zoomOut
  };
};
