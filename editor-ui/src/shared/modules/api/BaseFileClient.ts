import { BaseClient } from "./BaseClient";
import type { ApiError, FileDownloadOptions, FileUploadOptions } from "./types";

export abstract class BaseFileClient extends BaseClient {
  protected defaultHeaders: Record<string, string> = {};

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

  protected async upload<TResponse = void, TData extends Record<string, unknown> = Record<string, unknown>>(
    options: FileUploadOptions<TData>
  ): Promise<TResponse> {
    const url = this.buildUrl(options.url, options.urlParams, options.query);
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      ...(options.headers ?? {})
    };
    const formData = new FormData();
    const files = Array.from(options.files);
    const fieldName = options.fieldName ?? "file";

    files.forEach((file) => {
      formData.append(fieldName, file);
    });

    if (options.data) {
      Object.entries(options.data).forEach(([key, value]) => {
        if (value === null || value === undefined) return;
        formData.append(key, String(value));
      });
    }

    const response = await fetch(url, {
      method: options.method ?? "POST",
      headers,
      body: formData,
      signal: options.signal
    });

    const text = await response.text();
    const data = text ? (JSON.parse(text) as TResponse) : (undefined as TResponse);

    if (!response.ok) {
      throw this.getError(response, data);
    }

    return data;
  }

  protected async download(options: FileDownloadOptions): Promise<Blob> {
    const url = this.buildUrl(options.url, options.urlParams, options.query);
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      ...(options.headers ?? {})
    };

    const response = await fetch(url, {
      method: "GET",
      headers,
      signal: options.signal
    });

    if (!response.ok) {
      const text = await response.text();
      const data = text ? JSON.parse(text) : undefined;

      throw this.getError(response, data);
    }

    return response.blob();
  }
}
