import type { MonitoringObjectModel } from "./types";

export const monitoringObjectsMock: MonitoringObjectModel[] = [
  {
    id: "obj-1",
    name: "Цех №1",
    description: "Основной производственный цех",
    createdAt: "2025-09-01T10:00:00.000Z",
    updatedAt: "2025-12-01T12:00:00.000Z"
  },
  {
    id: "obj-2",
    name: "Котельная",
    description: "Узел теплоснабжения",
    createdAt: "2025-09-05T10:00:00.000Z",
    updatedAt: "2025-12-10T09:30:00.000Z"
  },
  {
    id: "obj-3",
    name: "Склад",
    description: null,
    createdAt: "2025-10-12T08:15:00.000Z",
    updatedAt: "2025-12-15T16:45:00.000Z"
  }
];

