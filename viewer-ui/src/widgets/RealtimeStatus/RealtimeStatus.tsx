import { useMemo } from "react";

import { realtimeClient, useRealtimeStore } from "@/shared/modules/realtime";
import {
  Badge,
  Button,
  Group,
  Popover,
  Stack,
  Text,
  Tooltip
} from "@/shared/ui/components";

interface StatusMeta {
  color: string;
  label: string;
  tooltip: string;
}

const STATUS_META: Record<string, StatusMeta> = {
  open: {
    color: "var(--mantine-color-alarm-ok-6)",
    label: "Подключён",
    tooltip: "Realtime подключён"
  },
  connecting: {
    color: "var(--mantine-color-alarm-warn-6)",
    label: "Подключаемся",
    tooltip: "Восстанавливаем соединение..."
  },
  reconnecting: {
    color: "var(--mantine-color-alarm-warn-6)",
    label: "Переподключение",
    tooltip: "Восстанавливаем соединение..."
  },
  closed: {
    color: "var(--mantine-color-alarm-comm-6)",
    label: "Отключён",
    tooltip: "Realtime отключён"
  }
};

export const RealtimeStatus = () => {
  const status = useRealtimeStore((state) => state.status);
  const activeTopicCount = useRealtimeStore((state) => state.activeTopicCount);

  const meta = useMemo(() => {
    return STATUS_META[status] ?? STATUS_META.closed;
  }, [status]);

  return (
    <Popover width={260} position="bottom-end" withArrow shadow="md">
      <Popover.Target>
        <Tooltip label={meta.tooltip}>
          <Button
            variant="ghost"
            aria-label="Realtime status"
            style={{ minWidth: 30, width: 30, height: 30, padding: 0 }}
          >
            <span
              style={{
                display: "inline-block",
                width: 10,
                height: 10,
                borderRadius: "50%",
                backgroundColor: meta.color
              }}
            />
          </Button>
        </Tooltip>
      </Popover.Target>

      <Popover.Dropdown>
        <Stack gap="xs">
          <Group justify="space-between" align="center">
            <Text size="sm" c="dimmed">
              Статус
            </Text>
            <Badge color={status === "open" ? "alarm-ok" : status === "closed" ? "alarm-comm" : "alarm-warn"}>
              {meta.label}
            </Badge>
          </Group>

          <Group justify="space-between" align="center">
            <Text size="sm" c="dimmed">
              Активные подписки
            </Text>
            <Text size="sm" fw={600}>
              {activeTopicCount}
            </Text>
          </Group>

          <Button variant="secondary" size="xs" onClick={() => realtimeClient.reconnect()}>
            Переподключить
          </Button>
        </Stack>
      </Popover.Dropdown>
    </Popover>
  );
};
