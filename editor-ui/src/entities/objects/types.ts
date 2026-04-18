export interface MonitoringObject {
  id: string;
  name: string;
  description: string | null;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
}

export interface CreateObjectPayload {
  name: string;
  description?: string;
}

export interface UpdateObjectPayload {
  name?: string;
  description?: string | null;
}
