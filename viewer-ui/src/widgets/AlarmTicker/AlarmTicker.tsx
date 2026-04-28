import { useEffect, useMemo } from "react";
import { Link } from "react-router";

import { useAlarmsStore } from "@/features/alarms";
import { APP_PATHS } from "@/shared/libs/router";
import { Badge, Group, Text } from "@/shared/ui/components";

const resolveColor = (state: string): string => {
  switch (state) {
    case "hihi":
    case "lolo":
      return "alarm-alarm";
    case "hi":
    case "lo":
      return "alarm-warn";
    case "uncertain":
      return "alarm-uncertain";
    case "bad":
      return "alarm-bad";
    case "comm_loss":
      return "alarm-comm";
    case "offline":
      return "alarm-offline";
    default:
      return "alarm-ok";
  }
};

export const AlarmTicker = () => {
  const activeUnacked = useAlarmsStore((state) => state.activeUnacked);
  const startRealtime = useAlarmsStore((state) => state.startRealtime);
  const stopRealtime = useAlarmsStore((state) => state.stopRealtime);

  useEffect(() => {
    startRealtime();

    return () => {
      stopRealtime();
    };
  }, [startRealtime, stopRealtime]);

  const items = useMemo(() => {
    return activeUnacked.slice(0, 5);
  }, [activeUnacked]);

  if (items.length === 0) {
    return (
      <Group gap="xs" wrap="nowrap">
        <Badge color="alarm-ok">OK</Badge>
        <Text size="sm" c="dimmed">
          No active alarms
        </Text>
      </Group>
    );
  }

  return (
    <Link
      to={APP_PATHS.ALARMS}
      style={{ color: "inherit", textDecoration: "none", minWidth: 0 }}
    >
      <Group gap={4} wrap="nowrap">
        {items.map((item) => (
          <Badge key={item.tagId} color={resolveColor(item.state)} title={item.tagName}>
            {item.tagName || item.tagId.slice(0, 8)}
          </Badge>
        ))}
      </Group>
    </Link>
  );
};
