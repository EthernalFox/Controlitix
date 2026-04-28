import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

import { api } from "@/shared/api";
import type { Quality, TrendSeries } from "@/shared/modules/charts";

dayjs.extend(utc);

type Agg = "last" | "avg" | "min" | "max";

export interface FetchTrendParams {
  from: string;
  to: string;
  step?: string;
  agg?: Agg;
  limit?: number;
}

export interface TrendTagOption {
  id: string;
  name: string;
}

interface ListResponse<T> {
  items: T[];
  total: number;
  offset: number;
  limit: number;
}

interface TagListItem {
  id: string;
  name: string;
}

const MAX_TAG_OPTIONS = 20;
const DEFAULT_LIMIT = 1_000;

const isMockEnabled = import.meta.env.DEV && import.meta.env.VITE_TRENDS_MOCK === "true";

const MOCK_TAGS: TrendTagOption[] = [
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30401", name: "boiler_1.t_out" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30402", name: "boiler_1.t_in" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30403", name: "boiler_1.pressure" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30404", name: "boiler_1.flow" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30405", name: "boiler_2.t_out" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30406", name: "boiler_2.t_in" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30407", name: "boiler_2.pressure" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30408", name: "boiler_2.flow" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30409", name: "pump_1.rpm" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30410", name: "pump_1.current" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30411", name: "pump_1.vibration" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30412", name: "pump_2.rpm" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30413", name: "pump_2.current" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30414", name: "pump_2.vibration" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30415", name: "tank_1.level" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30416", name: "tank_1.temperature" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30417", name: "valve_1.position" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30418", name: "valve_2.position" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30419", name: "header_1.pressure" },
  { id: "37e5f24f-b987-4ee9-a2ac-0a8d11f30420", name: "header_1.temperature" }
];

const parseDurationMs = (duration: string): number => {
  const match = /^(\d+)(s|m|h|d)$/i.exec(duration.trim());
  if (!match) {
    return 1_000;
  }

  const value = Number.parseInt(match[1], 10);
  const unit = match[2].toLowerCase();

  if (unit === "s") {
    return value * 1_000;
  }

  if (unit === "m") {
    return value * 60_000;
  }

  if (unit === "h") {
    return value * 3_600_000;
  }

  return value * 86_400_000;
};

const generateMockSeries = (tagId: string, params: FetchTrendParams): TrendSeries => {
  const tag =
    MOCK_TAGS.find((candidate) => candidate.id === tagId) ??
    ({
      id: tagId,
      name: `tag_${tagId.slice(0, 8)}`
    } satisfies TrendTagOption);

  const from = dayjs.utc(params.from);
  const to = dayjs.utc(params.to);
  const step = params.step ?? "10s";
  const agg = params.agg ?? "avg";
  const limit = Math.min(params.limit ?? DEFAULT_LIMIT, 5_000);
  const source = to.diff(from, "hour", true) > 24 ? "agg_1m" : "raw";
  const stepMs = parseDurationMs(step);
  const pointCount = Math.max(
    1,
    Math.min(limit, Math.floor(to.diff(from, "millisecond") / stepMs))
  );
  const startPhase =
    (Math.abs(tagId.split("").reduce((acc, char) => acc + char.charCodeAt(0), 0)) %
      360) *
    (Math.PI / 180);

  const points = Array.from({ length: pointCount + 1 }).map((_, index) => {
    const ts = from.add(index * stepMs, "millisecond");
    const phase = startPhase + index / 12;
    const noise = Math.sin(index / 5 + startPhase) * 0.2 + Math.cos(index / 17) * 0.1;
    const baseValue = 60 + Math.sin(phase) * 12 + noise * 8;

    if (index > 0 && index % 37 === 0) {
      return {
        ts: ts.toISOString(),
        v: null,
        q: "comm_loss" as const
      };
    }

    const quality: Quality = index % 41 === 0 ? "uncertain" : "ok";

    return {
      ts: ts.toISOString(),
      v: Number(baseValue.toFixed(3)),
      q: quality
    };
  });

  return {
    tagId: tag.id,
    tagName: tag.name,
    deviceId: "00000000-0000-0000-0000-000000000001",
    deviceName: "Mock Device",
    unit: {
      id: 7,
      name: "\u00B0C",
      symbol: "\u00B0C",
      category: "temperature"
    },
    dataType: {
      id: 2,
      name: "float32"
    },
    from: from.toISOString(),
    to: to.toISOString(),
    step,
    agg,
    source,
    points
  };
};

const toQueryParams = (params: FetchTrendParams): Record<string, string | number> => {
  const query: Record<string, string | number> = {
    from: params.from,
    to: params.to,
    agg: params.agg ?? "avg",
    limit: params.limit ?? DEFAULT_LIMIT
  };

  if (params.step) {
    query.step = params.step;
  }

  return query;
};

export const fetchTrend = async (
  tagId: string,
  params: FetchTrendParams
): Promise<TrendSeries> => {
  if (isMockEnabled) {
    return generateMockSeries(tagId, params);
  }

  return api.get<TrendSeries>(`/trends/${tagId}`, toQueryParams(params));
};

export const fetchTrendsBatch = async (
  tagIds: string[],
  params: FetchTrendParams
): Promise<{ series: TrendSeries[] }> => {
  if (isMockEnabled) {
    return {
      series: tagIds.map((id) => generateMockSeries(id, params))
    };
  }

  return api.get<{ series: TrendSeries[] }>("/trends", {
    ...toQueryParams(params),
    tag_ids: tagIds.join(",")
  });
};

export const fetchTrendTags = async (
  search: string,
  limit = MAX_TAG_OPTIONS
): Promise<TrendTagOption[]> => {
  const normalizedSearch = search.trim().toLowerCase();
  const safeLimit = Math.max(1, Math.min(MAX_TAG_OPTIONS, limit));

  if (isMockEnabled) {
    return MOCK_TAGS.filter((tag) => tag.name.toLowerCase().includes(normalizedSearch)).slice(
      0,
      safeLimit
    );
  }

  const response = await api.get<ListResponse<TagListItem>>("/tags", {
    search,
    limit: safeLimit
  });

  return response.items.map((tag) => ({
    id: tag.id,
    name: tag.name
  }));
};
