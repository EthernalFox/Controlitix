import type { QueryParamsObject } from "@shared/libs/utils";

export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export interface ApiError {
  status: number;
  message: string;
  data?: unknown;
}

export interface ApiRequestOptions<
  TBody = unknown,
  TQuery extends QueryParamsObject = QueryParamsObject
> {
  url: string;
  urlParams?: QueryParamsObject;
  query?: TQuery;
  body?: TBody;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

export interface FileUploadOptions<TData extends Record<string, unknown> = Record<string, unknown>> {
  url: string;
  urlParams?: QueryParamsObject;
  query?: QueryParamsObject;
  files: File[] | FileList;
  fieldName?: string;
  data?: TData;
  method?: "POST" | "PUT" | "PATCH";
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

export interface FileDownloadOptions {
  url: string;
  urlParams?: QueryParamsObject;
  query?: QueryParamsObject;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}
