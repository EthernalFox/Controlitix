import { buildQueryString, type QueryParamsObject } from "@shared/libs/utils";

export abstract class BaseClient {
  private static instances: Record<string, BaseClient> = {};

  protected baseUrl = "";

  protected buildUrl(
    url: string,
    urlParams?: QueryParamsObject,
    query?: QueryParamsObject
  ) {
    let resultUrl = `${this.baseUrl}${url}`;
    const queryParams: QueryParamsObject = { ...(query ?? {}) };

    Object.entries(urlParams ?? {}).forEach(([key, value]) => {
      const token = `:${key}`;
      const hasToken = resultUrl.includes(token);

      if (typeof value !== "undefined" && hasToken) {
        resultUrl = resultUrl.replace(token, encodeURIComponent(String(value)));
        delete queryParams[key];
      }
    });

    if (resultUrl.endsWith("/")) {
      resultUrl = resultUrl.slice(0, -1);
    }

    const queryString = buildQueryString(queryParams);
    if (queryString) {
      resultUrl = `${resultUrl}?${queryString}`;
    }

    return resultUrl;
  }

  public static getClient<TClient extends BaseClient>(
    Constructable: new () => TClient
  ): TClient {
    const key = Constructable.name;

    if (!BaseClient.instances[key]) {
      BaseClient.instances[key] = new Constructable();
    }

    return BaseClient.instances[key] as TClient;
  }
}
