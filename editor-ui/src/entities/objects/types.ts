export interface MonitoringObjectModel {
  id: string;
  name: string;
  description: string | null;
  createdAt: string;
  updatedAt: string;
}

export type SaveMonitoringObjectPayload<
  T extends MonitoringObjectModel = MonitoringObjectModel
> = Partial<Omit<T, "createdAt" | "updatedAt">> & { id?: T["id"] };

export interface MonitoringObjectsStore<
  T extends MonitoringObjectModel = MonitoringObjectModel
> {
  objects: T[];
  activeObjectId: T["id"] | null;
  isLoading: boolean;
  error: string | null;
  save: (payload: SaveMonitoringObjectPayload<T>) => Promise<T>;
  read: () => Promise<T[]>;
  delete: (id: T["id"]) => Promise<void>;
}
