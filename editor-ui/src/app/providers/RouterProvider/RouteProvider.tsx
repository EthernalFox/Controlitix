import { RouterProvider } from "react-router";

import { router } from "@shared/libs/router";

export const RouteProvider = () => {
  return <RouterProvider router={router} />;
};
