import type { MonitoringObjectModel } from "../types";

export interface MonitoringObjectApiModel {
  id: string;
  name: string;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export type MonitoringObjectId = MonitoringObjectApiModel["id"];

export type { MonitoringObjectModel };

export interface MonitoringObjectPayload {
  name?: string;
  description?: string | null;
}
