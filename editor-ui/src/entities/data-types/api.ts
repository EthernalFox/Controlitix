import { api } from "@shared/api";

import type { DataType } from "./types";

export const dataTypesApi = {
  list: () => api.get<DataType[]>("/data-types")
};
