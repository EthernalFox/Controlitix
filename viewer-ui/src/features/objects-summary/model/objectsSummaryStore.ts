import { create } from "zustand";

import type { AlarmEvent } from "@/features/alarms";
import { listObjects, type ObjectSummary } from "@/shared/api/objects";

import {
  applyAlarmEventToAggregates,
  fetchActiveAlarmsSnapshot,
  type ObjectAlarmsSnapshot
} from "./objectsSummaryApi";
import { createEmptyObjectAlarmCounts, type ObjectAlarmCounts } from "../lib/aggregateAlarms";

type ViewState = "idle" | "loading" | "ready" | "error";

interface ObjectsSummaryState {
  status: ViewState;
  objects: ObjectSummary[];
  alarmCounts: Map<string, ObjectAlarmCounts>;
  alarmsByTag: ObjectAlarmsSnapshot["alarmsByTag"];
  lastError: string | null;

  load: () => Promise<void>;
  refresh: () => Promise<void>;
  applyAlarmEvent: (event: AlarmEvent & { objectId?: string | null }) => void;
}

const mapError = (error: unknown): string => {
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }

  return "Не удалось загрузить данные дашборда";
};

export const useObjectsSummaryStore = create<ObjectsSummaryState>((set, get) => ({
  status: "idle",
  objects: [],
  alarmCounts: new Map(),
  alarmsByTag: new Map(),
  lastError: null,

  load: async () => {
    set({ status: "loading", lastError: null });

    try {
      const [objects, alarmsSnapshot] = await Promise.all([
        listObjects(),
        fetchActiveAlarmsSnapshot()
      ]);

      set({
        status: "ready",
        objects,
        alarmCounts: alarmsSnapshot.countsByObject,
        alarmsByTag: alarmsSnapshot.alarmsByTag,
        lastError: null
      });
    } catch (error) {
      set({
        status: "error",
        lastError: mapError(error)
      });
    }
  },

  refresh: async () => {
    await get().load();
  },

  applyAlarmEvent: (event) => {
    set((state) => {
      const next = applyAlarmEventToAggregates(
        state.alarmCounts,
        state.alarmsByTag,
        event
      );

      return {
        alarmCounts: next.countsByObject,
        alarmsByTag: next.alarmsByTag,
        status: state.status === "idle" ? "ready" : state.status
      };
    });
  }
}));

export const getObjectAlarmCounts = (
  map: Map<string, ObjectAlarmCounts>,
  objectId: string
): ObjectAlarmCounts => {
  const value = map.get(objectId);
  if (value) {
    return value;
  }

  return createEmptyObjectAlarmCounts();
};
