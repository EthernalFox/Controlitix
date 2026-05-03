import type { ComponentType } from "react";
import type { LazyRouteFunction, RouteObject } from "react-router";

type LoadedRouteModule = { default: ComponentType };

type RouteImporter = () => Promise<LoadedRouteModule>;

export const lazyLoader = (importer: RouteImporter): LazyRouteFunction<RouteObject> => {
  return async () => {
    const module = await importer();

    if (!module.default) {
      throw new Error(
        "lazyLoader: imported module has no default export - did you forget `export default`?"
      );
    }

    return {
      Component: module.default
    };
  };
};
