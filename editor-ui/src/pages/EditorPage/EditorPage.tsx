import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router";

import { diagramsApi } from "@entities/diagrams";
import type { Diagram } from "@entities/diagrams";
import { useFiguresStore } from "@entities/figures";
import { EditorCanvas, FRAME_PRESETS, useEditorShortcuts, useEditorStore } from "@features/editor";
import { Layout, Panel, Stack, Text } from "@shared/ui";
import { ConnectionLostBanner } from "@widgets/EditorBanners";
import { EditorFooter } from "@widgets/EditorFooter";
import { EditorToolbar } from "@widgets/EditorToolbar";
import { LayerList } from "@widgets/LayerList";
import { PropertiesPanel } from "@widgets/PropertiesPanel";
import { ShapePalette } from "@widgets/ShapePalette";

const isFramePreset = (value: unknown): value is { label: string; width: number; height: number } => {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const typedValue = value as Record<string, unknown>;
  return (
    typeof typedValue.label === "string" &&
    typeof typedValue.width === "number" &&
    typeof typedValue.height === "number"
  );
};

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

export default function EditorPage() {
  const { diagramId, objectId } = useParams<{ diagramId: string; objectId: string }>();
  const { fetchFigures, reset: resetFigures } = useFiguresStore();
  const {
    reset: resetEditor,
    setConnectionStatus,
    setFrame,
    setZoom,
    zoom
  } = useEditorStore();

  const onlineTimerRef = useRef<number | null>(null);

  const [diagram, setDiagram] = useState<Diagram | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEditorShortcuts({
    onUndo: () => window.dispatchEvent(new Event("editor:undo")),
    onRedo: () => window.dispatchEvent(new Event("editor:redo")),
    onDelete: () => window.dispatchEvent(new Event("editor:delete-selection")),
    onDuplicate: () => window.dispatchEvent(new Event("editor:duplicate-selection")),
    onZoomIn: () => setZoom(clamp(zoom * 1.1, 0.1, 5)),
    onZoomOut: () => setZoom(clamp(zoom / 1.1, 0.1, 5)),
    onFitToScreen: () => window.dispatchEvent(new Event("editor:fit-to-screen")),
    onPublish: () => window.dispatchEvent(new Event("editor:publish")),
    onBindTag: () => window.dispatchEvent(new Event("editor:open-binding"))
  });

  useEffect(() => {
    const syncConnectionStatus = (nextStatus: "online" | "offline") => {
      if (nextStatus === "offline") {
        if (onlineTimerRef.current !== null) {
          window.clearTimeout(onlineTimerRef.current);
          onlineTimerRef.current = null;
        }

        setConnectionStatus("offline");
        return;
      }

      if (onlineTimerRef.current !== null) {
        window.clearTimeout(onlineTimerRef.current);
      }

      onlineTimerRef.current = window.setTimeout(() => {
        setConnectionStatus("online");
      }, 500);
    };

    syncConnectionStatus(window.navigator.onLine ? "online" : "offline");

    const onOnline = () => syncConnectionStatus("online");
    const onOffline = () => syncConnectionStatus("offline");

    window.addEventListener("online", onOnline);
    window.addEventListener("offline", onOffline);

    return () => {
      window.removeEventListener("online", onOnline);
      window.removeEventListener("offline", onOffline);

      if (onlineTimerRef.current !== null) {
        window.clearTimeout(onlineTimerRef.current);
        onlineTimerRef.current = null;
      }
    };
  }, [setConnectionStatus]);

  useEffect(() => {
    if (!diagramId) {
      return;
    }

    const storedFrame = localStorage.getItem(`editor:frame:${diagramId}`);
    if (storedFrame) {
      try {
        const parsedFrame = JSON.parse(storedFrame);
        if (isFramePreset(parsedFrame)) {
          const knownPreset = FRAME_PRESETS.find((preset) => preset.label === parsedFrame.label) ?? parsedFrame;
          setFrame(knownPreset);
        }
      } catch {
        localStorage.removeItem(`editor:frame:${diagramId}`);
      }
    }
  }, [diagramId, setFrame]);

  useEffect(() => {
    if (!diagramId) {
      setError("Не удалось определить мнемосхему");
      setIsLoading(false);
      return;
    }

    let isCancelled = false;
    setIsLoading(true);
    setError(null);

    void Promise.all([diagramsApi.get(diagramId), fetchFigures(diagramId)])
      .then(([loadedDiagram]) => {
        if (isCancelled) {
          return;
        }

        setDiagram(loadedDiagram);
      })
      .catch((loadError) => {
        if (isCancelled) {
          return;
        }

        if (loadError instanceof Error) {
          setError(loadError.message);
        } else {
          setError("Не удалось загрузить данные редактора");
        }
      })
      .finally(() => {
        if (!isCancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      isCancelled = true;
      resetFigures();
      resetEditor();
    };
  }, [diagramId, fetchFigures, resetEditor, resetFigures]);

  const isPublished = useMemo(() => Boolean(diagram?.publishedAt), [diagram]);
  const connectionStatus = useEditorStore((state) => state.connectionStatus);

  if (!diagramId || !objectId) {
    return (
      <Stack p="md">
        <Text c="red">Не удалось определить параметры маршрута редактора.</Text>
      </Stack>
    );
  }

  if (isLoading) {
    return (
      <Stack p="md">
        <Text c="dimmed">Загрузка редактора...</Text>
      </Stack>
    );
  }

  if (error) {
    return (
      <Stack p="md">
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  return (
    <Layout
      mode="editor"
      header={
        <EditorToolbar
          diagramId={diagramId}
          objectId={objectId}
          diagramName={diagram?.name ?? null}
          isPublished={isPublished}
          onPublished={() =>
            setDiagram((state) =>
              state
                ? {
                    ...state,
                    publishedAt: new Date().toISOString()
                  }
                : state
            )
          }
        />
      }
      panel={<Panel top={<ShapePalette />} center={<LayerList />} />}
      aside={<PropertiesPanel diagramId={diagramId} />}
      footer={<EditorFooter diagramId={diagramId} />}
      sizes={{
        headerHeight: { base: 56, sm: 64 },
        footerHeight: { base: 48, sm: 56 },
        panelWidth: { base: 240, md: 280 },
        asideWidth: { base: 280, md: 320 }
      }}
    >
      <div style={{ width: "100%", height: "100%", display: "flex", flexDirection: "column" }}>
        {connectionStatus === "offline" && <ConnectionLostBanner />}
        <div style={{ flex: 1, minHeight: 0 }}>
          <EditorCanvas diagramId={diagramId} />
        </div>
      </div>
    </Layout>
  );
}
