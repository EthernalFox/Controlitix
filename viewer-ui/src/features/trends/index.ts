export { decimateTrendPoints } from "./lib/decimate";
export {
  RANGE_PRESET_OPTIONS,
  deriveCustomStep,
  getPresetLabel,
  resolveRange,
  validateCustomRange
} from "./lib/rangePresets";
export { fetchTrend, fetchTrendTags, fetchTrendsBatch } from "./model/trendsApi";
export type { FetchTrendParams, TrendTagOption } from "./model/trendsApi";
export { useTrendsStore } from "./model/trendsStore";
export type { RangePresetKey, TrendsRange } from "./model/trendsStore";
