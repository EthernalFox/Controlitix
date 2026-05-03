let accessToken: string | null = null;

let refreshAccessTokenHandler: (() => Promise<void>) | null = null;
let unauthenticatedHandler: (() => void) | null = null;
let sessionExpiredHandler: (() => void) | null = null;

export const getAccessToken = (): string | null => {
  return accessToken;
};

export const setAccessToken = (nextAccessToken: string | null): void => {
  accessToken = nextAccessToken;
};

export const clearAccessToken = (): void => {
  accessToken = null;
};

export const setRefreshAccessTokenHandler = (
  handler: (() => Promise<void>) | null
): void => {
  refreshAccessTokenHandler = handler;
};

export const runRefreshAccessToken = async (): Promise<void> => {
  if (!refreshAccessTokenHandler) {
    throw new Error("refresh handler is not configured");
  }

  await refreshAccessTokenHandler();
};

export const setUnauthenticatedHandler = (
  handler: (() => void) | null
): void => {
  unauthenticatedHandler = handler;
};

export const notifyUnauthenticated = (): void => {
  unauthenticatedHandler?.();
};

export const setSessionExpiredHandler = (
  handler: (() => void) | null
): void => {
  sessionExpiredHandler = handler;
};

export const notifySessionExpired = (): void => {
  sessionExpiredHandler?.();
};
