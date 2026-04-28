export interface Pagination {
  offset: number;
  limit: number;
}

export interface ListResponse<T> extends Pagination {
  items: T[];
  total: number;
}

export type Quality =
  | "ok"
  | "warn"
  | "alarm"
  | "uncertain"
  | "bad"
  | "comm"
  | "offline"
  | "ack";
