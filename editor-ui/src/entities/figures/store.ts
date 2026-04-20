import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { figuresApi } from "./api";
import type { CreateFigurePayload, Figure, UpdateFigurePayload } from "./types";

interface FiguresState {
  figures: Figure[];
  isLoading: boolean;
  error: string | null;
  fetchFigures: (diagramId: string) => Promise<void>;
  addFigure: (diagramId: string, payload: CreateFigurePayload) => Promise<Figure>;
  updateFigure: (figureId: string, payload: UpdateFigurePayload) => Promise<void>;
  patchFigureLocal: (figureId: string, payload: UpdateFigurePayload) => void;
  removeFigure: (figureId: string) => Promise<void>;
  reorderFigures: (orderedIds: string[]) => void;
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

const applyUpdatePayload = (figure: Figure, payload: UpdateFigurePayload): Figure => ({
  ...figure,
  type: payload.type ?? figure.type,
  params: payload.params ? { ...figure.params, ...payload.params } : figure.params,
  tagId: payload.tag_id === undefined ? figure.tagId : payload.tag_id,
  updatedAt: new Date().toISOString()
});

const upsertFigure = (figures: Figure[], figure: Figure) => {
  const targetIndex = figures.findIndex((item) => item.id === figure.id);
  if (targetIndex === -1) {
    return [...figures, figure];
  }

  return figures.map((item) => (item.id === figure.id ? figure : item));
};

export const useFiguresStore = create<FiguresState>((set, get) => ({
  figures: [],
  isLoading: false,
  error: null,

  fetchFigures: async (diagramId) => {
    set({ isLoading: true, error: null });

    try {
      const response = await figuresApi.list(diagramId, { offset: 0, limit: 1000 });
      set({ figures: response.items, isLoading: false });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  addFigure: async (diagramId, payload) => {
    set({ error: null });
    const createdFigures = await figuresApi.create(diagramId, [payload]);
    const createdFigure = createdFigures[0];

    set((state) => ({
      figures: createdFigure ? [...state.figures, createdFigure] : state.figures
    }));

    if (!createdFigure) {
      throw new Error("Figure was not created");
    }

    return createdFigure;
  },

  updateFigure: async (figureId, payload) => {
    set({ error: null });
    const previousFigure = get().figures.find((figure) => figure.id === figureId) ?? null;

    if (previousFigure) {
      set((state) => ({
        figures: state.figures.map((figure) =>
          figure.id === figureId ? applyUpdatePayload(figure, payload) : figure
        )
      }));
    }

    try {
      const updatedFigure = await figuresApi.update(figureId, payload);
      set((state) => ({
        figures: upsertFigure(state.figures, updatedFigure)
      }));
    } catch (error) {
      if (previousFigure) {
        set((state) => ({
          figures: upsertFigure(state.figures, previousFigure)
        }));
      }

      set({ error: toErrorMessage(error) });
      throw error;
    }
  },

  patchFigureLocal: (figureId, payload) => {
    set((state) => ({
      figures: state.figures.map((figure) =>
        figure.id === figureId ? applyUpdatePayload(figure, payload) : figure
      )
    }));
  },

  removeFigure: async (figureId) => {
    set({ error: null });
    await figuresApi.remove(figureId);

    set((state) => ({
      figures: state.figures.filter((figure) => figure.id !== figureId)
    }));
  },

  reorderFigures: (orderedIds) => {
    const figures = get().figures;
    const orderedMap = new Map(figures.map((figure) => [figure.id, figure]));
    const orderedFigures = orderedIds
      .map((id) => orderedMap.get(id))
      .filter((figure): figure is Figure => Boolean(figure));
    const restFigures = figures.filter((figure) => !orderedIds.includes(figure.id));

    set({ figures: [...orderedFigures, ...restFigures] });
  },

  reset: () => set({ figures: [], isLoading: false, error: null })
}));
