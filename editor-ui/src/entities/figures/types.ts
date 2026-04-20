import type { ShapeType } from "@entities/shapes";

export interface Figure {
  id: string;
  diagramId: string;
  tagId: string | null;
  type: ShapeType;
  params: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface CreateFigurePayload {
  type: ShapeType;
  params: Record<string, unknown>;
  tag_id?: string;
}

export interface UpdateFigurePayload {
  type?: ShapeType;
  params?: Record<string, unknown>;
  tag_id?: string | null;
}
