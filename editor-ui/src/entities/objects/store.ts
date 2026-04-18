import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { objectsApi } from "./api";
import type {
  CreateObjectPayload,
  MonitoringObject,
  UpdateObjectPayload
} from "./types";

interface ObjectsState {
  objects: MonitoringObject[];
  isLoading: boolean;
  error: string | null;
  fetchObjects: (search?: string) => Promise<void>;
  getObject: (id: string) => Promise<MonitoringObject>;
  createObject: (payload: CreateObjectPayload) => Promise<MonitoringObject>;
  updateObject: (id: string, payload: UpdateObjectPayload) => Promise<void>;
  deleteObject: (id: string) => Promise<void>;
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

export const useObjectsStore = create<ObjectsState>((set, get) => ({
  objects: [],
  isLoading: false,
  error: null,

  fetchObjects: async (search) => {
    set({ isLoading: true, error: null });

    try {
      const response = await objectsApi.list({ search });
      set({ objects: response.items, isLoading: false });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  getObject: async (id) => {
    const existingObject = get().objects.find((object) => object.id === id);
    if (existingObject) {
      return existingObject;
    }

    const object = await objectsApi.get(id);
    set((state) => {
      const alreadyExists = state.objects.some(
        (currentObject) => currentObject.id === object.id
      );
      if (alreadyExists) {
        return state;
      }

      return { objects: [...state.objects, object] };
    });

    return object;
  },

  createObject: async (payload) => {
    set({ error: null });
    const object = await objectsApi.create(payload);

    set((state) => ({
      objects: [object, ...state.objects]
    }));

    return object;
  },

  updateObject: async (id, payload) => {
    set({ error: null });
    const updatedObject = await objectsApi.update(id, payload);

    set((state) => ({
      objects: state.objects.map((object) =>
        object.id === id ? updatedObject : object
      )
    }));
  },

  deleteObject: async (id) => {
    set({ error: null });
    await objectsApi.delete(id);

    set((state) => ({
      objects: state.objects.filter((object) => object.id !== id)
    }));
  },

  reset: () => {
    set({
      objects: [],
      isLoading: false,
      error: null
    });
  }
}));
