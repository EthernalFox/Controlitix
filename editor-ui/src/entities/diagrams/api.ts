import { api } from "@shared/api";
import type { PaginatedResponse } from "@shared/api";

import type { CreateDiagramPayload, Diagram } from "./types";

interface ListDiagramsParams {
  offset?: number;
  limit?: number;
}

export const diagramsApi = {
  listByObject: (objectId: string, params?: ListDiagramsParams) =>
    api.get<PaginatedResponse<Diagram>>(
      `/objects/${objectId}/diagrams`,
      params as Record<string, string | number | boolean | null | undefined> | undefined
    ),
  get: (id: string) => api.get<Diagram>(`/diagrams/${id}`),
  create: (objectId: string, payload: CreateDiagramPayload) =>
    api.post<Diagram>(`/objects/${objectId}/diagrams`, payload),
  publish: (id: string) => api.post<Diagram>(`/diagrams/${id}/publish`, {}),
  delete: (id: string) => api.delete(`/diagrams/${id}`)
};
