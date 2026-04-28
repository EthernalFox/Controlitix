import { api } from "@/shared/api";
import type { Quality } from "@/shared/modules/charts";

export type DiagramFigureType =
  | "rect"
  | "circle"
  | "ellipse"
  | "wedge"
  | "line"
  | "image"
  | "text"
  | "ring"
  | "arc"
  | "tag"
  | "path";

export interface DiagramObjectItem {
  id: string;
  name: string;
  description: string;
  publishedDiagramCount: number;
  firstPublishedDiagramId: string;
}

export interface DiagramObjectListResponse {
  items: DiagramObjectItem[];
  total: number;
  offset: number;
  limit: number;
}

export interface DiagramListItem {
  id: string;
  objectId: string;
  name: string;
  description: string;
  publishedAt: string;
  figureCount: number;
  boundTagCount: number;
}

export interface DiagramListResponse {
  items: DiagramListItem[];
  total: number;
  offset: number;
  limit: number;
}

export interface DiagramTagMeta {
  id: string;
  name: string;
  deviceId: string;
  deviceName: string;
  unit: {
    id: number;
    name: string;
    symbol: string;
    category: string;
  };
  dataType: {
    id: number;
    name: string;
  };
}

export interface DiagramFigure {
  id: string;
  type: DiagramFigureType;
  tagId: string | null;
  params: Record<string, unknown>;
  tag?: DiagramTagMeta | null;
}

export interface DiagramDetails {
  id: string;
  objectId: string;
  objectName: string;
  name: string;
  description: string;
  publishedAt: string;
  canvas: {
    width: number;
    height: number;
    background: string;
  };
  figures: DiagramFigure[];
}

export interface DiagramSnapshotPoint {
  tagId: string;
  ts: string;
  v: number | null;
  q: Quality;
}

export interface DiagramSnapshot {
  diagramId: string;
  ts: string;
  values: DiagramSnapshotPoint[];
  missingTagIds: string[];
}

const normalizeFigure = (figure: DiagramFigure): DiagramFigure => {
  const rawParams = figure.params;

  return {
    ...figure,
    params:
      rawParams && typeof rawParams === "object" && !Array.isArray(rawParams)
        ? (rawParams as Record<string, unknown>)
        : {}
  };
};

const normalizeDiagram = (diagram: DiagramDetails): DiagramDetails => {
  return {
    ...diagram,
    canvas: {
      width: Number.isFinite(diagram.canvas?.width) ? Number(diagram.canvas.width) : 1920,
      height: Number.isFinite(diagram.canvas?.height) ? Number(diagram.canvas.height) : 1080,
      background: diagram.canvas?.background || "#F5F5F5"
    },
    figures: (diagram.figures ?? []).map(normalizeFigure)
  };
};

export const fetchObjects = async (
  limit = 50,
  offset = 0
): Promise<DiagramObjectListResponse> => {
  return api.get<DiagramObjectListResponse>("/objects", {
    limit,
    offset
  });
};

export const fetchObjectDiagrams = async (
  objectId: string,
  limit = 50,
  offset = 0
): Promise<DiagramListResponse> => {
  return api.get<DiagramListResponse>(`/objects/${objectId}/diagrams`, {
    limit,
    offset
  });
};

export const fetchDiagram = async (diagramId: string): Promise<DiagramDetails> => {
  const diagram = await api.get<DiagramDetails>(`/diagrams/${diagramId}`);

  return normalizeDiagram(diagram);
};

export const fetchDiagramSnapshot = async (
  diagramId: string
): Promise<DiagramSnapshot> => {
  return api.get<DiagramSnapshot>(`/diagrams/${diagramId}/snapshot`);
};
