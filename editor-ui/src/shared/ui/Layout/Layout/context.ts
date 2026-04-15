import { createContext, useContext } from "react";

export type LayoutContextValue = {
  navbarCollapsed: boolean;
  setNavbarCollapsed: (collapsed: boolean) => void;
  toggleNavbar: () => void;
};

const LayoutContext = createContext<LayoutContextValue | null>(null);

export const useLayout = () => {
  const context = useContext(LayoutContext);
  if (!context) {
    throw new Error("useLayout must be used within Layout");
  }
  return context;
};

export const LayoutProvider = LayoutContext.Provider;

