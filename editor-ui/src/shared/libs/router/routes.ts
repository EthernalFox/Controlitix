import { createElement } from "react";
import { Navigate, Outlet, type RouteObject } from "react-router";

import { lazyLoader } from "./LazyLoader";
import { ProtectedRoute } from "./ProtectedRoute";

export const routePaths = {
  login: "/login",
  accessDenied: "/access-denied",
  system403: "/403",
  system404: "/404",
  system500: "/500",
  sessionExpired: "/session-expired",
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
    path: routePaths.login,
    lazy: lazyLoader(() => import("@pages/LoginPage"))
  },
  {
    path: routePaths.accessDenied,
    lazy: lazyLoader(() => import("@pages/AccessDeniedPage"))
  },
  {
    path: routePaths.system403,
    lazy: lazyLoader(() => import("@pages/SystemErrorPage"))
  },
  {
    path: routePaths.system404,
    lazy: lazyLoader(() => import("@pages/SystemErrorPage"))
  },
  {
    path: routePaths.system500,
    lazy: lazyLoader(() => import("@pages/SystemErrorPage"))
  },
  {
    path: routePaths.sessionExpired,
    lazy: lazyLoader(() => import("@pages/SessionExpiredPage"))
  },
  {
    element: createElement(ProtectedRoute, null, createElement(Outlet)),
    children: [
      {
        lazy: lazyLoader(() => import("@widgets/ListLayout")),
        children: [
          {
            path: routePaths.objects,
            lazy: lazyLoader(() => import("@pages/ObjectsPage")),
            handle: { title: "РћР±СЉРµРєС‚С‹ РјРѕРЅРёС‚РѕСЂРёРЅРіР°" }
          },
          {
            path: routePaths.objectDetail,
            lazy: lazyLoader(() => import("@pages/ObjectDetailPage")),
            handle: { title: "РћР±СЉРµРєС‚" },
            children: [
              {
                index: true,
                element: createElement(Navigate, { to: "devices", replace: true })
              },
              {
                path: "devices",
                lazy: lazyLoader(() => import("@pages/DevicesPage")),
                handle: { title: "РЈСЃС‚СЂРѕР№СЃС‚РІР°" }
              },
              {
                path: "tags",
                lazy: lazyLoader(() => import("@pages/TagsPage")),
                handle: { title: "РўРµРіРё" }
              },
              {
                path: "diagrams",
                lazy: lazyLoader(() => import("@pages/DiagramsPage")),
                handle: { title: "РњРЅРµРјРѕСЃС…РµРјС‹" }
              }
            ]
          }
        ]
      },
      {
        path: routePaths.diagramEditor,
        lazy: lazyLoader(() => import("@pages/EditorPage")),
        handle: { title: "Р РµРґР°РєС‚РѕСЂ" }
      }
    ]
  },
  {
    path: "*",
    element: createElement(Navigate, { to: routePaths.system404, replace: true })
  }
];
