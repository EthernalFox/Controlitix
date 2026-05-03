import type { ObjectAlarmCounts } from "@/features/objects-summary/lib/aggregateAlarms";
import type { ObjectSummary } from "@/shared/api/objects";
import { APP_PATHS } from "@/shared/libs/router";
import { Badge, Card, Group, Stack, Text, Title, Tooltip } from "@/shared/ui/components";

import styles from "./ObjectCard.module.css";

interface ObjectCardProps {
  object: ObjectSummary;
  counts: ObjectAlarmCounts;
  onOpen: (path: string) => void;
}

const getPrimaryBadge = (counts: ObjectAlarmCounts): { color: string; text: string } => {
  if (counts.alarm > 0) {
    return {
      color: "alarm-alarm",
      text: `${counts.alarm} тревог`
    };
  }

  if (counts.warn > 0) {
    return {
      color: "alarm-warn",
      text: `${counts.warn} предупреждений`
    };
  }

  if (counts.comm > 0) {
    return {
      color: "alarm-comm",
      text: `${counts.comm} offline`
    };
  }

  return {
    color: "gray",
    text: "OK"
  };
};

const getDeviceHealthMarker = (total: number, offline: number): string => {
  if (offline <= 0 || total <= 0) {
    return "🟢";
  }

  if (offline < total / 2) {
    return "🟡";
  }

  return "🔴";
};

const getSubTitle = (object: ObjectSummary): string => {
  if (object.description && object.description.trim().length > 0) {
    return object.description;
  }

  return `${object.devicesTotal} устройств`;
};

export const ObjectCard = ({ object, counts, onOpen }: ObjectCardProps) => {
  const primary = getPrimaryBadge(counts);
  const canOpen = Boolean(object.defaultDiagramId);

  const clickPath = object.defaultDiagramId
    ? APP_PATHS.DIAGRAM_BY_ID(object.defaultDiagramId)
    : "";

  const healthMarker = getDeviceHealthMarker(object.devicesTotal, object.devicesOffline);

  const cardBody = (
    <Card
      withBorder
      p="md"
      className={`${styles.card} ${!canOpen ? styles.disabled : ""}`}
      onClick={() => {
        if (clickPath) {
          onOpen(clickPath);
        }
      }}
      role="button"
      tabIndex={canOpen ? 0 : -1}
      onKeyDown={(event) => {
        if (!clickPath) {
          return;
        }

        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onOpen(clickPath);
        }
      }}
    >
      <Stack gap="sm">
        <Stack gap={4}>
          <Title order={4} className={styles.title}>
            {object.name}
          </Title>
          <Text size="sm" c="dimmed" className={styles.subtitle}>
            {getSubTitle(object)}
          </Text>
        </Stack>

        <Group justify="space-between" align="center" wrap="nowrap">
          <Badge color={primary.color}>{primary.text}</Badge>
          <Text size="sm" fw={600}>
            {healthMarker} {object.devicesOffline}/{object.devicesTotal}
          </Text>
        </Group>

        <Group gap="xs" className={styles.secondaryCounts}>
          {counts.acked > 0 ? <Text size="xs">acked: {counts.acked}</Text> : null}
          {counts.bad > 0 ? <Text size="xs">bad: {counts.bad}</Text> : null}
        </Group>
      </Stack>
    </Card>
  );

  if (canOpen) {
    return cardBody;
  }

  return (
    <Tooltip label="Нет опубликованных мнемосхем">
      <div>{cardBody}</div>
    </Tooltip>
  );
};
