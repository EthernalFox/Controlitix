import { getAccessToken } from "@/shared/api";

import { useRealtimeStore } from "./store";
import type {
  ClientMessage,
  ConnectionStatus,
  ErrorMessage,
  RealtimeTopic,
  ServerMessage
} from "./types";

const MAX_RECONNECT_DELAY_MS = 30_000;
const BASE_RECONNECT_DELAY_MS = 1_000;
const RECONNECT_JITTER_MS = 200;
const WATCHDOG_TIMEOUT_MS = 45_000;

type MessageType = ServerMessage["t"];
type MessageByType<T extends MessageType> = Extract<ServerMessage, { t: T }>;

type HandlerMap = {
  [Key in MessageType]: Set<(message: MessageByType<Key>) => void>;
};

export interface RealtimeClient {
  start(): void;
  stop(): void;
  subscribe(topics: string[]): void;
  unsubscribe(topics: string[]): void;
  reconnect(): void;
  on<T extends ServerMessage["t"]>(
    type: T,
    handler: (message: MessageByType<T>) => void
  ): () => void;
}

class RealtimeClientImpl implements RealtimeClient {
  private socket: WebSocket | null = null;
  private reconnectTimerId: number | null = null;
  private watchdogTimerId: number | null = null;
  private reconnectAttempt = 0;
  private started = false;
  private forceReconnect = false;
  private connectedToken: string | null = null;
  private readonly topicRefCount = new Map<string, number>();

  private readonly handlers: HandlerMap = {
    welcome: new Set(),
    subscribed: new Set(),
    unsubscribed: new Set(),
    value: new Set(),
    snapshot: new Set(),
    pong: new Set(),
    error: new Set(),
    "config.changed": new Set(),
    topics_changed: new Set(),
    alarm: new Set(),
    alarms_snapshot: new Set()
  };

  start(): void {
    this.started = true;

    const token = getAccessToken();
    if (!token) {
      this.updateStatus("closed");
      return;
    }

    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      if (token !== this.connectedToken) {
        this.reconnect();
      }
      return;
    }

    this.connect(this.reconnectAttempt > 0);
  }

  stop(): void {
    this.started = false;
    this.forceReconnect = false;
    this.clearReconnectTimer();
    this.clearWatchdogTimer();

    const activeTopics = this.getActiveTopics();
    if (activeTopics.length > 0) {
      this.send({ t: "unsubscribe", topics: activeTopics });
    }

    this.topicRefCount.clear();
    useRealtimeStore.getState().setTopicRefCount(this.topicRefCount);

    if (this.socket && this.socket.readyState <= WebSocket.OPEN) {
      this.socket.close(1000, "manual_stop");
    }

    this.socket = null;
    this.connectedToken = null;
    useRealtimeStore.getState().reset(true);
  }

  reconnect(): void {
    if (!this.started) {
      return;
    }

    this.clearReconnectTimer();
    this.reconnectAttempt = 0;
    this.forceReconnect = true;

    if (this.socket && this.socket.readyState <= WebSocket.OPEN) {
      this.socket.close(1000, "token_refresh");
      return;
    }

    this.connect(true);
  }

  subscribe(topics: string[]): void {
    const normalizedTopics = normalizeTopics(topics);
    if (normalizedTopics.length === 0) {
      return;
    }

    const newTopics: string[] = [];

    normalizedTopics.forEach((topic) => {
      const currentCount = this.topicRefCount.get(topic) ?? 0;
      this.topicRefCount.set(topic, currentCount + 1);

      if (currentCount === 0) {
        newTopics.push(topic);
      }
    });

    useRealtimeStore.getState().setTopicRefCount(this.topicRefCount);

    if (newTopics.length > 0 && this.isOpen()) {
      this.send({ t: "subscribe", topics: newTopics });
    }
  }

  unsubscribe(topics: string[]): void {
    const normalizedTopics = normalizeTopics(topics);
    if (normalizedTopics.length === 0) {
      return;
    }

    const removedTopics: string[] = [];

    normalizedTopics.forEach((topic) => {
      const currentCount = this.topicRefCount.get(topic);
      if (!currentCount) {
        return;
      }

      if (currentCount <= 1) {
        this.topicRefCount.delete(topic);
        removedTopics.push(topic);
        return;
      }

      this.topicRefCount.set(topic, currentCount - 1);
    });

    useRealtimeStore.getState().setTopicRefCount(this.topicRefCount);

    if (removedTopics.length > 0 && this.isOpen()) {
      this.send({ t: "unsubscribe", topics: removedTopics });
    }
  }

  on<T extends ServerMessage["t"]>(
    type: T,
    handler: (message: MessageByType<T>) => void
  ): () => void {
    const typedHandlers = this.handlers[type] as Set<(message: MessageByType<T>) => void>;

    typedHandlers.add(handler);

    return () => {
      typedHandlers.delete(handler);
    };
  }

  private connect(reconnecting: boolean): void {
    const token = getAccessToken();
    if (!token) {
      this.updateStatus("closed");
      return;
    }

    this.connectedToken = token;
    this.updateStatus(reconnecting ? "reconnecting" : "connecting");

    const socket = new WebSocket(buildRealtimeURL(token));
    this.socket = socket;

    socket.onopen = () => {
      this.resetWatchdog();
    };

    socket.onmessage = (event) => {
      this.resetWatchdog();
      this.handleMessage(event.data);
    };

    socket.onclose = () => {
      this.clearWatchdogTimer();
      this.socket = null;
      this.connectedToken = null;

      if (!this.started) {
        this.updateStatus("closed");
        return;
      }

      if (this.forceReconnect) {
        this.forceReconnect = false;
        this.connect(true);
        return;
      }

      this.scheduleReconnect();
    };

    socket.onerror = () => {
      // Error details are delivered in close event and parsed server messages.
    };
  }

  private handleMessage(rawData: unknown): void {
    if (typeof rawData !== "string") {
      return;
    }

    let message: ServerMessage;
    try {
      message = JSON.parse(rawData) as ServerMessage;
    } catch {
      return;
    }

    if (!message || typeof message !== "object" || typeof message.t !== "string") {
      return;
    }

    switch (message.t) {
      case "welcome": {
        this.reconnectAttempt = 0;
        this.updateStatus("open");
        const topics = this.getActiveTopics();
        if (topics.length > 0) {
          this.send({ t: "subscribe", topics });
        }
        this.emit("welcome", message);
        return;
      }
      case "subscribed":
      case "unsubscribed":
      case "pong": {
        this.emit(message.t, message);
        return;
      }
      case "value":
      case "snapshot": {
        useRealtimeStore.getState().upsertLastMessage(message);
        this.emit(message.t, message);
        return;
      }
      case "error": {
        this.logError(message);
        this.emit("error", message);
        return;
      }
      case "config.changed": {
        useRealtimeStore.getState().setLastConfigChange(message);
        this.emit("config.changed", message);
        return;
      }
      case "topics_changed": {
        this.emit("topics_changed", message);
        return;
      }
      case "alarm":
      case "alarms_snapshot": {
        this.emit(message.t, message);
        return;
      }
      default:
        return;
    }
  }

  private emit<T extends MessageType>(type: T, message: MessageByType<T>): void {
    this.handlers[type].forEach((handler) => {
      handler(message);
    });
  }

  private logError(message: ErrorMessage): void {
    const topicSuffix = message.topics?.length
      ? ` topics=${message.topics.join(",")}`
      : "";

    console.warn(`[realtime] ${message.code}: ${message.detail}${topicSuffix}`);
  }

  private scheduleReconnect(): void {
    this.clearReconnectTimer();
    this.updateStatus("reconnecting");

    const expFactor = Math.min(this.reconnectAttempt, 6);
    const baseDelay = Math.min(
      MAX_RECONNECT_DELAY_MS,
      BASE_RECONNECT_DELAY_MS * 2 ** expFactor
    );
    const jitter = Math.floor((Math.random() * 2 - 1) * RECONNECT_JITTER_MS);
    const delay = Math.max(0, baseDelay + jitter);

    this.reconnectAttempt += 1;

    this.reconnectTimerId = window.setTimeout(() => {
      this.reconnectTimerId = null;
      this.connect(true);
    }, delay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimerId !== null) {
      window.clearTimeout(this.reconnectTimerId);
      this.reconnectTimerId = null;
    }
  }

  private resetWatchdog(): void {
    this.clearWatchdogTimer();

    this.watchdogTimerId = window.setTimeout(() => {
      if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        return;
      }

      this.socket.close(1001, "watchdog_timeout");
    }, WATCHDOG_TIMEOUT_MS);
  }

  private clearWatchdogTimer(): void {
    if (this.watchdogTimerId !== null) {
      window.clearTimeout(this.watchdogTimerId);
      this.watchdogTimerId = null;
    }
  }

  private isOpen(): boolean {
    return this.socket?.readyState === WebSocket.OPEN;
  }

  private send(message: ClientMessage): void {
    if (!this.isOpen()) {
      return;
    }

    this.socket?.send(JSON.stringify(message));
  }

  private getActiveTopics(): string[] {
    return Array.from(this.topicRefCount.entries())
      .filter(([, count]) => count > 0)
      .map(([topic]) => topic);
  }

  private updateStatus(status: ConnectionStatus): void {
    useRealtimeStore.getState().setStatus(status);
  }
}

const normalizeTopics = (topics: string[]): string[] => {
  const result: string[] = [];
  const unique = new Set<string>();

  topics.forEach((topic) => {
    const trimmedTopic = topic.trim();
    if (!trimmedTopic || unique.has(trimmedTopic)) {
      return;
    }

    unique.add(trimmedTopic);
    result.push(trimmedTopic);
  });

  return result;
};

const buildRealtimeURL = (accessToken: string): string => {
  const baseURL = (import.meta.env.VITE_API_URL ?? "/api").replace(/\/$/, "");
  const url = new URL(`${baseURL}/ws`, window.location.origin);

  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.searchParams.set("access_token", accessToken);

  return url.toString();
};

export const realtimeClient: RealtimeClient = new RealtimeClientImpl();

export type { RealtimeTopic };
