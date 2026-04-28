import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

dayjs.extend(utc);

export type RangePresetKey = "5m" | "1h" | "6h" | "24h" | "7d" | "30d" | "custom";

export type TrendsRange =
  | { kind: "preset"; preset: Exclude<RangePresetKey, "custom"> }
  | { kind: "custom"; from: string; to: string };

interface PresetConfig {
  label: string;
  step: string;
  windowMs: number;
}

const MINUTE_MS = 60_000;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;

export const MAX_CUSTOM_RANGE_MS = 90 * DAY_MS;

const PRESET_CONFIG: Record<Exclude<RangePresetKey, "custom">, PresetConfig> = {
  "5m": {
    label: "5 минут",
    step: "1s",
    windowMs: 5 * MINUTE_MS
  },
  "1h": {
    label: "1 час",
    step: "5s",
    windowMs: HOUR_MS
  },
  "6h": {
    label: "6 часов",
    step: "30s",
    windowMs: 6 * HOUR_MS
  },
  "24h": {
    label: "24 часа",
    step: "1m",
    windowMs: DAY_MS
  },
  "7d": {
    label: "7 дней",
    step: "5m",
    windowMs: 7 * DAY_MS
  },
  "30d": {
    label: "30 дней",
    step: "30m",
    windowMs: 30 * DAY_MS
  }
};

export const RANGE_PRESET_OPTIONS: Array<{
  value: RangePresetKey;
  label: string;
}> = [
  { value: "5m", label: "5м" },
  { value: "1h", label: "1ч" },
  { value: "6h", label: "6ч" },
  { value: "24h", label: "24ч" },
  { value: "7d", label: "7д" },
  { value: "30d", label: "30д" },
  { value: "custom", label: "Свой" }
];

const STEP_CANDIDATES = [
  { step: "1s", ms: 1_000 },
  { step: "2s", ms: 2_000 },
  { step: "5s", ms: 5_000 },
  { step: "10s", ms: 10_000 },
  { step: "30s", ms: 30_000 },
  { step: "1m", ms: MINUTE_MS },
  { step: "2m", ms: 2 * MINUTE_MS },
  { step: "5m", ms: 5 * MINUTE_MS },
  { step: "10m", ms: 10 * MINUTE_MS },
  { step: "30m", ms: 30 * MINUTE_MS },
  { step: "1h", ms: HOUR_MS }
];

export interface ResolvedRange {
  from: string;
  to: string;
  step: string;
}

export const resolveRange = (range: TrendsRange, stepOverride?: string): ResolvedRange => {
  const now = dayjs.utc();

  if (range.kind === "preset") {
    const preset = PRESET_CONFIG[range.preset];
    return {
      from: now.subtract(preset.windowMs, "millisecond").toISOString(),
      to: now.toISOString(),
      step: stepOverride ?? preset.step
    };
  }

  return {
    from: range.from,
    to: range.to,
    step: stepOverride ?? deriveCustomStep(range.from, range.to)
  };
};

export const deriveCustomStep = (fromIso: string, toIso: string): string => {
  const fromMs = dayjs.utc(fromIso).valueOf();
  const toMs = dayjs.utc(toIso).valueOf();

  if (!Number.isFinite(fromMs) || !Number.isFinite(toMs) || toMs <= fromMs) {
    return "1m";
  }

  const targetBucketMs = Math.max(1_000, Math.ceil((toMs - fromMs) / 300));
  const picked =
    STEP_CANDIDATES.find((candidate) => candidate.ms >= targetBucketMs) ??
    STEP_CANDIDATES[STEP_CANDIDATES.length - 1];

  return picked.step;
};

export const getPresetLabel = (preset: Exclude<RangePresetKey, "custom">): string => {
  return PRESET_CONFIG[preset].label;
};

export const validateCustomRange = (fromIso: string, toIso: string): string | null => {
  const from = dayjs.utc(fromIso);
  const to = dayjs.utc(toIso);
  const now = dayjs.utc();

  if (!from.isValid() || !to.isValid()) {
    return "Выберите корректные дату и время";
  }

  if (!from.isBefore(to)) {
    return "Начало позже конца";
  }

  if (to.diff(now, "millisecond") > MINUTE_MS) {
    return "Конец периода не может быть в будущем";
  }

  if (to.diff(from, "millisecond") > MAX_CUSTOM_RANGE_MS) {
    return "Период не может превышать 90 дней";
  }

  return null;
};
