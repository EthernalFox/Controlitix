import {
  getAccessToken,
  notifySessionExpired,
  notifyUnauthenticated,
  runRefreshAccessToken
} from "./auth-token";
import type { ApiError } from "./types";
import { ApiRequestError } from "./types";

const BASE_URL = (import.meta.env.VITE_API_URL ?? "/api").replace(/\/$/, "");

type Primitive = string | number | boolean | null | undefined;

type QueryParams = Record<string, Primitive>;

interface RequestOptions extends RequestInit {
  params?: QueryParams;
}

interface RequestMeta {
  skipAuthRefresh: boolean;
}

const isPlainObject = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const snakeToCamel = (value: string) =>
  value.replace(/_([a-z])/g, (_, char: string) => char.toUpperCase());

const toCamelCaseDeep = <T>(value: unknown): T => {
  if (Array.isArray(value)) {
    return value.map((item) => toCamelCaseDeep(item)) as T;
  }

  if (!isPlainObject(value)) {
    return value as T;
  }

  const entries = Object.entries(value).map(([key, itemValue]) => [
    snakeToCamel(key),
    toCamelCaseDeep(itemValue)
  ]);

  return Object.fromEntries(entries) as T;
};

const buildUrl = (path: string, params?: QueryParams) => {
  const preparedPath = path.startsWith("/") ? path : `/${path}`;
  const url = new URL(`${BASE_URL}${preparedPath}`, window.location.origin);

  Object.entries(params ?? {}).forEach(([key, value]) => {
    if (value === null || value === undefined || value === "") {
      return;
    }

    url.searchParams.set(key, String(value));
  });

  return `${url.pathname}${url.search}`;
};

const parseResponseBody = async (response: Response): Promise<unknown> => {
  const text = await response.text();
  if (!text) {
    return null;
  }

  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
};

const toApiError = (response: Response, data: unknown): ApiError => {
  const retryAfterHeader = response.headers.get("Retry-After");
  const retryAfter = retryAfterHeader ? Number.parseInt(retryAfterHeader, 10) : undefined;

  if (isPlainObject(data) && "type" in data && "title" in data && "status" in data) {
    const mappedError = toCamelCaseDeep<ApiError>(data);
    if (!Number.isNaN(retryAfter ?? Number.NaN)) {
      mappedError.retryAfter = retryAfter;
    }

    return mappedError;
  }

  return {
    type: "/problems/internal-error",
    title: response.statusText || "Request failed",
    status: response.status,
    detail: response.statusText || "Request failed",
    retryAfter: !Number.isNaN(retryAfter ?? Number.NaN) ? retryAfter : undefined
  };
};

const shouldSkipRefreshForPath = (path: string): boolean => {
  return (
    path.startsWith("/auth/login") ||
    path.startsWith("/auth/refresh") ||
    path.startsWith("/auth/logout")
  );
};

const shouldAttemptAuthRefresh = (
  path: string,
  apiError: ApiError,
  meta: RequestMeta
): boolean => {
  if (meta.skipAuthRefresh || shouldSkipRefreshForPath(path)) {
    return false;
  }

  if (apiError.status !== 401) {
    return false;
  }

  return (
    apiError.type === "/errors/auth/token-expired" ||
    apiError.type === "/errors/auth/invalid-token"
  );
};

const buildRequestInit = (options: RequestOptions): RequestInit => {
  const requestInit = { ...options } as RequestOptions;
  const { headers, body } = requestInit;
  delete requestInit.params;

  const accessToken = getAccessToken();

  return {
    ...requestInit,
    credentials: "include",
    headers: {
      Accept: "application/json",
      ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...(headers ?? {})
    },
    body
  };
};

async function request<T>(
  path: string,
  options: RequestOptions = {},
  meta: RequestMeta = { skipAuthRefresh: false }
): Promise<T> {
  const { params } = options;
  const url = buildUrl(path, params);

  try {
    const response = await fetch(url, buildRequestInit(options));
    const responseBody = await parseResponseBody(response);

    if (!response.ok) {
      const apiError = toApiError(response, responseBody);

      if (shouldAttemptAuthRefresh(path, apiError, meta)) {
        try {
          await runRefreshAccessToken();
        } catch {
          notifySessionExpired();
          throw new ApiRequestError(apiError);
        }

        try {
          return await request<T>(path, options, { skipAuthRefresh: true });
        } catch (retryError) {
          if (retryError instanceof ApiRequestError && retryError.payload.status === 401) {
            notifyUnauthenticated();
          }

          throw retryError;
        }
      }

      throw new ApiRequestError(apiError);
    }

    return toCamelCaseDeep<T>(responseBody);
  } catch (error) {
    if (error instanceof ApiRequestError) {
      throw error;
    }

    throw new ApiRequestError({
      type: "/problems/internal-error",
      title: "Request failed",
      status: 0,
      detail: "Network request failed"
    });
  }
}

export const api = {
  get: <T>(path: string, params?: QueryParams) => request<T>(path, { params }),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      ...(body === undefined ? {} : { body: JSON.stringify(body) })
    }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  delete: (path: string) => request<void>(path, { method: "DELETE" })
};
