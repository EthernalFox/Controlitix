import {
  fetchAlarms,
  type AlarmEvent,
  type AlarmRecord
} from "@/features/alarms";

import {
  aggregateAlarmsByObject,
  createEmptyObjectAlarmCounts,
  type ObjectAlarmCounts
} from "../lib/aggregateAlarms";

export interface ObjectAlarmsSnapshot {
  countsByObject: Map<string, ObjectAlarmCounts>;
  alarmsByTag: Map<
    string,
    {
      objectId: string;
      state: AlarmRecord["state"];
      acked: boolean;
    }
  >;
}

const normalizeTagId = (value: unknown): string => {
  if (typeof value !== "string") {
    return "";
  }

  return value.trim();
};

const normalizeObjectId = (value: unknown): string => {
  if (typeof value !== "string") {
    return "";
  }

  return value.trim();
};

const cloneCounts = (
  source: Map<string, ObjectAlarmCounts>
): Map<string, ObjectAlarmCounts> => {
  const copy = new Map<string, ObjectAlarmCounts>();
  source.forEach((value, key) => {
    copy.set(key, { ...value });
  });
  return copy;
};

const adjustByEvent = (
  counts: ObjectAlarmCounts,
  state: AlarmRecord["state"],
  acked: boolean,
  delta: 1 | -1
): void => {
  const mutate = (key: keyof ObjectAlarmCounts) => {
    counts[key] = Math.max(0, counts[key] + delta);
  };

  if (acked) {
    mutate("acked");
    return;
  }

  switch (state) {
    case "lolo":
    case "hihi":
      mutate("alarm");
      return;
    case "lo":
    case "hi":
      mutate("warn");
      return;
    case "bad":
    case "uncertain":
      mutate("bad");
      return;
    case "comm_loss":
    case "offline":
      mutate("comm");
      return;
    default:
      return;
  }
};

export const fetchActiveAlarmsSnapshot = async (): Promise<ObjectAlarmsSnapshot> => {
  const response = await fetchAlarms({
    status: "active",
    severity: "",
    objectId: "",
    from: "",
    to: "",
    limit: 500,
    offset: 0
  });

  const countsByObject = aggregateAlarmsByObject(response.items);
  const alarmsByTag = new Map<
    string,
    {
      objectId: string;
      state: AlarmRecord["state"];
      acked: boolean;
    }
  >();

  response.items.forEach((item) => {
    const tagId = normalizeTagId(item.tagId);
    const objectId = normalizeObjectId(item.objectId);

    if (!tagId || !objectId) {
      return;
    }

    alarmsByTag.set(tagId, {
      objectId,
      state: item.state,
      acked: item.acked
    });
  });

  return {
    countsByObject,
    alarmsByTag
  };
};

export async function fetchActiveAlarmsByObject(): Promise<
  Map<string, ObjectAlarmCounts>
> {
  const snapshot = await fetchActiveAlarmsSnapshot();
  return cloneCounts(snapshot.countsByObject);
}

export const applyAlarmEventToAggregates = (
  sourceCounts: Map<string, ObjectAlarmCounts>,
  sourceTagState: Map<string, { objectId: string; state: AlarmRecord["state"]; acked: boolean }>,
  event: AlarmEvent & { objectId?: string | null }
): {
  countsByObject: Map<string, ObjectAlarmCounts>;
  alarmsByTag: Map<string, { objectId: string; state: AlarmRecord["state"]; acked: boolean }>;
} => {
  const countsByObject = cloneCounts(sourceCounts);
  const alarmsByTag = new Map(sourceTagState);

  const tagId = normalizeTagId(event.tagId);
  if (!tagId) {
    return { countsByObject, alarmsByTag };
  }

  const known = alarmsByTag.get(tagId);

  const ensureObjectCounts = (objectId: string): ObjectAlarmCounts => {
    const existing = countsByObject.get(objectId);
    if (existing) {
      return existing;
    }

    const created = createEmptyObjectAlarmCounts();
    countsByObject.set(objectId, created);
    return created;
  };

  if (event.eventType === "cleared") {
    if (!known) {
      return { countsByObject, alarmsByTag };
    }

    const counts = ensureObjectCounts(known.objectId);
    adjustByEvent(counts, known.state, known.acked, -1);
    alarmsByTag.delete(tagId);

    return { countsByObject, alarmsByTag };
  }

  if (event.eventType === "acked") {
    if (!known) {
      const objectId = normalizeObjectId(event.objectId);
      if (!objectId) {
        return { countsByObject, alarmsByTag };
      }

      const counts = ensureObjectCounts(objectId);
      adjustByEvent(counts, event.stateTo, true, 1);
      alarmsByTag.set(tagId, {
        objectId,
        state: event.stateTo,
        acked: true
      });

      return { countsByObject, alarmsByTag };
    }

    const counts = ensureObjectCounts(known.objectId);
    adjustByEvent(counts, known.state, known.acked, -1);
    adjustByEvent(counts, known.state, true, 1);
    alarmsByTag.set(tagId, {
      objectId: known.objectId,
      state: known.state,
      acked: true
    });

    return { countsByObject, alarmsByTag };
  }

  if (event.eventType === "raised") {
    const objectId = normalizeObjectId(event.objectId) || known?.objectId || "";
    if (!objectId) {
      return { countsByObject, alarmsByTag };
    }

    const counts = ensureObjectCounts(objectId);

    if (known) {
      adjustByEvent(counts, known.state, known.acked, -1);
    }

    adjustByEvent(counts, event.stateTo, false, 1);
    alarmsByTag.set(tagId, {
      objectId,
      state: event.stateTo,
      acked: false
    });

    return { countsByObject, alarmsByTag };
  }

  return { countsByObject, alarmsByTag };
};
