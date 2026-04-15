import { LazyRouteFunction, RouteObject } from "react-router";

export const lazyLoader =
  (importPath: string): LazyRouteFunction<RouteObject> =>
  async () => {
    const module = await import(importPath);

    return {
      Component: module.default
    };
  };
