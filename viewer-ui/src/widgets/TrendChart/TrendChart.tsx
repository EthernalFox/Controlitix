import { useEffect, useMemo, useState } from "react";
import { Line } from "react-chartjs-2";

import type { TrendSeries } from "@/shared/modules/charts";
import { Card, Group, Loader, Stack, Text } from "@/shared/ui/components";

import { buildChartData } from "./buildChartData";
import { buildChartOptions, type TrendChartTheme } from "./buildChartOptions";

const DEFAULT_PALETTE = [
  "#659cff",
  "#0f68ff",
  "#004fcd",
  "#7a8cdd",
  "#3a51cd",
  "#22348f"
];

interface TrendChartProps {
  fromIso: string;
  toIso: string;
  series: TrendSeries[];
  palette?: string[];
  theme: TrendChartTheme;
  loading: boolean;
  hasSelectedTags: boolean;
}

const resolveChartHeight = (): number => {
  return window.innerHeight > 900 ? 480 : 360;
};

export const TrendChart = ({
  fromIso,
  toIso,
  series,
  palette = DEFAULT_PALETTE,
  theme,
  loading,
  hasSelectedTags
}: TrendChartProps) => {
  const [height, setHeight] = useState(resolveChartHeight);

  useEffect(() => {
    const onResize = () => {
      setHeight(resolveChartHeight());
    };

    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("resize", onResize);
    };
  }, []);

  const data = useMemo(() => {
    return buildChartData(series, palette);
  }, [palette, series]);

  const options = useMemo(() => {
    return buildChartOptions(fromIso, toIso, series, theme);
  }, [fromIso, series, theme, toIso]);

  if (!hasSelectedTags) {
    return (
      <Card withBorder p="md" style={{ minHeight: height }}>
        <Group justify="center" align="center" style={{ minHeight: height - 32 }}>
          <Text c="dimmed">Выберите тег слева, чтобы построить график</Text>
        </Group>
      </Card>
    );
  }

  if (loading && series.length === 0) {
    return (
      <Card withBorder p="md" style={{ minHeight: height }}>
        <Group justify="center" align="center" style={{ minHeight: height - 32 }}>
          <Stack align="center" gap="xs">
            <Loader />
            <Text c="dimmed">Загружаем тренд...</Text>
          </Stack>
        </Group>
      </Card>
    );
  }

  return (
    <Card withBorder p="xs" style={{ height }}>
      <Line data={data} options={options} />
    </Card>
  );
};
