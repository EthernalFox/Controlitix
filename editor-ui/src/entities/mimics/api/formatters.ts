import { createFormatter } from "@shared/modules/api";

import type { MimicApiModel, MimicModel } from "./types";

export const mimicsFormatter = createFormatter<MimicApiModel, MimicModel>(
  (raw) => ({
    id: raw.id,
    objectId: raw.object_id,
    name: raw.name,
    description: raw.description,
    publishedAt: raw.published_at,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at
  })
);
