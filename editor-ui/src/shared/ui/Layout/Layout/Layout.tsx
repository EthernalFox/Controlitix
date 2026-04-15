import { useMemo, useState } from "react";

import { AppShell } from "@shared/ui/components";

import { LayoutProvider } from "./context";
import type { LayoutProps, LayoutSizes, ResponsiveNumber } from "./types";
import { useResponsiveNumber } from "./useResponsiveNumber";

const DEFAULT_SIZES: LayoutSizes = {
  headerHeight: { base: 56, sm: 64 },
  footerHeight: { base: 48, sm: 56 },
  navbarWidth: { base: 240, md: 280 },
  panelWidth: { base: 320, md: 360 },
  asideWidth: { base: 320, md: 360 },
  breakpoint: "sm"
};

type ResponsiveObject = Partial<Record<"base" | "xs" | "sm" | "md" | "lg" | "xl", number>>;

const sumResponsive = (a: ResponsiveNumber, b: ResponsiveNumber): ResponsiveNumber => {
  if (typeof a === "number" && typeof b === "number") return a + b;

  const aObj: ResponsiveObject = typeof a === "number" ? { base: a } : (a as ResponsiveObject);
  const bObj: ResponsiveObject = typeof b === "number" ? { base: b } : (b as ResponsiveObject);

  const keys = ["base", "xs", "sm", "md", "lg", "xl"] as const;
  const result: Partial<Record<(typeof keys)[number], number>> = {};

  for (const key of keys) {
    const nextValue = (aObj[key] ?? 0) + (bObj[key] ?? 0);
    if (key === "base" || nextValue !== 0) result[key] = nextValue;
  }

  return result;
};

export const Layout = ({
  header,
  navbar,
  panel,
  aside,
  footer,
  children,
  sizes,
  defaultNavbarCollapsed = false,
  navbarCollapsed: navbarCollapsedProp,
  onNavbarCollapsedChange,
  panelHidden = false,
  asideHidden = false
}: LayoutProps) => {
  const mergedSizes = useMemo<LayoutSizes>(
    () => ({
      ...DEFAULT_SIZES,
      ...sizes
    }),
    [sizes]
  );

  const [uncontrolledNavbarCollapsed, setUncontrolledNavbarCollapsed] =
    useState(defaultNavbarCollapsed);

  const navbarCollapsed = navbarCollapsedProp ?? uncontrolledNavbarCollapsed;

  const setNavbarCollapsed = (collapsed: boolean) => {
    if (navbarCollapsedProp === undefined) {
      setUncontrolledNavbarCollapsed(collapsed);
    }
    onNavbarCollapsedChange?.(collapsed);
  };

  const toggleNavbar = () => setNavbarCollapsed(!navbarCollapsed);

  const headerHeight = mergedSizes.headerHeight;
  const footerHeight = mergedSizes.footerHeight;

  const navbarWidthCandidate = useResponsiveNumber(mergedSizes.navbarWidth, 240);
  const panelWidthCandidate = useResponsiveNumber(mergedSizes.panelWidth, 320);
  const asideWidthCandidate = useResponsiveNumber(mergedSizes.asideWidth, 320);

  const navbarWidthPx = navbar && !navbarCollapsed ? navbarWidthCandidate : 0;
  const panelWidthPx = panel && !panelHidden ? panelWidthCandidate : 0;
  const navbarContainerWidth = sumResponsive(
    navbar && !navbarCollapsed ? mergedSizes.navbarWidth : 0,
    panel && !panelHidden ? mergedSizes.panelWidth : 0
  );

  const asideWidth = aside && !asideHidden ? mergedSizes.asideWidth : 0;
  const asideWidthPx = aside && !asideHidden ? asideWidthCandidate : 0;

  const navbarEnabled = Boolean((navbar && !navbarCollapsed) || (panel && !panelHidden));
  const asideEnabled = Boolean(aside && !asideHidden);

  return (
    <LayoutProvider value={{ navbarCollapsed, setNavbarCollapsed, toggleNavbar }}>
      <AppShell
        header={header ? { height: headerHeight } : undefined}
        footer={footer ? { height: footerHeight } : undefined}
        navbar={
          navbarEnabled
            ? {
                width: navbarContainerWidth,
                breakpoint: mergedSizes.breakpoint,
                collapsed: { mobile: true }
              }
            : undefined
        }
        aside={
          asideEnabled
            ? {
                width: asideWidth,
                breakpoint: mergedSizes.breakpoint,
                collapsed: { mobile: true }
              }
            : undefined
        }
      >
        {header && <AppShell.Header>{header}</AppShell.Header>}

        {navbarEnabled && (
          <AppShell.Navbar>
            <div style={{ height: "100%", display: "flex" }}>
              {navbar && !navbarCollapsed && (
                <div
                  style={{
                    width: navbarWidthPx,
                    minWidth: navbarWidthPx,
                    maxWidth: navbarWidthPx,
                    overflow: "hidden"
                  }}
                >
                  {navbar}
                </div>
              )}
              {panel && !panelHidden && (
                <div
                  style={{
                    width: panelWidthPx,
                    minWidth: panelWidthPx,
                    maxWidth: panelWidthPx,
                    overflow: "hidden"
                  }}
                >
                  {panel}
                </div>
              )}
            </div>
          </AppShell.Navbar>
        )}

        {asideEnabled && (
          <AppShell.Aside>
            <div
              style={{
                width: asideWidthPx,
                minWidth: asideWidthPx,
                maxWidth: asideWidthPx,
                height: "100%",
                overflow: "hidden"
              }}
            >
              {aside}
            </div>
          </AppShell.Aside>
        )}

        <AppShell.Main>{children}</AppShell.Main>

        {footer && <AppShell.Footer>{footer}</AppShell.Footer>}
      </AppShell>
    </LayoutProvider>
  );
};
