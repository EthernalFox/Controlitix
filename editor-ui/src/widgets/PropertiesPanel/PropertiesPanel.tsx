import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { useFiguresStore } from "@entities/figures";
import { useEditorStore } from "@features/editor";
import { api } from "@shared/api";
import type { PaginatedResponse } from "@shared/api";
import {
  Button,
  ColorInput,
  NumberInput,
  ScrollArea,
  Select,
  Slider,
  Stack,
  Tabs,
  Text,
  TextInput
} from "@shared/ui";

interface PropertiesPanelProps {
  diagramId: string;
}

interface TagOption {
  id: string;
  name: string;
  deviceId: string;
}

type PropertiesTab = "properties" | "binding";

const readNumber = (value: unknown, fallback: number) => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return fallback;
};

const readString = (value: unknown, fallback: string) =>
  typeof value === "string" ? value : fallback;

const supportsSize = (type: string) => type === "rect" || type === "ellipse" || type === "image";

export const PropertiesPanel = ({ diagramId }: PropertiesPanelProps) => {
  void diagramId;

  const [activeTab, setActiveTab] = useState<PropertiesTab>("properties");
  const [searchQuery, setSearchQuery] = useState("");
  const [tags, setTags] = useState<TagOption[]>([]);
  const [isTagLoading, setIsTagLoading] = useState(false);

  const debounceRef = useRef<number | null>(null);
  const tagSearchRef = useRef<number | null>(null);

  const { figures, patchFigureLocal, updateFigure } = useFiguresStore();
  const { selection, setSaveStatus } = useEditorStore();

  const selectedFigure = useMemo(
    () => figures.find((figure) => figure.id === selection.ids[0]) ?? null,
    [figures, selection.ids]
  );

  const selectedCount = selection.ids.length;

  const upsertFigureDebounced = useCallback(
    (figureId: string, params: Record<string, unknown>) => {
      if (debounceRef.current !== null) {
        window.clearTimeout(debounceRef.current);
      }

      setSaveStatus("saving");
      debounceRef.current = window.setTimeout(() => {
        void updateFigure(figureId, { params })
          .then(() => {
            setSaveStatus("saved");
          })
          .catch(() => {
            setSaveStatus("error");
          });
      }, 500);
    },
    [setSaveStatus, updateFigure]
  );

  const updateFigureParams = (patch: Record<string, unknown>) => {
    if (!selectedFigure) {
      return;
    }

    const nextParams = { ...selectedFigure.params, ...patch };
    patchFigureLocal(selectedFigure.id, { params: patch });
    upsertFigureDebounced(selectedFigure.id, nextParams);
  };

  const updateFigureTag = async (tagId: string | null) => {
    if (!selectedFigure) {
      return;
    }

    patchFigureLocal(selectedFigure.id, { tag_id: tagId });
    setSaveStatus("saving");

    try {
      await updateFigure(selectedFigure.id, { tag_id: tagId });
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

  useEffect(() => {
    const onBindingRequested = () => {
      setActiveTab("binding");
    };

    window.addEventListener("editor:open-binding", onBindingRequested as EventListener);
    return () => {
      window.removeEventListener("editor:open-binding", onBindingRequested as EventListener);
    };
  }, []);

  useEffect(() => {
    return () => {
      if (debounceRef.current !== null) {
        window.clearTimeout(debounceRef.current);
      }

      if (tagSearchRef.current !== null) {
        window.clearTimeout(tagSearchRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (tagSearchRef.current !== null) {
      window.clearTimeout(tagSearchRef.current);
    }

    tagSearchRef.current = window.setTimeout(() => {
      setIsTagLoading(true);
      void api
        .get<PaginatedResponse<TagOption>>("/tags", { search: searchQuery, limit: 20 })
        .then((response) => {
          setTags(response.items);
        })
        .catch(() => {
          setTags([]);
        })
        .finally(() => {
          setIsTagLoading(false);
        });
    }, 300);
  }, [searchQuery]);

  if (selectedCount === 0 || !selectedFigure) {
    return (
      <Stack p="md">
        <Text c="dimmed">Выберите фигуру для редактирования</Text>
      </Stack>
    );
  }

  if (selectedCount >= 2) {
    return (
      <Stack p="md">
        <Text fw={600}>Выбрано фигур: {selectedCount}</Text>
        <Text size="sm" c="dimmed">
          Для мультивыделения доступны только общие операции (перемещение, удаление, порядок слоев).
        </Text>
      </Stack>
    );
  }

  const params = selectedFigure.params;
  const tagOptions = tags.map((tag) => ({
    value: tag.id,
    label: `${tag.name} (${tag.deviceId})`
  }));

  if (selectedFigure.tagId && !tagOptions.some((option) => option.value === selectedFigure.tagId)) {
    tagOptions.unshift({ value: selectedFigure.tagId, label: selectedFigure.tagId });
  }

  return (
    <Tabs value={activeTab} onChange={(value) => setActiveTab((value as PropertiesTab) ?? "properties")}> 
      <Tabs.List>
        <Tabs.Tab value="properties">Свойства</Tabs.Tab>
        <Tabs.Tab value="binding">Привязка данных</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="properties" pt="xs">
        <ScrollArea h="calc(100vh - 220px)" type="auto">
          <Stack p="sm">
            <NumberInput
              label="X"
              value={readNumber(params.x, 0)}
              onChange={(value) => updateFigureParams({ x: readNumber(value, 0) })}
              mono
            />
            <NumberInput
              label="Y"
              value={readNumber(params.y, 0)}
              onChange={(value) => updateFigureParams({ y: readNumber(value, 0) })}
              mono
            />

            {supportsSize(selectedFigure.type) && (
              <>
                <NumberInput
                  label="W"
                  value={readNumber(params.width, 120)}
                  onChange={(value) => updateFigureParams({ width: readNumber(value, 120) })}
                  mono
                />
                <NumberInput
                  label="H"
                  value={readNumber(params.height, 80)}
                  onChange={(value) => updateFigureParams({ height: readNumber(value, 80) })}
                  mono
                />
              </>
            )}

            {selectedFigure.type === "circle" && (
              <NumberInput
                label="Radius"
                value={readNumber(params.radius, 50)}
                onChange={(value) => updateFigureParams({ radius: readNumber(value, 50) })}
                mono
              />
            )}

            <NumberInput
              label="Rotation"
              value={readNumber(params.rotation, 0)}
              onChange={(value) => updateFigureParams({ rotation: readNumber(value, 0) })}
              mono
            />

            <ColorInput
              label="Fill"
              value={readString(params.fill, "#228BE6")}
              onChange={(value) => updateFigureParams({ fill: value })}
            />

            <ColorInput
              label="Stroke"
              value={readString(params.stroke, "#1971C2")}
              onChange={(value) => updateFigureParams({ stroke: value })}
            />

            <NumberInput
              label="Stroke width"
              value={readNumber(params.strokeWidth, 1)}
              onChange={(value) => updateFigureParams({ strokeWidth: readNumber(value, 1) })}
              mono
            />

            <Text size="sm" c="dimmed">
              Opacity
            </Text>
            <Slider
              value={Math.round(readNumber(params.opacity, 1) * 100)}
              min={0}
              max={100}
              onChange={(value) => updateFigureParams({ opacity: value / 100 })}
            />

            {selectedFigure.type === "text" && (
              <>
                <TextInput
                  label="Text"
                  value={readString(params.text, "")}
                  onChange={(event) => updateFigureParams({ text: event.currentTarget.value })}
                />
                <NumberInput
                  label="Font size"
                  value={readNumber(params.fontSize, 16)}
                  onChange={(value) => updateFigureParams({ fontSize: readNumber(value, 16) })}
                  mono
                />
              </>
            )}
          </Stack>
        </ScrollArea>
      </Tabs.Panel>

      <Tabs.Panel value="binding" pt="xs">
        <Stack p="sm">
          <Select
            label="Тег"
            placeholder="Поиск тега..."
            searchable
            clearable
            data={tagOptions}
            searchValue={searchQuery}
            onSearchChange={setSearchQuery}
            value={selectedFigure.tagId}
            onChange={(value) => void updateFigureTag(value)}
            nothingFoundMessage={isTagLoading ? "Загрузка..." : "Ничего не найдено"}
          />

          {selectedFigure.tagId && (
            <Button variant="secondary" onClick={() => void updateFigureTag(null)}>
              Отвязать
            </Button>
          )}
        </Stack>
      </Tabs.Panel>
    </Tabs>
  );
};
