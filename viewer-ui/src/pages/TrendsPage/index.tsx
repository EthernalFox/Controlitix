import { lazy, Suspense, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router";

import { resolveRange, useTrendsStore } from "@/features/trends";
import { useAppColorScheme } from "@/shared/libs/theme";
import { useRealtimeStore, useTagSubscription } from "@/shared/modules/realtime";
import { Card, Group, Loader, Stack, Text, Title } from "@/shared/ui/components";
import type { TrendChartTheme } from "@/widgets/TrendChart";
import { TrendLegend } from "@/widgets/TrendLegend";
import { TrendRangePicker } from "@/widgets/TrendRangePicker";
import { TrendTagPicker } from "@/widgets/TrendTagPicker";

const TrendChart = lazy(() =>
  import("@/widgets/TrendChart").then((module) => ({
    default: module.TrendChart
  }))
);

const readCssVar = (name: string, fallback: string): string => {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return value || fallback;
};

const LIVE_TAIL_PRESETS = new Set(["5m", "1h", "6h", "24h"]);
const CONFIG_CHANGE_COALESCE_MS = 5000;

export const TrendsPage = () => {
  const { tagId } = useParams<{ tagId: string }>();
  const { colorScheme } = useAppColorScheme();

  const selectedTagIds = useTrendsStore((state) => state.selectedTagIds);
  const range = useTrendsStore((state) => state.range);
  const step = useTrendsStore((state) => state.step);
  const seriesByTag = useTrendsStore((state) => state.seriesByTag);
  const loadingByTag = useTrendsStore((state) => state.loadingByTag);
  const liveTailActive = useTrendsStore((state) => state.liveTailActive);
  const addTag = useTrendsStore((state) => state.addTag);
  const refresh = useTrendsStore((state) => state.refresh);
  const appendLivePoint = useTrendsStore((state) => state.appendLivePoint);
  const refreshMetadataByTagIds = useTrendsStore(
    (state) => state.refreshMetadataByTagIds
  );
  const setLiveTailActive = useTrendsStore((state) => state.setLiveTailActive);
  const lastConfigChange = useRealtimeStore((state) => state.lastConfigChange);

  const [chartsReady, setChartsReady] = useState(false);
  const setupStartedRef = useRef(false);
  const lastMetaRefreshAtRef = useRef(0);

  useEffect(() => {
    if (setupStartedRef.current) {
      return;
    }

    setupStartedRef.current = true;

    void import("@/shared/modules/charts")
      .then((module) => {
        module.setupCharts();
        setChartsReady(true);
      })
      .catch(() => {
        setChartsReady(true);
      });
  }, []);

  useEffect(() => {
    if (!tagId || selectedTagIds.includes(tagId)) {
      return;
    }

    addTag(tagId);
    void useTrendsStore.getState().refresh();
  }, [addTag, selectedTagIds, tagId]);

  useEffect(() => {
    if (selectedTagIds.length === 0) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      void refresh();
    }, 200);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [range, refresh, selectedTagIds, step]);

  const canUseLiveTail =
    range.kind === "preset" && LIVE_TAIL_PRESETS.has(range.preset) && selectedTagIds.length > 0;

  const hasLoadedSeries = useMemo(() => {
    if (!canUseLiveTail) {
      return false;
    }

    return selectedTagIds.every((id) => {
      return Boolean(seriesByTag[id]) && !loadingByTag[id];
    });
  }, [canUseLiveTail, loadingByTag, selectedTagIds, seriesByTag]);

  const liveTailEnabled = canUseLiveTail && hasLoadedSeries;

  useEffect(() => {
    setLiveTailActive(liveTailEnabled);
  }, [liveTailEnabled, setLiveTailActive]);

  useEffect(() => {
    return () => {
      setLiveTailActive(false);
    };
  }, [setLiveTailActive]);

  useTagSubscription(liveTailEnabled ? selectedTagIds : [], (message) => {
    if (message.isSnapshot) {
      return;
    }

    appendLivePoint(message.tag_id, {
      ts: message.ts,
      v: message.v,
      q: message.q
    });
  });

  useEffect(() => {
    if (!lastConfigChange || lastConfigChange.entity_type !== "tag") {
      return;
    }

    const changedTagID = lastConfigChange.entity_id.trim();
    if (!changedTagID || !selectedTagIds.includes(changedTagID)) {
      return;
    }

    const eventTimestamp = Date.parse(lastConfigChange.timestamp);
    if (
      Number.isFinite(eventTimestamp) &&
      eventTimestamp > 0 &&
      eventTimestamp <= lastMetaRefreshAtRef.current
    ) {
      return;
    }

    const now = Date.now();
    if (now - lastMetaRefreshAtRef.current < CONFIG_CHANGE_COALESCE_MS) {
      return;
    }

    lastMetaRefreshAtRef.current = now;
    void refreshMetadataByTagIds([changedTagID]);
  }, [lastConfigChange, refreshMetadataByTagIds, selectedTagIds]);

  const shouldAutoRefresh = canUseLiveTail && !liveTailActive;

  useEffect(() => {
    if (!shouldAutoRefresh) {
      return;
    }

    const intervalId = window.setInterval(() => {
      void refresh();
    }, 15_000);

    return () => {
      window.clearInterval(intervalId);
    };
  }, [refresh, shouldAutoRefresh]);

  const rangeWindow = useMemo(() => {
    return resolveRange(range, step);
  }, [range, step]);

  const series = useMemo(() => {
    return selectedTagIds
      .map((id) => seriesByTag[id])
      .filter((item): item is NonNullable<typeof item> => Boolean(item));
  }, [selectedTagIds, seriesByTag]);

  const isLoading = useMemo(() => {
    return selectedTagIds.some((tag) => Boolean(loadingByTag[tag]));
  }, [loadingByTag, selectedTagIds]);

  const chartPalette = useMemo(() => {
    const fallbackPrimary = colorScheme === "dark" ? "#659cff" : "#2372fe";
    const fallbackSecondary = colorScheme === "dark" ? "#7a8cdd" : "#445ccf";

    return [
      readCssVar("--mantine-color-primary-4", fallbackPrimary),
      readCssVar("--mantine-color-primary-6", "#0f68ff"),
      readCssVar("--mantine-color-primary-8", "#004fcd"),
      readCssVar("--mantine-color-secondary-4", fallbackSecondary),
      readCssVar("--mantine-color-secondary-6", "#3a51cd"),
      readCssVar("--mantine-color-secondary-8", "#22348f")
    ];
  }, [colorScheme]);

  const chartTheme = useMemo<TrendChartTheme>(() => {
    const textFallback = colorScheme === "dark" ? "#f8f9fa" : "#1f2933";
    const gridFallback = colorScheme === "dark" ? "#6d8594" : "#9aa9b3";
    const bodyFallback = colorScheme === "dark" ? "#111422" : "#ffffff";

    return {
      axisText: readCssVar("--mantine-color-text", textFallback),
      axisGrid: readCssVar("--mantine-color-dimmed", gridFallback),
      tooltipBackground: readCssVar("--mantine-color-body", bodyFallback),
      tooltipText: readCssVar("--mantine-color-text", textFallback)
    };
  }, [colorScheme]);

  return (
    <Stack gap="md">
      <Title order={2}>Тренды</Title>

      <Group align="flex-start" wrap="wrap" gap="md">
        <Card withBorder p="md" style={{ flex: "1 1 300px", maxWidth: 340 }}>
          <TrendTagPicker />
        </Card>

        <Stack gap="md" style={{ flex: "999 1 720px", minWidth: 320 }}>
          <Card withBorder p="md">
            <TrendRangePicker />
          </Card>

          {chartsReady ? (
            <Suspense
              fallback={
                <Card withBorder p="md" style={{ minHeight: 360 }}>
                  <Group justify="center" align="center" style={{ minHeight: 328 }}>
                    <Loader />
                  </Group>
                </Card>
              }
            >
              <TrendChart
                fromIso={rangeWindow.from}
                toIso={rangeWindow.to}
                series={series}
                palette={chartPalette}
                theme={chartTheme}
                loading={isLoading}
                hasSelectedTags={selectedTagIds.length > 0}
              />
            </Suspense>
          ) : (
            <Card withBorder p="md" style={{ minHeight: 360 }}>
              <Group justify="center" align="center" style={{ minHeight: 328 }}>
                <Stack align="center" gap="xs">
                  <Loader />
                  <Text c="dimmed">Инициализация графика...</Text>
                </Stack>
              </Group>
            </Card>
          )}

          <TrendLegend />
        </Stack>
      </Group>
    </Stack>
  );
};
