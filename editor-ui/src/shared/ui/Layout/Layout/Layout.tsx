import { useMemo, useState } from "react";

import { AppShell } from "@shared/ui/components";
import type { LayoutMode } from "@shared/ui/Layout/types";

import { AmbientOrbs } from "../AmbientOrbs";
import { LayoutProvider } from "./context";
import type { LayoutProps, LayoutSizes, ResponsiveNumber } from "./types";
import { useResponsiveNumber } from "./useResponsiveNumber";
import glassStyles from "../styles/glass.module.css";

const DEFAULT_SIZES: LayoutSizes = {
  headerHeight: 56,
  footerHeight: 48,
  navbarWidth: 240,
  panelWidth: 240,
  asideWidth: 280,
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

const joinClasses = (...classes: Array<string | undefined | false>) =>
  classes.filter(Boolean).join(" ");

const warnIncompatibleSlot = (mode: LayoutMode, slotName: "navbar" | "panel") => {
  if (import.meta.env.DEV) {
    console.warn(`Layout: slot "${slotName}" is ignored in mode="${mode}"`);
  }
};

export const Layout = ({
  header,
  navbar,
  panel,
  aside,
  footer,
  children,
  mode = "list",
  sizes,
  defaultNavbarCollapsed = false,
  defaultPanelCollapsed = false,
  defaultAsideCollapsed = false,
  navbarCollapsed: navbarCollapsedProp,
  panelCollapsed: panelCollapsedProp,
  asideCollapsed: asideCollapsedProp,
  onNavbarCollapsedChange,
  onPanelCollapsedChange,
  onAsideCollapsedChange,
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
  const [uncontrolledPanelCollapsed, setUncontrolledPanelCollapsed] =
    useState(defaultPanelCollapsed);
  const [uncontrolledAsideCollapsed, setUncontrolledAsideCollapsed] =
    useState(defaultAsideCollapsed);

  const navbarCollapsed = navbarCollapsedProp ?? uncontrolledNavbarCollapsed;
  const panelCollapsed = panelCollapsedProp ?? uncontrolledPanelCollapsed;
  const asideCollapsed = asideCollapsedProp ?? uncontrolledAsideCollapsed;

  const setNavbarCollapsed = (collapsed: boolean) => {
    if (navbarCollapsedProp === undefined) {
      setUncontrolledNavbarCollapsed(collapsed);
    }
    onNavbarCollapsedChange?.(collapsed);
  };

  const setPanelCollapsed = (collapsed: boolean) => {
    if (panelCollapsedProp === undefined) {
      setUncontrolledPanelCollapsed(collapsed);
    }
    onPanelCollapsedChange?.(collapsed);
  };

  const setAsideCollapsed = (collapsed: boolean) => {
    if (asideCollapsedProp === undefined) {
      setUncontrolledAsideCollapsed(collapsed);
    }
    onAsideCollapsedChange?.(collapsed);
  };

  const toggleNavbar = () => setNavbarCollapsed(!navbarCollapsed);
  const togglePanel = () => setPanelCollapsed(!panelCollapsed);
  const toggleAside = () => setAsideCollapsed(!asideCollapsed);

  const isListMode = mode === "list";
  const isEditorMode = mode === "editor";

  const showNavbarSlot = isListMode && Boolean(navbar);
  const showPanelSlot = isEditorMode && Boolean(panel) && !panelHidden;
  const showAsideSlot = Boolean(aside) && !asideHidden;

  if (isListMode && panel && !panelHidden) {
    warnIncompatibleSlot(mode, "panel");
  }

  if (isEditorMode && navbar) {
    warnIncompatibleSlot(mode, "navbar");
  }

  const headerHeight = mergedSizes.headerHeight;
  const footerHeight = mergedSizes.footerHeight;

  const navbarWidthCandidate = useResponsiveNumber(mergedSizes.navbarWidth, 240);
  const panelWidthCandidate = useResponsiveNumber(mergedSizes.panelWidth, 240);
  const asideWidthCandidate = useResponsiveNumber(mergedSizes.asideWidth, 280);

  const navbarWidthPx = showNavbarSlot && !navbarCollapsed ? navbarWidthCandidate : 0;
  const panelWidthPx = showPanelSlot && !panelCollapsed ? panelWidthCandidate : 0;
  const asideWidthPx = showAsideSlot && !asideCollapsed ? asideWidthCandidate : 0;

  const navbarContainerWidth = sumResponsive(
    showNavbarSlot && !navbarCollapsed ? mergedSizes.navbarWidth : 0,
    showPanelSlot && !panelCollapsed ? mergedSizes.panelWidth : 0
  );

  const asideWidth = showAsideSlot && !asideCollapsed ? mergedSizes.asideWidth : 0;

  const navbarEnabled = showNavbarSlot || showPanelSlot;
  const asideEnabled = showAsideSlot;

  return (
    <LayoutProvider
      value={{
        mode,
        navbarCollapsed,
        panelCollapsed,
        asideCollapsed,
        setNavbarCollapsed,
        setPanelCollapsed,
        setAsideCollapsed,
        toggleNavbar,
        togglePanel,
        toggleAside
      }}
    >
      <AmbientOrbs />

      <AppShell
        padding={0}
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
        {header && (
          <AppShell.Header>
            <div
              className={joinClasses(glassStyles.glass, glassStyles.glassFlat)}
              style={{ height: "100%", willChange: "backdrop-filter" }}
            >
              {header}
            </div>
          </AppShell.Header>
        )}

        {navbarEnabled && (
          <AppShell.Navbar>
            <div
              className={glassStyles.glass}
              style={{
                height: "100%",
                display: "flex",
                borderRadius: 0,
                willChange: "backdrop-filter"
              }}
            >
              {showNavbarSlot && (
                <div
                  className={glassStyles.slot}
                  style={{
                    width: navbarWidthPx,
                    minWidth: navbarWidthPx,
                    maxWidth: navbarWidthPx,
                    height: "100%"
                  }}
                >
                  {navbar}
                </div>
              )}

              {showPanelSlot && (
                <div
                  className={glassStyles.slot}
                  style={{
                    width: panelWidthPx,
                    minWidth: panelWidthPx,
                    maxWidth: panelWidthPx,
                    height: "100%"
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
            <div className={glassStyles.glass} style={{ height: "100%", borderRadius: 0 }}>
              <div
                className={glassStyles.slot}
                style={{
                  width: asideWidthPx,
                  minWidth: asideWidthPx,
                  maxWidth: asideWidthPx,
                  height: "100%"
                }}
              >
                {aside}
              </div>
            </div>
          </AppShell.Aside>
        )}

        <AppShell.Main
          style={{
            minWidth: 0,
            overflowX: "hidden",
            marginInlineStart: "var(--app-shell-navbar-offset, 0px)",
            marginInlineEnd: "var(--app-shell-aside-offset, 0px)",
            paddingInlineStart: 0,
            paddingInlineEnd: 0
          }}
        >
          {children}
        </AppShell.Main>

        {footer && (
          <AppShell.Footer>
            <div className={joinClasses(glassStyles.glass, glassStyles.glassFlat)} style={{ height: "100%" }}>
              {footer}
            </div>
          </AppShell.Footer>
        )}
      </AppShell>
    </LayoutProvider>
  );
};
