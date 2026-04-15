import type { RouteObject } from "react-router";

import { lazyLoader } from "./LazyLoader";

export const routePaths = {
  objects: "/objects",
  mimics: "/mimics",
  devices: "/devices",
  draw: "/draw"
} as const;

export const routes: RouteObject[] = [
  {
    path: "/",
    lazy: lazyLoader("../../../pages/MonitoringObjectsPage"),
    handle: { title: "Объекты мониторинга" }
  },
  {
    path: routePaths.objects,
    lazy: lazyLoader("../../../pages/MonitoringObjectsPage"),
    handle: { title: "Объекты мониторинга" }
  },
  {
    path: routePaths.mimics,
    lazy: lazyLoader("../../../pages/MimicsPage"),
    handle: { title: "Мнемосхемы" }
  },
  {
    path: routePaths.devices,
    lazy: lazyLoader("../../../pages/DevicesPage"),
    handle: { title: "Устройства" }
  },
  {
    path: routePaths.draw,
    lazy: lazyLoader("../../../pages/DrawPage"),
    handle: { title: "Редактор" }
  }
];
