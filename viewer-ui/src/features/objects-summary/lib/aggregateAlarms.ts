import type { AlarmRecord } from "@/features/alarms";

export interface ObjectAlarmCounts {
  alarm: number;
  warn: number;
  acked: number;
  bad: number;
  comm: number;
}

const EMPTY_COUNTS: ObjectAlarmCounts = {
  alarm: 0,
  warn: 0,
  acked: 0,
  bad: 0,
  comm: 0
};

const ensure = (
  map: Map<string, ObjectAlarmCounts>,
  objectId: string
): ObjectAlarmCounts => {
  const existing = map.get(objectId);
  if (existing) {
    return existing;
  }

  const created: ObjectAlarmCounts = { ...EMPTY_COUNTS };
  map.set(objectId, created);
  return created;
};

const inc = (counts: ObjectAlarmCounts, key: keyof ObjectAlarmCounts): void => {
  counts[key] += 1;
};

const applyAlarmRecord = (
  map: Map<string, ObjectAlarmCounts>,
  item: AlarmRecord
): void => {
  const objectId = item.objectId.trim();
  if (!objectId) {
    return;
  }

  const counts = ensure(map, objectId);

  if (item.acked) {
    inc(counts, "acked");
    return;
  }

  switch (item.state) {
    case "lolo":
    case "hihi":
      inc(counts, "alarm");
      return;
    case "lo":
    case "hi":
      inc(counts, "warn");
      return;
    case "bad":
    case "uncertain":
      inc(counts, "bad");
      return;
    case "comm_loss":
    case "offline":
      inc(counts, "comm");
      return;
    default:
      return;
  }
};

export const createEmptyObjectAlarmCounts = (): ObjectAlarmCounts => {
  return { ...EMPTY_COUNTS };
};

export function aggregateAlarmsByObject(
  items: AlarmRecord[]
): Map<string, ObjectAlarmCounts> {
  const result = new Map<string, ObjectAlarmCounts>();

  items.forEach((item) => {
    applyAlarmRecord(result, item);
  });

  return result;
}
