import { api } from "@shared/api";
import type { PaginatedResponse } from "@shared/api";

import type {
  CreateDevicePayload,
  Device,
  DeviceType,
  DeviceWithParams,
  UpdateDevicePayload
} from "./types";

interface ListDevicesParams {
  offset?: number;
  limit?: number;
  name?: string;
  type_id?: number;
}

export const devicesApi = {
  listByObject: (objectId: string, params?: ListDevicesParams) =>
    api.get<PaginatedResponse<Device>>(
      `/objects/${objectId}/devices`,
      params as Record<string, string | number | boolean | null | undefined> | undefined
    ),
  get: (id: string) => api.get<DeviceWithParams>(`/devices/${id}`),
  create: (objectId: string, payload: CreateDevicePayload) =>
    api.post<DeviceWithParams>(`/objects/${objectId}/devices`, payload),
  update: (id: string, payload: UpdateDevicePayload) =>
    api.patch<DeviceWithParams>(`/devices/${id}`, payload),
  delete: (id: string) => api.delete(`/devices/${id}`)
};

export const deviceTypesApi = {
  list: () => api.get<DeviceType[]>("/device-types")
};

