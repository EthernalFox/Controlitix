import type { KeyboardEvent, MouseEvent } from "react";

import type { Diagram } from "@entities/diagrams";
import {
  ActionIcon,
  Badge,
  Card,
  Group,
  Stack,
  Text,
  Tooltip
} from "@shared/ui";

interface DiagramCardProps {
  diagram: Diagram;
  onEdit: (id: string) => void;
  onDelete: (diagram: Diagram) => void;
}

const FALLBACK_NAME = "Без названия";

const formatDate = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short"
  }).format(new Date(value));

const stopPropagation = (event: MouseEvent<HTMLButtonElement>) => {
  event.stopPropagation();
};

export const DiagramCard = ({ diagram, onEdit, onDelete }: DiagramCardProps) => {
  const isPublished = diagram.publishedAt !== null;
  const statusLabel = isPublished ? "Опубликована" : "Черновик";
  const statusColor = isPublished ? "green" : "gray";
  const dateLabel = isPublished
    ? `Опубликована: ${formatDate(diagram.publishedAt ?? diagram.createdAt)}`
    : `Создана: ${formatDate(diagram.createdAt)}`;

  const onCardKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onEdit(diagram.id);
    }
  };

  return (
    <div
      role="button"
      tabIndex={0}
      style={{ cursor: "pointer", height: "100%" }}
      onClick={() => onEdit(diagram.id)}
      onKeyDown={onCardKeyDown}
    >
      <Card withBorder h="100%" p="md">
        <Stack gap="sm" h="100%">
          <Card withBorder p="md" bg="gray.0">
            <Text c="dimmed" ta="center" fw={600}>
              СХЕМА
            </Text>
          </Card>

          <Stack gap={6} style={{ flexGrow: 1 }}>
            <Text fw={600} lineClamp={2}>
              {diagram.name ?? FALLBACK_NAME}
            </Text>
            {diagram.description && (
              <Text c="dimmed" size="sm" lineClamp={2}>
                {diagram.description}
              </Text>
            )}
            <Badge color={statusColor} variant="light" w="fit-content">
              {statusLabel}
            </Badge>
            <Text c="dimmed" size="xs">
              {dateLabel}
            </Text>
          </Stack>

          <Group justify="flex-end" gap="xs">
            <Tooltip label="Редактировать">
              <ActionIcon
                variant="light"
                aria-label="Edit diagram"
                onClick={(event) => {
                  stopPropagation(event);
                  onEdit(diagram.id);
                }}
              >
                E
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Удалить">
              <ActionIcon
                variant="light"
                color="red"
                aria-label="Delete diagram"
                onClick={(event) => {
                  stopPropagation(event);
                  onDelete(diagram);
                }}
              >
                D
              </ActionIcon>
            </Tooltip>
          </Group>
        </Stack>
      </Card>
    </div>
  );
};

