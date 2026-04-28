import { Application } from "@pixi/react";
import { useEffect, useRef, useState } from "react";
import type { RefObject } from "react";

import type {
  DiagramDetails,
  DiagramSnapshotPoint,
  DiagramViewportState
} from "@/features/diagram";
import { ensurePixiSetup } from "@/shared/libs/pixi/setup";
import type { ConnectionStatus } from "@/shared/modules/realtime";
import { Text } from "@/shared/ui/components";

import styles from "./DiagramPage.module.css";
import { DiagramScene } from "./DiagramScene";
import type { DiagramViewportHandlers } from "./DiagramViewport";
interface DiagramCanvasProps {
  diagram: DiagramDetails;
  viewport: DiagramViewportState;
  valuesByTag: Record<string, DiagramSnapshotPoint | undefined>;
  connectionStatus: ConnectionStatus;
  hostRef: RefObject<HTMLDivElement | null>;
  viewportHandlers: DiagramViewportHandlers;
  onInitialFit: () => void;
}

interface StageSize {
  width: number;
  height: number;
}

export const DiagramCanvas = ({
  diagram,
  viewport,
  valuesByTag,
  connectionStatus,
  hostRef,
  viewportHandlers,
  onInitialFit
}: DiagramCanvasProps) => {
  const [stageSize, setStageSize] = useState<StageSize>({ width: 0, height: 0 });
  const lastFittedDiagramIdRef = useRef<string | null>(null);

  useEffect(() => {
    ensurePixiSetup();
  }, []);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return;
    }

    const updateSize = () => {
      setStageSize({ width: host.clientWidth, height: host.clientHeight });
    };

    updateSize();

    const observer = new ResizeObserver(() => {
      updateSize();
    });

    observer.observe(host);

    return () => {
      observer.disconnect();
    };
  }, [hostRef]);

  useEffect(() => {
    if (stageSize.width <= 0 || stageSize.height <= 0) {
      return;
    }

    if (lastFittedDiagramIdRef.current === diagram.id) {
      return;
    }

    lastFittedDiagramIdRef.current = diagram.id;
    onInitialFit();
  }, [diagram.id, onInitialFit, stageSize.height, stageSize.width]);

  const showConnectionWarning = connectionStatus !== "open";

  return (
    <div
      ref={hostRef}
      className={styles.canvasHost}
      onWheel={viewportHandlers.onWheel}
      onPointerDown={viewportHandlers.onPointerDown}
      onPointerMove={viewportHandlers.onPointerMove}
      onPointerUp={viewportHandlers.onPointerUp}
      onPointerLeave={viewportHandlers.onPointerLeave}
      onContextMenu={viewportHandlers.onContextMenu}
    >
      {showConnectionWarning ? (
        <div className={styles.connectionBanner}>
          <Text size="sm">
            ���������� ��������. ������ ����� ���� �����������.
          </Text>
        </div>
      ) : null}

      {stageSize.width > 0 && stageSize.height > 0 ? (
        <Application width={stageSize.width} height={stageSize.height} antialias>
          <DiagramScene
            diagram={diagram}
            viewport={viewport}
            valuesByTag={valuesByTag}
          />
        </Application>
      ) : null}
    </div>
  );
};
