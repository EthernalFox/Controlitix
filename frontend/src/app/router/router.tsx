import { createBrowserRouter } from "react-router-dom";
import overviewRoute from "./routes/overviewRoute";
import AppLayout from "@app/layout/AppLayout";

export const router = createBrowserRouter([{
    path: '/',
    element: <AppLayout />,
    children: [overviewRoute]
}])