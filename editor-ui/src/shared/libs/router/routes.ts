import { createElement } from "react";
import { Navigate, type RouteObject } from "react-router";

import { lazyLoader } from "./LazyLoader";

export const routePaths = {
  objects: "/objects",
  objectDetail: "/objects/:objectId",
  objectDevices: "/objects/:objectId/devices",
  objectTags: "/objects/:objectId/tags",
  objectDiagrams: "/objects/:objectId/diagrams",
  diagramEditor: "/objects/:objectId/diagrams/:diagramId/edit"
} as const;

export const routes: RouteObject[] = [
  {
    path: "/",
    element: createElement(Navigate, { to: routePaths.objects, replace: true })
  },
  {
    lazy: lazyLoader("../../../widgets/ListLayout"),
    children: [
      {
        path: routePaths.objects,
        lazy: lazyLoader("../../../pages/ObjectsPage"),
        handle: { title: "Объекты мониторинга" }
      },
      {
        path: routePaths.objectDetail,
        lazy: lazyLoader("../../../pages/ObjectDetailPage"),
        handle: { title: "Объект" },
        children: [
          {
            index: true,
            element: createElement(Navigate, { to: "devices", replace: true })
          },
          {
            path: "devices",
            lazy: lazyLoader("../../../pages/DevicesPage"),
            handle: { title: "Устройства" }
          },
          {
            path: "tags",
            lazy: lazyLoader("../../../pages/TagsPage"),
            handle: { title: "Теги" }
          },
          {
            path: "diagrams",
            lazy: lazyLoader("../../../pages/DiagramsPage"),
            handle: { title: "Мнемосхемы" }
          }
        ]
      }
    ]
  },
  {
    path: routePaths.diagramEditor,
    lazy: lazyLoader("../../../pages/EditorPage"),
    handle: { title: "Редактор" }
  }
];
