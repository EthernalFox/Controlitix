import { createFormatter } from "@shared/modules/api";

import type { MonitoringObjectApiModel, MonitoringObjectModel } from "./types";

export const monitoringObjectsFormatter = createFormatter<
  MonitoringObjectApiModel,
  MonitoringObjectModel
>((raw) => ({
  id: raw.id,
  name: raw.name,
  description: raw.description,
  createdAt: raw.created_at,
  updatedAt: raw.updated_at
}));
