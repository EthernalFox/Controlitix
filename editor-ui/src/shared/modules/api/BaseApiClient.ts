import type { QueryParamsObject } from "@shared/libs/utils";

import { BaseClient } from "./BaseClient";
import type { ApiError, ApiRequestOptions, HttpMethod } from "./types";

export abstract class BaseApiClient extends BaseClient {
  protected defaultHeaders: Record<string, string> = {
    Accept: "application/json"
  };

  protected getError(response: Response, data: unknown): ApiError {
    const message =
      typeof data === "object" && data !== null && "message" in data
        ? String((data as { message?: unknown }).message ?? response.statusText)
        : response.statusText || "Request failed";

    return {
      status: response.status,
      message,
      data
    };
  }

  protected async request<
    TResponse,
    TBody = unknown,
    TQuery extends QueryParamsObject = QueryParamsObject
  >(
    method: HttpMethod,
    options: ApiRequestOptions<TBody, TQuery>
  ): Promise<TResponse> {
    const url = this.buildUrl(options.url, options.urlParams, options.query);

    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      ...(options.headers ?? {})
    };

    const requestInit: RequestInit = {
      method,
      headers,
      signal: options.signal
    };

    if (options.body !== undefined) {
      requestInit.body = JSON.stringify(options.body);
      headers["Content-Type"] = "application/json";
    }

    const response = await fetch(url, requestInit);
    const text = await response.text();
    const data = text ? (JSON.parse(text) as TResponse) : (undefined as TResponse);

    if (!response.ok) {
      throw this.getError(response, data);
    }

    return data;
  }

  protected get<
    TResponse,
    TQuery extends QueryParamsObject = QueryParamsObject
  >(
    options: ApiRequestOptions<undefined, TQuery>
  ) {
    return this.request<TResponse, undefined, TQuery>("GET", options);
  }

  protected post<TResponse, TBody = unknown>(
    options: ApiRequestOptions<TBody>
  ) {
    return this.request<TResponse, TBody>("POST", options);
  }

  protected put<TResponse, TBody = unknown>(
    options: ApiRequestOptions<TBody>
  ) {
    return this.request<TResponse, TBody>("PUT", options);
  }

  protected patch<TResponse, TBody = unknown>(
    options: ApiRequestOptions<TBody>
  ) {
    return this.request<TResponse, TBody>("PATCH", options);
  }

  protected delete<
    TResponse,
    TQuery extends QueryParamsObject = QueryParamsObject
  >(
    options: ApiRequestOptions<undefined, TQuery>
  ) {
    return this.request<TResponse, undefined, TQuery>("DELETE", options);
  }
}
