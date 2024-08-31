import {
    RouterProvider,
  } from "react-router-dom";
import { router } from "../router/router";

export const RouteProvider = () => <RouterProvider router={router} />