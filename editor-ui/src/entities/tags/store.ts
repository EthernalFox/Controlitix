import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { tagsApi } from "./api";
import type {
  CreateTagPayload,
  Tag,
  TagEditDiff,
  TagsListFilters
} from "./types";

interface TagsState {
  tags: Tag[];
  total: number;
  isLoading: boolean;
  error: string | null;
  filters: TagsListFilters;
  setFilters: (filters: Partial<TagsListFilters>) => void;
  fetchTags: () => Promise<void>;
  getTag: (id: string) => Promise<Tag>;
  createTag: (deviceId: string, payload: CreateTagPayload) => Promise<Tag>;
  saveTagEdits: (id: string, diff: TagEditDiff) => Promise<Tag>;
  deleteTag: (id: string) => Promise<void>;
  reset: () => void;
}

const DEFAULT_FILTERS: TagsListFilters = {
  offset: 0,
  limit: 50
};

const toErrorMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Request failed";
};

const upsertTag = (tags: Tag[], tag: Tag) => {
  const existingTag = tags.some((item) => item.id === tag.id);
  if (!existingTag) {
    return [tag, ...tags];
  }

  return tags.map((item) => (item.id === tag.id ? tag : item));
};

const maybeSetOffset = (
  currentFilters: TagsListFilters,
  partialFilters: Partial<TagsListFilters>
): number => {
  if (partialFilters.offset !== undefined) {
    return partialFilters.offset;
  }

  const hasAnyFilterChange = Object.keys(partialFilters).some((key) => key !== "offset");
  if (hasAnyFilterChange) {
    return 0;
  }

  return currentFilters.offset ?? 0;
};

export const useTagsStore = create<TagsState>((set, get) => ({
  tags: [],
  total: 0,
  isLoading: false,
  error: null,
  filters: DEFAULT_FILTERS,

  setFilters: (partialFilters) => {
    set((state) => ({
      filters: {
        ...state.filters,
        ...partialFilters,
        offset: maybeSetOffset(state.filters, partialFilters)
      }
    }));
  },

  fetchTags: async () => {
    const { filters } = get();
    set({ isLoading: true, error: null });

    try {
      const response = await tagsApi.list(filters);
      const detailedTags = await Promise.all(
        response.items.map(async (tag) => {
          try {
            return await tagsApi.get(tag.id);
          } catch {
            return tag;
          }
        })
      );

      set({
        tags: detailedTags,
        total: response.total,
        isLoading: false
      });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  getTag: async (id) => {
    const tag = await tagsApi.get(id);
    set((state) => ({ tags: upsertTag(state.tags, tag) }));
    return tag;
  },

  createTag: async (deviceId, payload) => {
    set({ error: null });
    const createdTag = await tagsApi.create(deviceId, payload);

    set((state) => ({
      tags: [createdTag, ...state.tags],
      total: state.total + 1
    }));

    return createdTag;
  },

  saveTagEdits: async (id, diff) => {
    set({ error: null });

    if (diff.meta) {
      await tagsApi.update(id, diff.meta);
    }

    if (diff.params) {
      await tagsApi.updateParams(id, diff.params);
    }

    if (diff.setpoints === "delete") {
      await tagsApi.deleteSetpoints(id);
    } else if (diff.setpoints) {
      await tagsApi.putSetpoints(id, diff.setpoints);
    }

    if (diff.scaling === "delete") {
      await tagsApi.deleteScaling(id);
    } else if (diff.scaling) {
      await tagsApi.putScaling(id, diff.scaling);
    }

    const updatedTag = await tagsApi.get(id);
    set((state) => ({ tags: upsertTag(state.tags, updatedTag) }));

    return updatedTag;
  },

  deleteTag: async (id) => {
    set({ error: null });
    await tagsApi.delete(id);

    set((state) => ({
      tags: state.tags.filter((tag) => tag.id !== id),
      total: Math.max(0, state.total - 1)
    }));
  },

  reset: () => {
    set({
      tags: [],
      total: 0,
      isLoading: false,
      error: null,
      filters: DEFAULT_FILTERS
    });
  }
}));
