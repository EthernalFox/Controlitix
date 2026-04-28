import type { ChartData } from "chart.js";

import { decimateTrendPoints } from "@/features/trends";
import type { Quality, TrendSeries } from "@/shared/modules/charts";

export interface TrendChartPoint {
  x: number;
  y: number | null;
  q: Quality;
}

const DEFAULT_PALETTE = [
  "#659cff",
  "#0f68ff",
  "#004fcd",
  "#7a8cdd",
  "#3a51cd",
  "#22348f"
];

export const buildChartData = (
  series: TrendSeries[],
  palette: string[] = DEFAULT_PALETTE
): ChartData<"line", TrendChartPoint[]> => {
  return {
    datasets: series.map((item, index) => {
      const color = palette[index % palette.length];
      const points = decimateTrendPoints(item.points, 500).map((point) => ({
        x: Date.parse(point.ts),
        y: point.v,
        q: point.q
      }));

      return {
        label: `${item.tagName} (${item.unit.symbol || item.unit.name})`,
        data: points,
        borderColor: color,
        backgroundColor: color,
        pointBackgroundColor: color,
        pointHoverRadius: 4,
        pointRadius: 0,
        borderWidth: 2,
        spanGaps: false,
        tension: 0.2
      };
    })
  };
};
