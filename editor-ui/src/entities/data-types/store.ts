import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { dataTypesApi } from "./api";
import type { DataType } from "./types";

interface DataTypesState {
  items: DataType[];
  isLoading: boolean;
  isLoaded: boolean;
  error: string | null;
  fetch: () => Promise<void>;
  reset: () => void;
}

const toErrorMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Request failed";
};

export const useDataTypesStore = create<DataTypesState>((set, get) => ({
  items: [],
  isLoading: false,
  isLoaded: false,
  error: null,

  fetch: async () => {
    if (get().isLoaded || get().isLoading) {
      return;
    }

    set({ isLoading: true, error: null });

    try {
      const items = await dataTypesApi.list();
      set({ items, isLoading: false, isLoaded: true });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  reset: () => {
    set({
      items: [],
      isLoading: false,
      isLoaded: false,
      error: null
    });
  }
}));
