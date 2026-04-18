import type { ApiError } from "./types";
import { ApiRequestError } from "./types";

const BASE_URL = (import.meta.env.VITE_API_URL ?? "/api").replace(/\/$/, "");

type Primitive = string | number | boolean | null | undefined;

type QueryParams = Record<string, Primitive>;

interface RequestOptions extends RequestInit {
  params?: QueryParams;
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
  if (isPlainObject(data) && "type" in data && "title" in data && "status" in data) {
    return toCamelCaseDeep<ApiError>(data);
  }

  return {
    type: "/problems/internal-error",
    title: response.statusText || "Request failed",
    status: response.status,
    detail: response.statusText || "Request failed"
  };
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { params, headers, body, ...requestInit } = options;
  const url = buildUrl(path, params);

  const response = await fetch(url, {
    ...requestInit,
    headers: {
      Accept: "application/json",
      ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...(headers ?? {})
    },
    body
  });

  const responseBody = await parseResponseBody(response);

  if (!response.ok) {
    throw new ApiRequestError(toApiError(response, responseBody));
  }

  return toCamelCaseDeep<T>(responseBody);
}

export const api = {
  get: <T>(path: string, params?: QueryParams) => request<T>(path, { params }),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  delete: (path: string) => request<void>(path, { method: "DELETE" })
};
