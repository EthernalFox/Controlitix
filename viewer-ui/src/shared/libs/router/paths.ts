export const APP_PATHS = {
  LOGIN: "/login",
  DASHBOARD: "/",
  DASHBOARD_ALIAS: "/dashboard",
  DIAGRAM_BY_ID: (diagramId: string) => `/diagrams/${diagramId}`,
  DIAGRAM: (objectId: string, diagramId: string) =>
    `/objects/${objectId}/diagrams/${diagramId}`,
  ALARMS: "/alarms",
  TRENDS: "/trends",
  TRENDS_TAG: (tagId: string) => `/trends/${tagId}`,
  NOT_FOUND: "/404",
  FORBIDDEN: "/403",
  SERVER_ERROR: "/500",
  SESSION_EXPIRED: "/session-expired"
} as const;
