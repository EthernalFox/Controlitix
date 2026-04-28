import type { Quality } from "@/shared/modules/charts";

const QUALITY_COLORS: Record<Quality, string> = {
  ok: "#4CAF50",
  hi: "#FFA726",
  hihi: "#EF5350",
  uncertain: "#42A5F5",
  bad: "#78909C",
  comm_loss: "#546E7A",
  offline: "#37474F",
  acknowledged: "#7E57C2"
};

export const qualityToColor = (quality: Quality): string => {
  return QUALITY_COLORS[quality];
};
