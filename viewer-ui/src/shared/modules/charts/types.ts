export type Quality =
  | "ok"
  | "hi"
  | "hihi"
  | "uncertain"
  | "bad"
  | "comm_loss"
  | "offline"
  | "acknowledged";

export interface TrendPoint {
  ts: string;
  v: number | null;
  q: Quality;
}

export interface TrendSeries {
  tagId: string;
  tagName: string;
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
  from: string;
  to: string;
  step: string;
  agg: "last" | "avg" | "min" | "max";
  source: "raw" | "agg_1m";
  points: TrendPoint[];
}
