import { routePaths } from "./routes";

const replacePathParam = (path: string, key: string, value: string) =>
  path.replace(`:${key}`, encodeURIComponent(value));

export const buildObjectPath = (objectId: string) =>
  replacePathParam(routePaths.objectDetail, "objectId", objectId);

export const buildDevicesPath = (objectId: string) =>
  replacePathParam(routePaths.objectDevices, "objectId", objectId);

export const buildTagsPath = (objectId: string) =>
  replacePathParam(routePaths.objectTags, "objectId", objectId);

export const buildDiagramsPath = (objectId: string) =>
  replacePathParam(routePaths.objectDiagrams, "objectId", objectId);

export const buildDiagramEditorPath = (objectId: string, diagramId: string) =>
  replacePathParam(
    replacePathParam(routePaths.diagramEditor, "objectId", objectId),
    "diagramId",
    diagramId
  );
