import {
  IconArrowBackUp,
  IconArrowForwardUp,
  IconArrowLeft,
  IconCircle,
  IconGrid4x4,
  IconLine,
  IconPhoto,
  IconPointer,
  IconSquare,
  IconTypography,
  IconVectorSpline
} from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";

import { diagramsApi } from "@entities/diagrams";
import { useEditorStore } from "@features/editor";
import { buildDiagramsPath } from "@shared/libs/router";
import {
  ActionIcon,
  Badge,
  Button,
  Divider,
  Group,
  Header,
  Modal,
  Stack,
  Text,
  Tooltip
} from "@shared/ui";
import { UserMenu } from "@widgets/UserMenu";

import styles from "./EditorToolbar.module.css";

interface EditorToolbarProps {
  objectId: string;
  diagramId: string;
  diagramName: string | null;
  isPublished: boolean;
  onPublished: () => void;
}

const TOOL_BUTTONS = [
  { key: "select", label: "Select", icon: IconPointer, shortcut: "V" },
  { key: "rect", label: "Rect", icon: IconSquare, shortcut: "R" },
  { key: "circle", label: "Circle", icon: IconCircle, shortcut: "C" },
  { key: "line", label: "Line", icon: IconLine, shortcut: "L" },
  { key: "text", label: "Text", icon: IconTypography, shortcut: "T" },
  { key: "image", label: "Image", icon: IconPhoto, shortcut: "I" }
] as const;

export const EditorToolbar = ({
  diagramId,
  diagramName,
  isPublished,
  objectId,
  onPublished
}: EditorToolbarProps) => {
  const navigate = useNavigate();
  const {
    activeTool,
    gridEnabled,
    setActiveTool,
    setGridEnabled,
    setSaveStatus,
    setSnapEnabled,
    snapEnabled
  } = useEditorStore();

  const [isPublishModalOpen, setIsPublishModalOpen] = useState(false);
  const [isPublishing, setIsPublishing] = useState(false);
  const [publishError, setPublishError] = useState<string | null>(null);

  const publish = async () => {
    setPublishError(null);
    setIsPublishing(true);
    setSaveStatus("saving");

    try {
      await diagramsApi.publish(diagramId);
      onPublished();
      setSaveStatus("saved");
      setIsPublishModalOpen(false);
    } catch (error) {
      setSaveStatus("error");
      if (error instanceof Error) {
        setPublishError(error.message);
      } else {
        setPublishError("Не удалось опубликовать мнемосхему");
      }
    } finally {
      setIsPublishing(false);
    }
  };

  useEffect(() => {
    const onPublishHotkey = () => {
      void publish();
    };

    window.addEventListener("editor:publish", onPublishHotkey as EventListener);
    return () => {
      window.removeEventListener("editor:publish", onPublishHotkey as EventListener);
    };
  }, [publish]);

  return (
    <>
      <Header
        before={
          <Tooltip label="Назад">
            <ActionIcon
              variant="subtle"
              onClick={() => navigate(buildDiagramsPath(objectId))}
              aria-label="Back to diagrams"
            >
              <IconArrowLeft size={18} />
            </ActionIcon>
          </Tooltip>
        }
        main={
          <Group gap="xs" wrap="nowrap">
            <Tooltip label="Undo (Cmd/Ctrl+Z)">
              <ActionIcon
                variant="subtle"
                onClick={() => window.dispatchEvent(new Event("editor:undo"))}
                aria-label="Undo"
              >
                <IconArrowBackUp size={16} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Redo (Cmd/Ctrl+Shift+Z)">
              <ActionIcon
                variant="subtle"
                onClick={() => window.dispatchEvent(new Event("editor:redo"))}
                aria-label="Redo"
              >
                <IconArrowForwardUp size={16} />
              </ActionIcon>
            </Tooltip>

            <Divider orientation="vertical" />

            {TOOL_BUTTONS.map((tool) => {
              const Icon = tool.icon;
              const isActive = activeTool === tool.key;

              return (
                <Tooltip key={tool.key} label={tool.label}>
                  <ActionIcon
                    variant={isActive ? "filled" : "subtle"}
                    onClick={() => setActiveTool(tool.key)}
                    aria-label={tool.label}
                    title={tool.shortcut}
                  >
                    <Icon size={18} />
                  </ActionIcon>
                </Tooltip>
              );
            })}

            <Divider orientation="vertical" />

            <Tooltip label="Grid">
              <ActionIcon
                variant={gridEnabled ? "filled" : "subtle"}
                onClick={() => setGridEnabled(!gridEnabled)}
                aria-label="Toggle grid"
              >
                <IconGrid4x4 size={16} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Snap">
              <ActionIcon
                variant={snapEnabled ? "filled" : "subtle"}
                onClick={() => setSnapEnabled(!snapEnabled)}
                aria-label="Toggle snap"
              >
                <IconVectorSpline size={16} />
              </ActionIcon>
            </Tooltip>
          </Group>
        }
        after={
          <Group gap="sm" wrap="nowrap">
            <Text size="sm" fw={600} truncate maw={220} className={styles.toolbarTitle}>
              {diagramName ?? "Без названия"}
            </Text>
            <Badge color={isPublished ? "green" : "gray"} variant="light">
              {isPublished ? "Опубликована" : "Черновик"}
            </Badge>
            <Button onClick={() => setIsPublishModalOpen(true)}>Опубликовать</Button>
            <UserMenu />
          </Group>
        }
      />

      <Modal
        opened={isPublishModalOpen}
        onClose={() => setIsPublishModalOpen(false)}
        title="Опубликовать мнемосхему?"
      >
        <Stack>
          <Text size="sm" c="dimmed">
            После публикации изменения попадут в рабочую конфигурацию.
          </Text>
          {publishError && <Text c="red">{publishError}</Text>}
          <Button onClick={() => void publish()} loading={isPublishing}>
            Опубликовать
          </Button>
          <Button variant="secondary" onClick={() => setIsPublishModalOpen(false)} disabled={isPublishing}>
            Отмена
          </Button>
        </Stack>
      </Modal>
    </>
  );
};
