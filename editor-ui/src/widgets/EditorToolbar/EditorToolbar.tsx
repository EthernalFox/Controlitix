import {
  IconArrowLeft,
  IconCircle,
  IconLine,
  IconPhoto,
  IconPointer,
  IconSquare,
  IconTypography
} from "@tabler/icons-react";
import { useState } from "react";
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

interface EditorToolbarProps {
  objectId: string;
  diagramId: string;
  diagramName: string | null;
  isPublished: boolean;
  onPublished: () => void;
}

const TOOL_BUTTONS = [
  { key: "select", label: "Select", icon: IconPointer },
  { key: "rect", label: "Rect", icon: IconSquare },
  { key: "circle", label: "Circle", icon: IconCircle },
  { key: "line", label: "Line", icon: IconLine },
  { key: "text", label: "Text", icon: IconTypography },
  { key: "image", label: "Image", icon: IconPhoto }
] as const;

export const EditorToolbar = ({
  diagramId,
  diagramName,
  isPublished,
  objectId,
  onPublished
}: EditorToolbarProps) => {
  const navigate = useNavigate();
  const { activeTool, setActiveTool, setSaveStatus } = useEditorStore();

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
            {TOOL_BUTTONS.map((tool, index) => {
              const Icon = tool.icon;
              const isActive = activeTool === tool.key;

              return (
                <Group key={tool.key} gap="xs" wrap="nowrap">
                  {index === 1 && <Divider orientation="vertical" />}
                  <Tooltip label={tool.label}>
                    <ActionIcon
                      variant={isActive ? "filled" : "subtle"}
                      onClick={() => setActiveTool(tool.key)}
                      aria-label={tool.label}
                    >
                      <Icon size={18} />
                    </ActionIcon>
                  </Tooltip>
                </Group>
              );
            })}
          </Group>
        }
        after={
          <Group gap="sm" wrap="nowrap">
            <Text size="sm" fw={600} truncate maw={220}>
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
          <Button
            variant="secondary"
            onClick={() => setIsPublishModalOpen(false)}
            disabled={isPublishing}
          >
            Отмена
          </Button>
        </Stack>
      </Modal>
    </>
  );
};
