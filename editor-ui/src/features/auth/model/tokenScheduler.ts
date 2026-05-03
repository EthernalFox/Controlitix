const REFRESH_LEEWAY_MS = 60_000;

let refreshTimer: number | null = null;
let refreshInFlight: Promise<void> | null = null;

export const runRefreshWithMutex = async (
  refreshAction: () => Promise<void>
): Promise<void> => {
  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      await refreshAction();
    })().finally(() => {
      refreshInFlight = null;
    });
  }

  await refreshInFlight;
};

export const scheduleTokenRefresh = (
  expiresAt: number | null,
  refreshAction: () => Promise<void>
): void => {
  clearTokenRefreshSchedule();

  if (!expiresAt) {
    return;
  }

  const refreshInMs = Math.max(0, expiresAt - Date.now() - REFRESH_LEEWAY_MS);

  refreshTimer = window.setTimeout(() => {
    void refreshAction();
  }, refreshInMs);
};

export const clearTokenRefreshSchedule = (): void => {
  if (refreshTimer !== null) {
    window.clearTimeout(refreshTimer);
    refreshTimer = null;
  }
};
