import { createContext, useContext } from "react";

import type { LayoutMode } from "@shared/ui/Layout/types";

export type LayoutContextValue = {
  mode: LayoutMode;
  navbarCollapsed: boolean;
  panelCollapsed: boolean;
  asideCollapsed: boolean;

  setNavbarCollapsed: (collapsed: boolean) => void;
  setPanelCollapsed: (collapsed: boolean) => void;
  setAsideCollapsed: (collapsed: boolean) => void;

  toggleNavbar: () => void;
  togglePanel: () => void;
  toggleAside: () => void;
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
