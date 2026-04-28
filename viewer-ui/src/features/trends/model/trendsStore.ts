import { create } from "zustand";

import {
  resolveRange,
  type RangePresetKey,
  type TrendsRange
} from "@/features/trends/lib/rangePresets";
import {
  fetchTrend,
  fetchTrendsBatch,
  type FetchTrendParams
} from "@/features/trends/model/trendsApi";
import { ApiRequestError } from "@/shared/api";
import type { TrendPoint, TrendSeries } from "@/shared/modules/charts";

interface TrendsState {
  selectedTagIds: string[];
  range: TrendsRange;
  step?: string;
  seriesByTag: Record<string, TrendSeries | undefined>;
  loadingByTag: Record<string, boolean>;
  errorByTag: Record<string, string | undefined>;
  liveTailActive: boolean;

  addTag: (tagId: string) => void;
  removeTag: (tagId: string) => void;
  setRange: (range: TrendsState["range"]) => void;
  setLiveTailActive: (active: boolean) => void;
  appendLivePoint: (tagId: string, point: TrendPoint) => void;
  refreshMetadataByTagIds: (tagIds: string[]) => Promise<void>;
  refresh: () => Promise<void>;
  reload: (tagId: string) => Promise<void>;
}

const MAX_SELECTED_TAGS = 6;

const defaultRange: TrendsRange = {
  kind: "preset",
  preset: "1h"
};

const resolveMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.title;
  }

  return "Не удалось загрузить тренд";
};

const buildFetchParams = (state: TrendsState): FetchTrendParams => {
  const resolvedRange = resolveRange(state.range, state.step);
  return {
    from: resolvedRange.from,
    to: resolvedRange.to,
    step: resolvedRange.step,
    agg: "avg",
    limit: 5_000
  };
};

export const useTrendsStore = create<TrendsState>((set, get) => ({
  selectedTagIds: [],
  range: defaultRange,
  step: undefined,
  seriesByTag: {},
  loadingByTag: {},
  errorByTag: {},
  liveTailActive: false,

  addTag: (tagId) => {
    const selectedTagIds = get().selectedTagIds;

    if (selectedTagIds.includes(tagId) || selectedTagIds.length >= MAX_SELECTED_TAGS) {
      return;
    }

    set({
      selectedTagIds: [...selectedTagIds, tagId]
    });
  },

  removeTag: (tagId) => {
    set((state) => {
      const selectedTagIds = state.selectedTagIds.filter((id) => id !== tagId);
      const seriesByTag = { ...state.seriesByTag };
      const loadingByTag = { ...state.loadingByTag };
      const errorByTag = { ...state.errorByTag };

      delete seriesByTag[tagId];
      delete loadingByTag[tagId];
      delete errorByTag[tagId];

      return {
        selectedTagIds,
        seriesByTag,
        loadingByTag,
        errorByTag
      };
    });
  },

  setRange: (range) => {
    set({ range });
  },

  setLiveTailActive: (active) => {
    set({ liveTailActive: active });
  },

  appendLivePoint: (tagId, point) => {
    set((state) => {
      const series = state.seriesByTag[tagId];
      if (!series) {
        return {};
      }

      const rangeWindow = resolveRange(state.range, state.step);
      const minTs = Date.parse(rangeWindow.from);
      const pointTs = Date.parse(point.ts);

      if (!Number.isFinite(pointTs) || pointTs < minTs) {
        return {};
      }

      const nextPoints = [...series.points, point].filter((item) => {
        const ts = Date.parse(item.ts);
        return Number.isFinite(ts) && ts >= minTs;
      });

      return {
        seriesByTag: {
          ...state.seriesByTag,
          [tagId]: {
            ...series,
            from: rangeWindow.from,
            to: point.ts,
            points: nextPoints
          }
        }
      };
    });
  },

  refreshMetadataByTagIds: async (tagIds) => {
    const uniqueTagIDs = Array.from(
      new Set(
        tagIds
          .map((tagId) => tagId.trim())
          .filter((tagId) => tagId.length > 0)
      )
    );

    if (uniqueTagIDs.length === 0) {
      return;
    }

    const now = new Date();
    const minuteAgo = new Date(now.getTime() - 60_000);

    try {
      const response = await fetchTrendsBatch(uniqueTagIDs, {
        from: minuteAgo.toISOString(),
        to: now.toISOString(),
        agg: "last",
        limit: 1
      });
      const seriesByTagID = new Map(response.series.map((series) => [series.tagId, series]));

      set((state) => {
        const nextSeriesByTag = { ...state.seriesByTag };

        uniqueTagIDs.forEach((tagID) => {
          const currentSeries = nextSeriesByTag[tagID];
          const metadataSource = seriesByTagID.get(tagID);
          if (!currentSeries || !metadataSource) {
            return;
          }

          nextSeriesByTag[tagID] = {
            ...currentSeries,
            tagName: metadataSource.tagName,
            deviceId: metadataSource.deviceId,
            deviceName: metadataSource.deviceName,
            unit: metadataSource.unit,
            dataType: metadataSource.dataType
          };
        });

        return { seriesByTag: nextSeriesByTag };
      });
    } catch {
      // Metadata refresh is best-effort and should not break the current chart view.
    }
  },

  refresh: async () => {
    const state = get();
    const tagIds = state.selectedTagIds;

    if (tagIds.length === 0) {
      return;
    }

    const loadingByTag: Record<string, boolean> = {};
    const errorByTag: Record<string, string | undefined> = { ...state.errorByTag };

    tagIds.forEach((tagId) => {
      loadingByTag[tagId] = true;
      errorByTag[tagId] = undefined;
    });

    set({
      loadingByTag: { ...state.loadingByTag, ...loadingByTag },
      errorByTag
    });

    try {
      const params = buildFetchParams(get());
      const response = await fetchTrendsBatch(tagIds, params);
      const receivedByTag = new Map(response.series.map((series) => [series.tagId, series]));

      set((currentState) => {
        const nextSeriesByTag = { ...currentState.seriesByTag };
        const nextLoadingByTag = { ...currentState.loadingByTag };
        const nextErrorByTag = { ...currentState.errorByTag };

        tagIds.forEach((tagId) => {
          nextLoadingByTag[tagId] = false;
          const currentSeries = receivedByTag.get(tagId);

          if (!currentSeries) {
            nextErrorByTag[tagId] = "Не удалось загрузить тренд";
            return;
          }

          nextSeriesByTag[tagId] = currentSeries;
          nextErrorByTag[tagId] =
            currentSeries.points.length === 0 ? "Нет данных за выбранный период" : undefined;
        });

        return {
          seriesByTag: nextSeriesByTag,
          loadingByTag: nextLoadingByTag,
          errorByTag: nextErrorByTag
        };
      });
    } catch (error) {
      const message = resolveMessage(error);

      set((currentState) => {
        const nextLoadingByTag = { ...currentState.loadingByTag };
        const nextErrorByTag = { ...currentState.errorByTag };

        tagIds.forEach((tagId) => {
          nextLoadingByTag[tagId] = false;
          nextErrorByTag[tagId] = message;
        });

        return {
          loadingByTag: nextLoadingByTag,
          errorByTag: nextErrorByTag
        };
      });
    }
  },

  reload: async (tagId) => {
    if (!get().selectedTagIds.includes(tagId)) {
      return;
    }

    set((state) => ({
      loadingByTag: {
        ...state.loadingByTag,
        [tagId]: true
      },
      errorByTag: {
        ...state.errorByTag,
        [tagId]: undefined
      }
    }));

    try {
      const params = buildFetchParams(get());
      const series = await fetchTrend(tagId, params);

      set((state) => ({
        seriesByTag: {
          ...state.seriesByTag,
          [tagId]: series
        },
        loadingByTag: {
          ...state.loadingByTag,
          [tagId]: false
        },
        errorByTag: {
          ...state.errorByTag,
          [tagId]: series.points.length === 0 ? "Нет данных за выбранный период" : undefined
        }
      }));
    } catch (error) {
      set((state) => ({
        loadingByTag: {
          ...state.loadingByTag,
          [tagId]: false
        },
        errorByTag: {
          ...state.errorByTag,
          [tagId]: resolveMessage(error)
        }
      }));
    }
  }
}));

export type { RangePresetKey, TrendsRange };
