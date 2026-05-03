import { createTheme, type CSSVariablesResolver } from "@mantine/core";

import { ALARM_COLORS, THEME_PALETTE } from "./constants";
import badgeClasses from "../../ui/components/Badge/Badge.module.css";
import buttonClasses from "../../ui/components/Button/Button.module.css";
import cardClasses from "../../ui/components/Card/Card.module.css";
import modalClasses from "../../ui/components/Modal/Modal.module.css";
import notificationClasses from "../../ui/components/Notification/Notification.module.css";
import switchClasses from "../../ui/components/Switch/Switch.module.css";
import tabsClasses from "../../ui/components/Tabs/Tabs.module.css";
import textInputClasses from "../../ui/components/TextInput/TextInput.module.css";

export const theme = createTheme({
  primaryColor: "deepBlue",
  fontFamily: "Inter, sans-serif",
  fontFamilyMonospace: "JetBrains Mono, monospace",
  defaultRadius: "sm",
  radius: {
    xs: "2px",
    sm: "4px",
    md: "8px",
    lg: "14px",
    xl: "18px"
  },
  spacing: {
    xs: "4px",
    sm: "8px",
    md: "16px",
    lg: "24px",
    xl: "32px"
  },
  fontSizes: {
    xs: "11px",
    sm: "13px",
    md: "14px",
    lg: "16px",
    xl: "20px"
  },
  headings: {
    fontFamily: "Inter, sans-serif",
    fontWeight: "600",
    sizes: {
      h1: { fontSize: "20px", lineHeight: "1.4" },
      h2: { fontSize: "18px", lineHeight: "1.4" },
      h3: { fontSize: "16px", lineHeight: "1.4" },
      h4: { fontSize: "14px", lineHeight: "1.4" },
      h5: { fontSize: "13px", lineHeight: "1.4" },
      h6: { fontSize: "11px", lineHeight: "1.4" }
    }
  },
  colors: THEME_PALETTE,
  other: {
    alarm: ALARM_COLORS
  },
  components: {
    Button: {
      defaultProps: { radius: "sm" },
      classNames: buttonClasses
    },
    TextInput: {
      classNames: textInputClasses
    },
    NumberInput: {
      classNames: textInputClasses
    },
    PasswordInput: {
      classNames: textInputClasses
    },
    Textarea: {
      classNames: textInputClasses
    },
    Tabs: {
      classNames: tabsClasses
    },
    Switch: {
      classNames: switchClasses
    },
    Modal: {
      classNames: modalClasses
    },
    Drawer: {
      classNames: modalClasses
    },
    Popover: {
      classNames: modalClasses
    },
    Menu: {
      classNames: modalClasses
    },
    Card: {
      classNames: cardClasses
    },
    Badge: {
      classNames: badgeClasses
    },
    Notification: {
      classNames: notificationClasses
    }
  }
});

export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {
    "--ctrx-header-h": "56px",
    "--ctrx-panel-w": "240px",
    "--ctrx-aside-w": "280px",
    "--ctrx-footer-h": "48px",
    "--ctrx-glass-blur": "saturate(180%) blur(20px)",
    "--ctrx-orb-blur": "blur(80px)",
    "--ctrx-alarm-ok": ALARM_COLORS.ok,
    "--ctrx-alarm-warn": ALARM_COLORS.warn,
    "--ctrx-alarm-crit": ALARM_COLORS.crit,
    "--ctrx-alarm-uncertain": ALARM_COLORS.uncertain,
    "--ctrx-alarm-bad": ALARM_COLORS.bad,
    "--ctrx-alarm-comm": ALARM_COLORS.comm,
    "--ctrx-alarm-offline": ALARM_COLORS.offline,
    "--ctrx-alarm-ack": ALARM_COLORS.ack
  },
  light: {
    "--ctrx-canvas-bg": "#f5f5f5",
    "--ctrx-canvas-grid": "#dee2e6",
    "--ctrx-glass-bg": "rgba(255,255,255,0.55)",
    "--ctrx-glass-bg-strong": "rgba(255,255,255,0.72)",
    "--ctrx-glass-bg-soft": "rgba(255,255,255,0.32)",
    "--ctrx-glass-border": "rgba(255,255,255,0.7)",
    "--ctrx-glass-border-inner": "rgba(34,139,230,0.08)",
    "--ctrx-glass-shadow":
      "0 1px 0 rgba(255,255,255,0.7) inset, 0 -1px 0 rgba(0,0,0,0.04) inset, 0 8px 32px -8px rgba(15,23,42,0.18), 0 2px 8px -2px rgba(15,23,42,0.08)",
    "--ctrx-glass-highlight":
      "linear-gradient(180deg, rgba(255,255,255,0.5) 0%, rgba(255,255,255,0) 50%)",
    "--ctrx-orb-1": "#bad7ff",
    "--ctrx-orb-2": "#ffd6e8",
    "--ctrx-orb-3": "#c8f0dc",
    "--ctrx-orb-4": "#ffe4b0"
  },
  dark: {
    "--ctrx-canvas-bg": "#111422",
    "--ctrx-canvas-grid": "#2d3154",
    "--ctrx-glass-bg": "rgba(35,39,64,0.55)",
    "--ctrx-glass-bg-strong": "rgba(35,39,64,0.78)",
    "--ctrx-glass-bg-soft": "rgba(26,29,46,0.4)",
    "--ctrx-glass-border": "rgba(255,255,255,0.08)",
    "--ctrx-glass-border-inner": "rgba(76,154,255,0.18)",
    "--ctrx-glass-shadow":
      "0 1px 0 rgba(255,255,255,0.06) inset, 0 -1px 0 rgba(0,0,0,0.3) inset, 0 12px 40px -8px rgba(0,0,0,0.5), 0 2px 8px -2px rgba(0,0,0,0.3)",
    "--ctrx-glass-highlight":
      "linear-gradient(180deg, rgba(255,255,255,0.07) 0%, rgba(255,255,255,0) 60%)",
    "--ctrx-orb-1": "#1e3a6b",
    "--ctrx-orb-2": "#4c2a52",
    "--ctrx-orb-3": "#1f4738",
    "--ctrx-orb-4": "#5a4019"
  }
});

export const defaultColorScheme = "light" as const;
