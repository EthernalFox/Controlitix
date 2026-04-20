export interface Diagram {
  id: string;
  objectId: string;
  name: string | null;
  description: string | null;
  publishedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateDiagramPayload {
  name?: string;
  description?: string;
}

