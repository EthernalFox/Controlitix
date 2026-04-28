import { useMemo } from "react";

import { useTrendsStore } from "@/features/trends";
import {
  Badge,
  Button,
  Card,
  Group,
  Stack,
  Text,
  Tooltip
} from "@/shared/ui/components";

const PALETTE = [
  "#659cff",
  "#0f68ff",
  "#004fcd",
  "#7a8cdd",
  "#3a51cd",
  "#22348f"
];

const formatLatestValue = (values: Array<number | null>): string => {
  const latestValue = [...values].reverse().find((value) => value !== null);

  if (latestValue === undefined || latestValue === null) {
    return "—";
  }

  return latestValue.toFixed(3);
};

export const TrendLegend = () => {
  const selectedTagIds = useTrendsStore((state) => state.selectedTagIds);
  const seriesByTag = useTrendsStore((state) => state.seriesByTag);
  const loadingByTag = useTrendsStore((state) => state.loadingByTag);
  const errorByTag = useTrendsStore((state) => state.errorByTag);
  const removeTag = useTrendsStore((state) => state.removeTag);
  const reload = useTrendsStore((state) => state.reload);

  const rows = useMemo(() => {
    return selectedTagIds.map((tagId, index) => {
      const series = seriesByTag[tagId];
      const unit = series?.unit.symbol || series?.unit.name || "";
      const latestValue = formatLatestValue(series?.points.map((point) => point.v) ?? []);
      const error = errorByTag[tagId];

      return {
        tagId,
        color: PALETTE[index % PALETTE.length],
        label: series?.tagName ?? tagId,
        latestValue,
        unit,
        loading: Boolean(loadingByTag[tagId]),
        error
      };
    });
  }, [errorByTag, loadingByTag, selectedTagIds, seriesByTag]);

  if (rows.length === 0) {
    return (
      <Card withBorder p="sm">
        <Text c="dimmed" size="sm">
          Легенда появится после выбора тегов.
        </Text>
      </Card>
    );
  }

  return (
    <Stack gap="xs">
      {rows.map((row) => (
        <Card key={row.tagId} withBorder p="xs">
          <Group justify="space-between" wrap="nowrap" align="center">
            <Group gap="sm" wrap="nowrap" style={{ minWidth: 0 }}>
              <div
                style={{
                  width: 12,
                  height: 12,
                  borderRadius: 4,
                  backgroundColor: row.color,
                  flexShrink: 0
                }}
              />

              <Stack gap={0} style={{ minWidth: 0 }}>
                <Text size="sm" truncate>
                  {row.label}
                </Text>
                <Text size="xs" c="dimmed">
                  {row.latestValue} {row.unit}
                </Text>
              </Stack>
            </Group>

            <Group gap={6} wrap="nowrap">
              {row.loading ? <Badge variant="light">Загрузка</Badge> : null}

              {row.error ? (
                <Tooltip label={row.error} withArrow>
                  <Badge color="alarm-warn" variant="filled">
                    !
                  </Badge>
                </Tooltip>
              ) : null}

              <Button size="xs" variant="secondary" onClick={() => void reload(row.tagId)}>
                Обновить
              </Button>
              <Button size="xs" variant="ghost" onClick={() => removeTag(row.tagId)}>
                Убрать
              </Button>
            </Group>
          </Group>
        </Card>
      ))}
    </Stack>
  );
};
