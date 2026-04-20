import { api } from "@shared/api";

import type { Unit } from "./types";

export const unitsApi = {
  list: () => api.get<Unit[]>("/units")
};
