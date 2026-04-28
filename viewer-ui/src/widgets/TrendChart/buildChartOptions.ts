import type { ChartOptions } from "chart.js";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

import type { TrendSeries } from "@/shared/modules/charts";

import type { TrendChartPoint } from "./buildChartData";

dayjs.extend(utc);

export interface TrendChartTheme {
  axisText: string;
  axisGrid: string;
  tooltipBackground: string;
  tooltipText: string;
}

const qualityLabels: Record<string, string> = {
  ok: "OK",
  hi: "HI",
  hihi: "HIHI",
  uncertain: "Uncertain",
  bad: "Bad",
  comm_loss: "Comm Loss",
  offline: "Offline",
  acknowledged: "Acknowledged"
};

const resolveTimeUnit = (fromIso: string, toIso: string): "minute" | "hour" | "day" => {
  const from = dayjs.utc(fromIso);
  const to = dayjs.utc(toIso);
  const spanMinutes = to.diff(from, "minute");

  if (spanMinutes <= 180) {
    return "minute";
  }

  if (spanMinutes <= 72 * 60) {
    return "hour";
  }

  return "day";
};

const formatValue = (value: number | null): string => {
  if (value === null) {
    return "—";
  }

  return value.toFixed(3);
};

const resolveYAxisTitle = (series: TrendSeries[]): string => {
  if (series.length !== 1) {
    return "Value";
  }

  const singleSeries = series[0];
  const unit = singleSeries.unit.symbol || singleSeries.unit.name;

  return `${singleSeries.tagName} (${unit})`;
};

export const buildChartOptions = (
  fromIso: string,
  toIso: string,
  series: TrendSeries[],
  theme: TrendChartTheme
): ChartOptions<"line"> => {
  const timeUnit = resolveTimeUnit(fromIso, toIso);

  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    normalized: true,
    parsing: false,
    onClick: (event, _elements, chart) => {
      const nativeEvent = event.native as MouseEvent | null;

      if (nativeEvent?.detail === 2 && typeof chart.resetZoom === "function") {
        chart.resetZoom();
      }
    },
    interaction: {
      mode: "nearest",
      intersect: false
    },
    scales: {
      x: {
        type: "time",
        time: {
          unit: timeUnit,
          minUnit: "second",
          displayFormats: {
            minute: "DD.MM HH:mm",
            hour: "DD.MM HH:mm",
            day: "DD.MM.YYYY"
          }
        },
        ticks: {
          color: theme.axisText
        },
        grid: {
          color: theme.axisGrid
        }
      },
      y: {
        type: "linear",
        ticks: {
          color: theme.axisText
        },
        grid: {
          color: theme.axisGrid
        },
        title: {
          display: true,
          text: resolveYAxisTitle(series),
          color: theme.axisText
        }
      }
    },
    plugins: {
      legend: {
        display: false
      },
      tooltip: {
        displayColors: true,
        backgroundColor: theme.tooltipBackground,
        titleColor: theme.tooltipText,
        bodyColor: theme.tooltipText,
        footerColor: theme.tooltipText,
        bodyFont: {
          family: "JetBrains Mono, monospace"
        },
        callbacks: {
          title: (items) => {
            const point = items[0];
            return dayjs(point.parsed.x).format("DD.MM.YYYY HH:mm:ss");
          },
          label: (context) => {
            const raw = context.raw as TrendChartPoint;
            return `${context.dataset.label}: ${formatValue(raw.y)}`;
          },
          afterLabel: (context) => {
            const raw = context.raw as TrendChartPoint;
            return `Качество: ${qualityLabels[raw.q] ?? raw.q}`;
          }
        }
      },
      decimation: {
        enabled: true,
        algorithm: "lttb",
        samples: 500
      },
      zoom: {
        pan: {
          enabled: true,
          mode: "x",
          threshold: 5
        },
        zoom: {
          wheel: {
            enabled: true
          },
          pinch: {
            enabled: true
          },
          drag: {
            enabled: true,
            backgroundColor: "rgba(60, 130, 254, 0.2)"
          },
          mode: "x"
        }
      }
    }
  };
};
