import { api } from "@shared/api";
import type { PaginatedResponse } from "@shared/api";

import type {
  CreateTagPayload,
  Tag,
  TagScalingPayload,
  TagsListFilters,
  TagSetpoints,
  UpdateTagParamsPayload,
  UpdateTagPayload
} from "./types";

export const tagsApi = {
  list: (filters: TagsListFilters) =>
    api.get<PaginatedResponse<Tag>>(
      "/tags",
      filters as Record<string, string | number | boolean | null | undefined>
    ),
  get: (id: string) => api.get<Tag>(`/tags/${id}`),
  create: (deviceId: string, payload: CreateTagPayload) =>
    api.post<Tag>(`/devices/${deviceId}/tags`, payload),
  update: (id: string, payload: UpdateTagPayload) =>
    api.patch<Tag>(`/tags/${id}`, payload),
  updateParams: (id: string, payload: UpdateTagParamsPayload) =>
    api.put<Tag>(`/tags/${id}/params`, payload),
  putSetpoints: (id: string, payload: TagSetpoints) =>
    api.put<Tag>(`/tags/${id}/setpoints`, payload),
  deleteSetpoints: (id: string) => api.delete(`/tags/${id}/setpoints`),
  putScaling: (id: string, payload: TagScalingPayload) =>
    api.put<Tag>(`/tags/${id}/scaling`, payload),
  deleteScaling: (id: string) => api.delete(`/tags/${id}/scaling`),
  delete: (id: string) => api.delete(`/tags/${id}`)
};
