import { notifications } from "@mantine/notifications";
import { create } from "zustand";

import {
  acknowledgeAlarm,
  bulkAcknowledge,
  fetchAlarms,
  type AlarmAck,
  type AlarmListResponse,
  type AlarmRecord,
  type AlarmsListParams,
  type BulkAckResponse
} from "@/features/alarms/model/alarmsApi";
import { ApiRequestError } from "@/shared/api";
import { realtimeClient } from "@/shared/modules/realtime";
import type {
  AlarmMessage,
  AlarmsBatchMessage,
  AlarmRecordPayload,
  AlarmsSnapshotMessage
} from "@/shared/modules/realtime/types";

const defaultQuery: AlarmsListParams = {
  status: "active",
  severity: "",
  objectId: "",
  from: "",
  to: "",
  limit: 50,
  offset: 0
};

const ALARMS_TOPIC = "alarms";

let realtimeRefCount = 0;
let realtimeUnsubscribeSnapshot: (() => void) | null = null;
let realtimeUnsubscribeAlarm: (() => void) | null = null;
let realtimeUnsubscribeAlarmBatch: (() => void) | null = null;
let refreshTimerId: number | null = null;

const toAlarmRecord = (item: AlarmRecordPayload): AlarmRecord => {
  return {
    tagId: item.tag_id,
    tagName: item.tag_name,
    deviceId: item.device_id,
    deviceName: item.device_name,
    objectId: item.object_id,
    objectName: item.object_name,
    state: item.state,
    value: item.value,
    quality: item.quality,
    enteredAt: item.entered_at,
    lastSeenAt: item.last_seen_at,
    acked: item.acked,
    ack: item.ack
      ? {
          actorId: item.ack.actor_id,
          ackedAt: item.ack.acked_at,
          note: item.ack.note
        }
      : null
  };
};

const mapSnapshot = (message: AlarmsSnapshotMessage): AlarmRecord[] => {
  return message.items.map(toAlarmRecord);
};

const resolveErrorMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  return "Не удалось выполнить запрос тревог.";
};

const upsertAlarmRecord = (items: AlarmRecord[], nextRecord: AlarmRecord): AlarmRecord[] => {
  const without = items.filter((item) => item.tagId !== nextRecord.tagId);
  return [nextRecord, ...without];
};

const updateByEvent = (items: AlarmRecord[], event: AlarmMessage): AlarmRecord[] => {
  if (event.event_type === "cleared") {
    return items.filter((item) => item.tagId !== event.tag_id);
  }

  if (event.event_type === "acked") {
    return items.map((item) => {
      if (item.tagId !== event.tag_id) {
        return item;
      }

      const ack: AlarmAck = {
        actorId: event.actor_id || "unknown",
        ackedAt: event.ts,
        note: event.note
      };

      return {
        ...item,
        acked: true,
        ack
      };
    });
  }

  const existing = items.find((item) => item.tagId === event.tag_id);
  const next: AlarmRecord = {
    tagId: event.tag_id,
    tagName: existing?.tagName || event.tag_id,
    deviceId: existing?.deviceId || "",
    deviceName: existing?.deviceName || "",
    objectId: existing?.objectId || event.object_id || "",
    objectName: existing?.objectName || "",
    state: event.state_to,
    value: event.value,
    quality: event.quality,
    enteredAt: existing?.enteredAt || event.ts,
    lastSeenAt: event.ts,
    acked: false,
    ack: null
  };

  return upsertAlarmRecord(items, next);
};

const scheduleActiveRefresh = () => {
  if (refreshTimerId !== null) {
    window.clearTimeout(refreshTimerId);
  }

  refreshTimerId = window.setTimeout(() => {
    refreshTimerId = null;
    void useAlarmsStore.getState().refresh();
  }, 800);
};

const startRealtimeInternal = () => {
  if (realtimeRefCount > 0) {
    realtimeRefCount += 1;
    return;
  }

  realtimeRefCount = 1;
  realtimeClient.subscribe([ALARMS_TOPIC]);

  realtimeUnsubscribeSnapshot = realtimeClient.on("alarms_snapshot", (message) => {
    useAlarmsStore.getState().applySnapshot(message);
  });

  realtimeUnsubscribeAlarm = realtimeClient.on("alarm", (message) => {
    useAlarmsStore.getState().applyEvent(message);
  });

  realtimeUnsubscribeAlarmBatch = realtimeClient.on("alarms_batch", (message) => {
    useAlarmsStore.getState().applyBatch(message);
  });
};

const stopRealtimeInternal = () => {
  if (realtimeRefCount <= 0) {
    realtimeRefCount = 0;
    return;
  }

  realtimeRefCount -= 1;
  if (realtimeRefCount > 0) {
    return;
  }

  realtimeClient.unsubscribe([ALARMS_TOPIC]);

  realtimeUnsubscribeSnapshot?.();
  realtimeUnsubscribeSnapshot = null;

  realtimeUnsubscribeAlarm?.();
  realtimeUnsubscribeAlarm = null;

  realtimeUnsubscribeAlarmBatch?.();
  realtimeUnsubscribeAlarmBatch = null;
};

interface AlarmsState {
  query: AlarmsListParams;
  items: AlarmRecord[];
  total: number;
  loading: boolean;
  error: string;
  activeUnacked: AlarmRecord[];
  pendingAckTagIds: string[];

  load: (queryPatch?: Partial<AlarmsListParams>) => Promise<void>;
  refresh: () => Promise<void>;
  acknowledge: (tagId: string, note?: string) => Promise<void>;
  acknowledgeBulk: (tagIds: string[], note: string | null) => Promise<BulkAckResponse>;
  applySnapshot: (message: AlarmsSnapshotMessage) => void;
  applyEvent: (message: AlarmMessage) => void;
  applyBatch: (message: AlarmsBatchMessage) => void;
  startRealtime: () => void;
  stopRealtime: () => void;
}

const applyListResult = (
  response: AlarmListResponse,
  nextQuery: AlarmsListParams
): Pick<AlarmsState, "items" | "total" | "query" | "error" | "loading"> => {
  return {
    items: response.items,
    total: response.total,
    query: nextQuery,
    error: "",
    loading: false
  };
};

const withUniqueTagIds = (tagIds: string[]): string[] => {
  const result: string[] = [];
  const seen = new Set<string>();

  tagIds.forEach((tagId) => {
    const normalized = tagId.trim();
    if (!normalized || seen.has(normalized)) {
      return;
    }

    seen.add(normalized);
    result.push(normalized);
  });

  return result;
};

export const useAlarmsStore = create<AlarmsState>((set, get) => ({
  query: defaultQuery,
  items: [],
  total: 0,
  loading: false,
  error: "",
  activeUnacked: [],
  pendingAckTagIds: [],

  load: async (queryPatch = {}) => {
    const mergedQuery: AlarmsListParams = {
      ...get().query,
      ...queryPatch
    };

    set({ loading: true, error: "", query: mergedQuery });

    try {
      const response = await fetchAlarms(mergedQuery);
      set(applyListResult(response, mergedQuery));
    } catch (error) {
      set({
        loading: false,
        error: resolveErrorMessage(error)
      });
    }
  },

  refresh: async () => {
    const query = get().query;

    try {
      const response = await fetchAlarms(query);
      set(applyListResult(response, query));
    } catch (error) {
      set({ error: resolveErrorMessage(error) });
    }
  },

  acknowledge: async (tagId, note) => {
    const state = get();
    const previousItems = state.items;
    const previousActive = state.activeUnacked;

    const nowIso = new Date().toISOString();

    const optimisticAck: AlarmAck = {
      actorId: "you",
      ackedAt: nowIso,
      note: note ?? null
    };

    const optimisticItems = previousItems.map((item) => {
      if (item.tagId !== tagId) {
        return item;
      }

      return {
        ...item,
        acked: true,
        ack: optimisticAck
      };
    });

    set({
      items: optimisticItems,
      pendingAckTagIds: Array.from(new Set([...state.pendingAckTagIds, tagId])),
      activeUnacked: previousActive.filter((item) => item.tagId !== tagId)
    });

    try {
      const acknowledged = await acknowledgeAlarm(tagId, { note: note ?? null });

      if (state.query.status === "active") {
        set((current) => ({
          items: current.items.filter((item) => item.tagId !== tagId),
          pendingAckTagIds: current.pendingAckTagIds.filter((id) => id !== tagId)
        }));
      } else {
        set((current) => ({
          items: current.items.map((item) => {
            if (item.tagId !== tagId) {
              return item;
            }

            return {
              ...item,
              acked: acknowledged.acked,
              ack: acknowledged.ack
            };
          }),
          pendingAckTagIds: current.pendingAckTagIds.filter((id) => id !== tagId)
        }));
      }
    } catch (error) {
      set({
        items: previousItems,
        activeUnacked: previousActive,
        pendingAckTagIds: state.pendingAckTagIds
      });

      notifications.show({
        color: "alarm-alarm",
        title: "Квитирование не выполнено",
        message: resolveErrorMessage(error)
      });
    }
  },

  acknowledgeBulk: async (tagIds, note) => {
    const normalizedTagIds = withUniqueTagIds(tagIds);
    const state = get();
    const previousItems = state.items;
    const previousActive = state.activeUnacked;

    const nowIso = new Date().toISOString();
    const optimisticAck: AlarmAck = {
      actorId: "you",
      ackedAt: nowIso,
      note: note ?? null
    };

    const selectedSet = new Set(normalizedTagIds);

    const optimisticItems = previousItems.map((item) => {
      if (!selectedSet.has(item.tagId)) {
        return item;
      }

      return {
        ...item,
        acked: true,
        ack: optimisticAck
      };
    });

    set({
      items: optimisticItems,
      pendingAckTagIds: Array.from(new Set([...state.pendingAckTagIds, ...normalizedTagIds])),
      activeUnacked: previousActive.filter((item) => !selectedSet.has(item.tagId))
    });

    try {
      const response = await bulkAcknowledge(normalizedTagIds, note);
      const failedTagIDs = new Set(
        response.items
          .filter((item) => item.status !== "acked")
          .map((item) => item.tagId)
      );
      const ackedTagIDs = new Set(
        response.items
          .filter((item) => item.status === "acked")
          .map((item) => item.tagId)
      );

      set((current) => {
        const previousByTagID = new Map(previousItems.map((item) => [item.tagId, item]));

        let nextItems = current.items;
        if (current.query.status === "active") {
          nextItems = current.items.filter((item) => !ackedTagIDs.has(item.tagId));
        }

        nextItems = nextItems.map((item) => {
          if (failedTagIDs.has(item.tagId)) {
            return previousByTagID.get(item.tagId) || item;
          }

          return item;
        });

        failedTagIDs.forEach((failedTagID) => {
          const exists = nextItems.some((item) => item.tagId === failedTagID);
          if (exists) {
            return;
          }

          const previous = previousByTagID.get(failedTagID);
          if (previous) {
            nextItems = [previous, ...nextItems];
          }
        });

        return {
          items: nextItems,
          pendingAckTagIds: current.pendingAckTagIds.filter(
            (tagID) => !selectedSet.has(tagID)
          ),
          activeUnacked: previousActive.filter((item) => !ackedTagIDs.has(item.tagId))
        };
      });

      return response;
    } catch (error) {
      set({
        items: previousItems,
        activeUnacked: previousActive,
        pendingAckTagIds: state.pendingAckTagIds
      });
      throw error;
    }
  },

  applySnapshot: (message) => {
    const snapshot = mapSnapshot(message);

    set((state) => {
      if (state.query.status === "active" && state.query.offset === 0 && !state.loading) {
        return {
          activeUnacked: snapshot,
          items: snapshot,
          total: snapshot.length
        };
      }

      return {
        activeUnacked: snapshot
      };
    });
  },

  applyEvent: (message) => {
    set((state) => {
      const nextActive = updateByEvent(state.activeUnacked, message);
      let nextItems = state.items;
      let nextTotal = state.total;

      if (state.query.status === "active" && state.query.offset === 0) {
        nextItems = updateByEvent(state.items, message);
        nextTotal = nextItems.length;
      }

      return {
        activeUnacked: nextActive,
        items: nextItems,
        total: nextTotal
      };
    });

    if (get().query.status === "active") {
      scheduleActiveRefresh();
    }
  },

  applyBatch: (message) => {
    set((state) => {
      let nextActive = state.activeUnacked;
      let nextItems = state.items;

      message.events.forEach((event) => {
        nextActive = updateByEvent(nextActive, {
          ...event,
          t: "alarm"
        });

        if (state.query.status === "active" && state.query.offset === 0) {
          nextItems = updateByEvent(nextItems, {
            ...event,
            t: "alarm"
          });
        }
      });

      return {
        activeUnacked: nextActive,
        items: nextItems,
        total:
          state.query.status === "active" && state.query.offset === 0
            ? nextItems.length
            : state.total
      };
    });

    if (get().query.status === "active") {
      scheduleActiveRefresh();
    }
  },

  startRealtime: () => {
    startRealtimeInternal();
  },

  stopRealtime: () => {
    stopRealtimeInternal();
  }
}));
