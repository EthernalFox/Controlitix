import { api } from "@/shared/api";

export type AlarmState =
  | "ok"
  | "lo"
  | "hi"
  | "lolo"
  | "hihi"
  | "uncertain"
  | "bad"
  | "comm_loss"
  | "offline";

export type AlarmListStatus = "active" | "acked" | "cleared";
export type AlarmSeverityFilter = "" | "warn" | "alarm";

export interface AlarmAck {
  actorId: string;
  ackedAt: string;
  note: string | null;
}

export interface AlarmRecord {
  tagId: string;
  tagName: string;
  deviceId: string;
  deviceName: string;
  objectId: string;
  objectName: string;
  state: AlarmState;
  value: number | null;
  quality: string;
  enteredAt: string;
  lastSeenAt: string;
  acked: boolean;
  ack: AlarmAck | null;
}

export interface AlarmEvent {
  id?: string;
  eventType: "raised" | "cleared" | "acked" | "suppressed" | "unsuppressed";
  tagId: string;
  objectId?: string | null;
  stateFrom: AlarmState;
  stateTo: AlarmState;
  value: number | null;
  quality: string;
  ts: string;
  actorId: string | null;
  note: string | null;
}

export interface AlarmDetail {
  current: AlarmRecord | null;
  events: AlarmEvent[];
}

export interface AlarmListResponse {
  items: AlarmRecord[];
  total: number;
  offset: number;
  limit: number;
}

export interface AlarmsListParams {
  status: AlarmListStatus;
  severity: AlarmSeverityFilter;
  objectId: string;
  from: string;
  to: string;
  limit: number;
  offset: number;
}

export interface AcknowledgeAlarmPayload {
  note?: string | null;
}

export interface BulkAckItemResult {
  tagId: string;
  status: "acked" | "not_active" | "not_found" | "already_acked";
  state?: AlarmState;
}

export interface BulkAckResponse {
  items: BulkAckItemResult[];
  ackedAt: string;
  actorId: string;
  successN: number;
  failedN: number;
}

const toParams = (params: AlarmsListParams): Record<string, string | number> => {
  return {
    status: params.status,
    severity: params.severity,
    object_id: params.objectId,
    from: params.from,
    to: params.to,
    limit: params.limit,
    offset: params.offset
  };
};

export const fetchAlarms = async (params: AlarmsListParams): Promise<AlarmListResponse> => {
  return api.get<AlarmListResponse>("/alarms", toParams(params));
};

export const fetchAlarm = async (tagId: string): Promise<AlarmDetail> => {
  return api.get<AlarmDetail>(`/alarms/${tagId}`);
};

export const acknowledgeAlarm = async (
  tagId: string,
  payload: AcknowledgeAlarmPayload
): Promise<AlarmRecord> => {
  return api.post<AlarmRecord>(`/alarms/${tagId}/acknowledge`, payload);
};

export const bulkAcknowledge = async (
  tagIds: string[],
  note: string | null
): Promise<BulkAckResponse> => {
  return api.post<BulkAckResponse>("/alarms/acknowledge", {
    items: tagIds.map((tagId) => ({ tag_id: tagId })),
    note
  });
};
