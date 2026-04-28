import type { Quality } from "@/shared/modules/charts";

export type RealtimeTopic =
  | `tag:${string}`
  | `object:${string}`
  | `diagram:${string}`
  | "alarms";

export type ConnectionStatus =
  | "connecting"
  | "open"
  | "reconnecting"
  | "closed";

export type RealtimeErrorCode =
  | "limit-exceeded"
  | "invalid-topic"
  | "auth-required"
  | "unauthorized"
  | "rate-limited";

export interface SubscribeMessage {
  t: "subscribe";
  topics: string[];
}

export interface UnsubscribeMessage {
  t: "unsubscribe";
  topics: string[];
}

export interface PingMessage {
  t: "ping";
}

export type ClientMessage = SubscribeMessage | UnsubscribeMessage | PingMessage;

export interface WelcomeMessage {
  t: "welcome";
  session_id: string;
  server_time: string;
  limits: {
    max_subscriptions: number;
    debounce_ms: number;
  };
}

export interface SubscribedMessage {
  t: "subscribed";
  topics: string[];
}

export interface UnsubscribedMessage {
  t: "unsubscribed";
  topics: string[];
}

export interface ValueMessage {
  t: "value";
  tag_id: string;
  ts: string;
  v: number | null;
  q: Quality;
}

export interface SnapshotMessage {
  t: "snapshot";
  tag_id: string;
  ts: string;
  v: number | null;
  q: Quality;
  reason: "subscribe" | "heartbeat";
}

export interface PongMessage {
  t: "pong";
}

export interface ErrorMessage {
  t: "error";
  code: RealtimeErrorCode;
  detail: string;
  topics?: string[];
}

export interface ConfigChangedMessage {
  t: "config.changed";
  entity_type: string;
  entity_id: string;
  operation: string;
  timestamp: string;
  payload?: Record<string, unknown>;
}

export interface TopicsChangedMessage {
  t: "topics_changed";
  topic: string;
  added_count: number;
  removed_count: number;
}

export interface AlarmAckPayload {
  actor_id: string;
  acked_at: string;
  note: string | null;
}

export interface AlarmRecordPayload {
  tag_id: string;
  tag_name: string;
  device_id: string;
  device_name: string;
  object_id: string;
  object_name: string;
  state: "ok" | "lo" | "hi" | "lolo" | "hihi" | "uncertain" | "bad" | "comm_loss" | "offline";
  value: number | null;
  quality: Quality;
  entered_at: string;
  last_seen_at: string;
  acked: boolean;
  ack: AlarmAckPayload | null;
}

export interface AlarmMessage {
  t: "alarm";
  event_type: "raised" | "cleared" | "acked" | "suppressed" | "unsuppressed";
  tag_id: string;
  state_from: "ok" | "lo" | "hi" | "lolo" | "hihi" | "uncertain" | "bad" | "comm_loss" | "offline";
  state_to: "ok" | "lo" | "hi" | "lolo" | "hihi" | "uncertain" | "bad" | "comm_loss" | "offline";
  value: number | null;
  quality: Quality;
  ts: string;
  actor_id: string | null;
  note: string | null;
}

export interface AlarmsSnapshotMessage {
  t: "alarms_snapshot";
  ts: string;
  items: AlarmRecordPayload[];
}

export type ServerMessage =
  | WelcomeMessage
  | SubscribedMessage
  | UnsubscribedMessage
  | ValueMessage
  | SnapshotMessage
  | PongMessage
  | ErrorMessage
  | ConfigChangedMessage
  | TopicsChangedMessage
  | AlarmMessage
  | AlarmsSnapshotMessage;
