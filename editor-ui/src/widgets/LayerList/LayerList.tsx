import {
  IconCircle,
  IconEye,
  IconEyeOff,
  IconLine,
  IconPhoto,
  IconSquare,
  IconTypography
} from "@tabler/icons-react";
import { useState } from "react";


import { useFiguresStore } from "@entities/figures";
import { useEditorStore } from "@features/editor";
import { ActionIcon, Group, ScrollArea, Stack, Text } from "@shared/ui";

const ICON_BY_TYPE: Record<string, React.ComponentType<{ size?: number }>> = {
  rect: IconSquare,
  circle: IconCircle,
  line: IconLine,
  text: IconTypography,
  image: IconPhoto
};

const buildFigureTitle = (type: string, index: number, textValue?: string) => {
  if (type === "text" && textValue && textValue.trim() !== "") {
    return textValue;
  }

  return `${type} ${index + 1}`;
};

const reorderArray = (items: string[], sourceId: string, targetId: string) => {
  const sourceIndex = items.indexOf(sourceId);
  const targetIndex = items.indexOf(targetId);
  if (sourceIndex === -1 || targetIndex === -1 || sourceIndex === targetIndex) {
    return items;
  }

  const reordered = [...items];
  const [movedItem] = reordered.splice(sourceIndex, 1);
  reordered.splice(targetIndex, 0, movedItem);
  return reordered;
};

export const LayerList = () => {
  const [draggingFigureId, setDraggingFigureId] = useState<string | null>(null);
  const { figures, patchFigureLocal, reorderFigures, updateFigure } = useFiguresStore();
  const { selectFigure, selectedFigureIds, setSaveStatus } = useEditorStore();

  const displayFigures = [...figures].reverse();

  const toggleVisibility = async (figureId: string, visible: boolean) => {
    patchFigureLocal(figureId, { params: { visible } });
    setSaveStatus("saving");

    try {
      await updateFigure(figureId, { params: { visible } });
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

  return (
    <Stack gap="xs" p="sm" h="100%">
      <Text fw={600} size="sm">
        Слои
      </Text>

      <ScrollArea h="100%" type="auto">
        <Stack gap={4} pr="xs">
          {displayFigures.map((figure, index) => {
            const Icon = ICON_BY_TYPE[figure.type] ?? IconSquare;
            const isSelected = selectedFigureIds.includes(figure.id);
            const isVisible = figure.params.visible !== false;

            return (
              <Group
                key={figure.id}
                justify="space-between"
                wrap="nowrap"
                p="xs"
                draggable
                onDragStart={() => setDraggingFigureId(figure.id)}
                onDragOver={(event) => event.preventDefault()}
                onDrop={() => {
                  if (!draggingFigureId) {
                    return;
                  }

                  const displayOrder = displayFigures.map((item) => item.id);
                  const nextDisplayOrder = reorderArray(
                    displayOrder,
                    draggingFigureId,
                    figure.id
                  );
                  reorderFigures([...nextDisplayOrder].reverse());
                  setDraggingFigureId(null);
                }}
                onDragEnd={() => setDraggingFigureId(null)}
                style={{
                  borderRadius: 8,
                  cursor: "pointer",
                  background: isSelected ? "rgba(34, 139, 230, 0.12)" : "transparent",
                  opacity: isVisible ? 1 : 0.55
                }}
                onClick={() => selectFigure(figure.id)}
              >
                <Group gap="xs" wrap="nowrap" style={{ minWidth: 0 }}>
                  <Icon size={14} />
                  <Text size="sm" truncate>
                    {buildFigureTitle(figure.type, index, String(figure.params.text ?? ""))}
                  </Text>
                </Group>

                <ActionIcon
                  variant="subtle"
                  onClick={(event) => {
                    event.stopPropagation();
                    void toggleVisibility(figure.id, !isVisible);
                  }}
                  aria-label={isVisible ? "Hide layer" : "Show layer"}
                >
                  {isVisible ? <IconEye size={14} /> : <IconEyeOff size={14} />}
                </ActionIcon>
              </Group>
            );
          })}

          {displayFigures.length === 0 && (
            <Text size="sm" c="dimmed">
              Нет фигур
            </Text>
          )}
        </Stack>
      </ScrollArea>
    </Stack>
  );
};
