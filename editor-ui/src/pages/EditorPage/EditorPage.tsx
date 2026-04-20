import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";

import { diagramsApi } from "@entities/diagrams";
import type { Diagram } from "@entities/diagrams";
import { useFiguresStore } from "@entities/figures";
import { EditorCanvas, FRAME_PRESETS, useEditorStore } from "@features/editor";
import { Layout, Panel, Stack, Text } from "@shared/ui";
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

export default function EditorPage() {
  const { diagramId, objectId } = useParams<{ diagramId: string; objectId: string }>();
  const { fetchFigures, reset: resetFigures } = useFiguresStore();
  const { reset: resetEditor, setFrame } = useEditorStore();

  const [diagram, setDiagram] = useState<Diagram | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!diagramId) {
      return;
    }

    const storedFrame = localStorage.getItem(`editor:frame:${diagramId}`);
    if (storedFrame) {
      try {
        const parsedFrame = JSON.parse(storedFrame);
        if (isFramePreset(parsedFrame)) {
          const knownPreset =
            FRAME_PRESETS.find((preset) => preset.label === parsedFrame.label) ?? parsedFrame;
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
      <EditorCanvas diagramId={diagramId} />
    </Layout>
  );
}
