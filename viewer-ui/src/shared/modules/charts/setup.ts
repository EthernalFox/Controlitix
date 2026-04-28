import {
  Chart as ChartJS,
  Decimation,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  TimeScale,
  Tooltip
} from "chart.js";
import zoomPlugin from "chartjs-plugin-zoom";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

import "chartjs-adapter-dayjs-4";

let isRegistered = false;

export const setupCharts = (): void => {
  if (isRegistered) {
    return;
  }

  dayjs.extend(utc);

  ChartJS.register(
    LineController,
    LineElement,
    PointElement,
    LinearScale,
    TimeScale,
    Tooltip,
    Legend,
    Filler,
    Decimation,
    zoomPlugin
  );

  isRegistered = true;
};
