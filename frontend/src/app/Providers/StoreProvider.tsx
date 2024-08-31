import { Outlet } from "react-router-dom";
import { Provider } from "react-redux";
import { store } from "../store/store";

export const StoreProvider = () => (
  <Provider store={store}>
    <Outlet />
  </Provider>
);
