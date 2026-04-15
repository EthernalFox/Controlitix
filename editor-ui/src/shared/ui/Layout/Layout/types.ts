import type { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

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

  sizes?: Partial<LayoutSizes>;

  defaultNavbarCollapsed?: boolean;
  navbarCollapsed?: boolean;
  onNavbarCollapsedChange?: (collapsed: boolean) => void;

  panelHidden?: boolean;
  asideHidden?: boolean;
};
