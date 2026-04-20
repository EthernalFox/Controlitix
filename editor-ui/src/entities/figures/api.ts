import { api } from "@shared/api";
import type { PaginatedResponse } from "@shared/api";

import type { CreateFigurePayload, Figure, UpdateFigurePayload } from "./types";

interface ListFiguresParams {
  offset?: number;
  limit?: number;
}

export const figuresApi = {
  list: (diagramId: string, params?: ListFiguresParams) =>
    api.get<PaginatedResponse<Figure>>(
      `/diagrams/${diagramId}/figures`,
      params as Record<string, string | number | boolean | null | undefined> | undefined
    ),
  create: (diagramId: string, figures: CreateFigurePayload[]) =>
    api.post<Figure[]>(`/diagrams/${diagramId}/figures`, figures),
  update: (figureId: string, payload: UpdateFigurePayload) =>
    api.patch<Figure>(`/figures/${figureId}`, payload),
  remove: (figureId: string) => api.delete(`/figures/${figureId}`)
};
