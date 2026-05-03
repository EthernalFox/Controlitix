export {
  fetchActiveAlarmsByObject,
  type ObjectAlarmsSnapshot
} from "./model/objectsSummaryApi";
export {
  getObjectAlarmCounts,
  useObjectsSummaryStore
} from "./model/objectsSummaryStore";
export type { ObjectAlarmCounts } from "./lib/aggregateAlarms";
