import { create } from "zustand";

import type { MonitoringObjectModel, MonitoringObjectsStore } from "./types";

export const ObjectStore = create<MonitoringObjectsStore<MonitoringObjectModel>>(() => ({
  objects: [],
  activeObjectId: null,
  isLoading: false,
  error: null,
  save: null as unknown as MonitoringObjectsStore<MonitoringObjectModel>["save"],
  read: null as unknown as MonitoringObjectsStore<MonitoringObjectModel>["read"],
  delete: null as unknown as MonitoringObjectsStore<MonitoringObjectModel>["delete"]
}));
