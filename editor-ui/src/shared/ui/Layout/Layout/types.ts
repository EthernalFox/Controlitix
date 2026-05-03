import type { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";
import type { LayoutMode } from "@shared/ui/Layout/types";

export type Breakpoint = "base" | "xs" | "sm" | "md" | "lg" | "xl";

export type ResponsiveNumber =
  | number
  | Partial<Record<Exclude<Breakpoint, "base"> | "base", number>>;

export type LayoutSlots = {
  header: ReactNode;
  navbar: ReactNode;
  panel: ReactNode;
  aside: ReactNode;
  footer: ReactNode;
};

export type LayoutSizes = {
  headerHeight: ResponsiveNumber;
  footerHeight: ResponsiveNumber;
  navbarWidth: ResponsiveNumber;
  panelWidth: ResponsiveNumber;
  asideWidth: ResponsiveNumber;
  breakpoint: Exclude<Breakpoint, "base">;
};

export type LayoutProps = RequireAtLeastOne<Partial<LayoutSlots>, keyof LayoutSlots> & {
  children: ReactNode;
  mode?: LayoutMode;

  sizes?: Partial<LayoutSizes>;

  defaultNavbarCollapsed?: boolean;
  defaultPanelCollapsed?: boolean;
  defaultAsideCollapsed?: boolean;

  navbarCollapsed?: boolean;
  panelCollapsed?: boolean;
  asideCollapsed?: boolean;

  onNavbarCollapsedChange?: (collapsed: boolean) => void;
  onPanelCollapsedChange?: (collapsed: boolean) => void;
  onAsideCollapsedChange?: (collapsed: boolean) => void;

  panelHidden?: boolean;
  asideHidden?: boolean;
};
