import type { PaginatedResponse } from "@shared/api";
import { api } from "@shared/api";

import type {
  CreateObjectPayload,
  MonitoringObject,
  UpdateObjectPayload
} from "./types";

interface ListObjectsParams {
  offset?: number;
  limit?: number;
  search?: string;
}

export const objectsApi = {
  list: (params?: ListObjectsParams) =>
    api.get<PaginatedResponse<MonitoringObject>>(
      "/objects",
      params as Record<string, string | number | boolean | null | undefined> | undefined
    ),
  get: (id: string) => api.get<MonitoringObject>(`/objects/${id}`),
  create: (payload: CreateObjectPayload) =>
    api.post<MonitoringObject>("/objects", payload),
  update: (id: string, payload: UpdateObjectPayload) =>
    api.patch<MonitoringObject>(`/objects/${id}`, payload),
  delete: (id: string) => api.delete(`/objects/${id}`)
};
