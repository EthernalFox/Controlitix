import type { MimicModel } from "../types";

export interface MimicApiModel {
  id: string;
  object_id: string;
  name: string | null;
  description: string | null;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export type MimicId = MimicApiModel["id"];

export type MimicObjectId = MimicApiModel["object_id"];

export type { MimicModel };

export interface MimicPayload {
  name?: string | null;
  description?: string | null;
}
