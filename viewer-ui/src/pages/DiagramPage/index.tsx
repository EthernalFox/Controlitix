import { useCallback, useEffect, useMemo, useRef } from "react";
import { useNavigate, useParams } from "react-router";

import {
  fetchDiagram,
  fetchDiagramSnapshot,
  useDiagramStore
} from "@/features/diagram";
import { ApiRequestError } from "@/shared/api";
import { unloadTextures } from "@/shared/libs/pixi/setup";
import { APP_PATHS } from "@/shared/libs/router";
import {
  useRealtimeStore,
  useTopicSubscription,
  type ConfigChangedMessage,
  type RealtimeTagValue
} from "@/shared/modules/realtime";
import { Button, Card, Group, Loader, Stack, Text, notify } from "@/shared/ui/components";

import { DiagramCanvas } from "./DiagramCanvas";
import { DiagramHeader } from "./DiagramHeader";
import styles from "./DiagramPage.module.css";
import { useDiagramViewport } from "./DiagramViewport";

const DIAGRAM_CONFIG_NOTIFICATION_ID = "diagram-config-changed";
const CONFIG_CHANGE_COALESCE_MS = 5000;

const resolveErrorMessage = (error: unknown): string => {
  if (!(error instanceof ApiRequestError)) {
    return "Не удалось загрузить мнемосхему";
  }

  if (error.payload.type === "/errors/diagrams/not-published") {
    return "Мнемосхема не опубликована";
  }

  if (error.payload.type === "/errors/diagrams/not-found") {
    return "Мнемосхема не найдена";
  }

  return error.payload.detail || error.payload.title || "Не удалось загрузить мнемосхему";
};

const collectImageSources = (
  figures: Array<{ type: string; params: Record<string, unknown> }>
): string[] => {
  const uniqueSources = new Set<string>();

  figures.forEach((figure) => {
    if (figure.type !== "image") {
      return;
    }

    const source = typeof figure.params.src === "string" ? figure.params.src.trim() : "";
    if (!source) {
      return;
    }

    uniqueSources.add(source);
  });

  return Array.from(uniqueSources);
};

const isRelevantConfigChange = (
  message: ConfigChangedMessage | null,
  diagramId: string | undefined
): boolean => {
  if (!message || !diagramId) {
    return false;
  }

  if (message.entity_type === "diagram") {
    return message.entity_id === diagramId;
  }

  if (message.entity_type !== "figure") {
    return false;
  }

  const payload = message.payload;
  const payloadDiagramID =
    payload && typeof payload === "object" && typeof payload.diagram_id === "string"
      ? payload.diagram_id
      : null;

  if (payloadDiagramID) {
    return payloadDiagramID === diagramId;
  }

  return true;
};

const toTimestamp = (value: string): number => {
  const parsed = Date.parse(value);
  if (!Number.isFinite(parsed)) {
    return 0;
  }

  return parsed;
};

export const DiagramPage = () => {
  const { diagramId } = useParams<{ objectId: string; diagramId: string }>();
  const navigate = useNavigate();

  const diagram = useDiagramStore((state) => state.diagram);
  const valuesByTag = useDiagramStore((state) => state.valuesByTag);
  const viewport = useDiagramStore((state) => state.viewport);
  const isLoading = useDiagramStore((state) => state.isLoading);
  const error = useDiagramStore((state) => state.error);
  const setDiagram = useDiagramStore((state) => state.setDiagram);
  const setSnapshot = useDiagramStore((state) => state.setSnapshot);
  const setValue = useDiagramStore((state) => state.setValue);
  const setViewport = useDiagramStore((state) => state.setViewport);
  const setLoading = useDiagramStore((state) => state.setLoading);
  const setError = useDiagramStore((state) => state.setError);
  const resetDiagramState = useDiagramStore((state) => state.resetDiagramState);
  const connectionStatus = useRealtimeStore((state) => state.status);
  const lastConfigChange = useRealtimeStore((state) => state.lastConfigChange);

  const lastSuccessfulFetchAtRef = useRef<number>(0);
  const lastConfigHandledAtRef = useRef<number>(0);
  const notificationAutoRefreshRef = useRef(false);
  const notificationVisibleRef = useRef(false);

  const loadDiagram = useCallback(
    async (silent = false) => {
      if (!diagramId) {
        setError("Некорректный идентификатор мнемосхемы");
        return;
      }

      if (!silent) {
        setLoading(true);
      }
      setError(null);

      try {
        const [nextDiagram, snapshot] = await Promise.all([
          fetchDiagram(diagramId),
          fetchDiagramSnapshot(diagramId)
        ]);

        setDiagram(nextDiagram);
        setSnapshot(snapshot.ts, snapshot.values, snapshot.missingTagIds);
        lastSuccessfulFetchAtRef.current = Date.now();
      } catch (loadError: unknown) {
        if (!silent) {
          setDiagram(null);
        }

        setError(resolveErrorMessage(loadError));
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [diagramId, setDiagram, setError, setLoading, setSnapshot]
  );

  useEffect(() => {
    void loadDiagram(false);

    return () => {
      notify.hide(DIAGRAM_CONFIG_NOTIFICATION_ID);
      notificationVisibleRef.current = false;
      resetDiagramState();
    };
  }, [loadDiagram, resetDiagramState]);

  const diagramTopics = useMemo(() => {
    if (!diagramId) {
      return [];
    }

    return [`diagram:${diagramId}`];
  }, [diagramId]);

  const handleRealtimeValue = useCallback(
    (message: RealtimeTagValue) => {
      setValue({
        tagId: message.tag_id,
        ts: message.ts,
        v: message.v,
        q: message.q
      });
    },
    [setValue]
  );

  useTopicSubscription(diagramTopics, handleRealtimeValue);

  useEffect(() => {
    const configChange = lastConfigChange;
    if (!configChange || !isRelevantConfigChange(configChange, diagramId)) {
      return;
    }

    const eventTimestamp = toTimestamp(configChange.timestamp);
    if (eventTimestamp > 0 && eventTimestamp <= lastSuccessfulFetchAtRef.current) {
      return;
    }

    const now = Date.now();
    if (now - lastConfigHandledAtRef.current < CONFIG_CHANGE_COALESCE_MS) {
      return;
    }

    if (notificationVisibleRef.current) {
      return;
    }

    lastConfigHandledAtRef.current = now;
    notificationVisibleRef.current = true;
    notificationAutoRefreshRef.current = true;

    notify.show({
      id: DIAGRAM_CONFIG_NOTIFICATION_ID,
      title: "Мнемосхема обновлена",
      message: (
        <Group gap="sm" align="center" wrap="nowrap">
          <Text size="sm">Откройте новую версию</Text>
          <Button
            size="xs"
            variant="secondary"
            onClick={() => {
              notificationAutoRefreshRef.current = false;
              notify.hide(DIAGRAM_CONFIG_NOTIFICATION_ID);
              void loadDiagram(true);
            }}
          >
            Обновить
          </Button>
        </Group>
      ),
      color: "alarm-warn",
      autoClose: 30000,
      onClose: () => {
        notificationVisibleRef.current = false;
        if (!notificationAutoRefreshRef.current) {
          return;
        }

        notificationAutoRefreshRef.current = false;
        void loadDiagram(true);
      }
    });
  }, [diagramId, lastConfigChange, loadDiagram]);

  const imageSources = useMemo(() => {
    if (!diagram) {
      return [];
    }

    return collectImageSources(diagram.figures);
  }, [diagram]);

  useEffect(() => {
    return () => {
      void unloadTextures(imageSources);
    };
  }, [imageSources]);

  const canvasHostRef = useRef<HTMLDivElement | null>(null);

  const viewportControls = useDiagramViewport({
    containerRef: canvasHostRef,
    canvas: {
      width: diagram?.canvas.width ?? 1920,
      height: diagram?.canvas.height ?? 1080
    },
    viewport,
    onChange: setViewport
  });

  if (isLoading && !diagram) {
    return (
      <Card withBorder p="lg" className={styles.centerBox}>
        <Group justify="center" align="center" className={styles.centerBox}>
          <Loader />
        </Group>
      </Card>
    );
  }

  if (error && !diagram) {
    return (
      <Card withBorder p="lg" className={styles.centerBox}>
        <Stack gap="md" align="center" justify="center" className={styles.centerBox}>
          <Text>{error}</Text>
          <Button
            variant="secondary"
            onClick={() => void navigate(APP_PATHS.DASHBOARD, { replace: true })}
          >
            Назад
          </Button>
        </Stack>
      </Card>
    );
  }

  if (!diagram) {
    return null;
  }

  return (
    <div className={styles.page}>
      <DiagramHeader
        objectName={diagram.objectName}
        diagramName={diagram.name}
        scale={viewport.scale}
        onBack={() => void navigate(APP_PATHS.DASHBOARD)}
        onFit={viewportControls.fitToScreen}
        onReset={viewportControls.resetToActualSize}
        onZoomIn={viewportControls.zoomIn}
        onZoomOut={viewportControls.zoomOut}
      />

      <DiagramCanvas
        diagram={diagram}
        viewport={viewport}
        valuesByTag={valuesByTag}
        connectionStatus={connectionStatus}
        hostRef={canvasHostRef}
        viewportHandlers={viewportControls.handlers}
        onInitialFit={viewportControls.fitToScreen}
      />
    </div>
  );
};
