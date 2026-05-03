import { api } from "./client";

interface RawObjectSummary {
  id: string;
  name?: string;
  description?: string | null;
  diagramsCount?: number;
  defaultDiagramId?: string | null;
  devicesTotal?: number;
  devicesOffline?: number;
  publishedDiagramCount?: number;
  firstPublishedDiagramId?: string;
}

interface RawObjectListResponse {
  items?: RawObjectSummary[];
}

export interface ObjectSummary {
  id: string;
  name: string;
  description: string | null;
  diagramsCount: number;
  defaultDiagramId: string | null;
  devicesTotal: number;
  devicesOffline: number;
}

const toInt = (value: unknown, fallback = 0): number => {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return fallback;
  }

  return Math.max(0, Math.trunc(value));
};

const toDefaultDiagramId = (value: unknown): string | null => {
  if (typeof value !== "string") {
    return null;
  }

  const normalized = value.trim();
  return normalized.length > 0 ? normalized : null;
};

export async function listObjects(): Promise<ObjectSummary[]> {
  const response = await api.get<RawObjectListResponse>("/objects", {
    limit: 500,
    offset: 0
  });

  const items = Array.isArray(response.items) ? response.items : [];

  return items.map((item) => ({
    id: item.id,
    name: (item.name ?? "").trim() || "Без названия",
    description: item.description ?? null,
    diagramsCount: toInt(item.diagramsCount ?? item.publishedDiagramCount ?? 0),
    defaultDiagramId: toDefaultDiagramId(
      item.defaultDiagramId ?? item.firstPublishedDiagramId ?? null
    ),
    devicesTotal: toInt(item.devicesTotal ?? 0),
    devicesOffline: toInt(item.devicesOffline ?? 0)
  }));
}
