import { createTheme } from "@mantine/core";

export const defaultColorScheme = "dark" as const;

export const COLOR_SCHEME_STORAGE_KEY = "viewer-ui:color-scheme";

export const mantineTheme = createTheme({
  primaryColor: "primary",
  fontFamily: "Inter, sans-serif",
  fontFamilyMonospace: "JetBrains Mono, monospace",
  defaultRadius: "sm",
  colors: {
    primary: ["#e5f0ff", "#cddfff", "#9abfff", "#659cff", "#3c82fe", "#2372fe", "#0f68ff", "#0058e5", "#004fcd", "#0045b4"],
    secondary: ["#ebeefb", "#d4daf4", "#a8b3e9", "#7a8cdd", "#5870d4", "#445ccf", "#3a51cd", "#2f44b7", "#293ca4", "#22348f"],
    "alarm-ok": ["#ebfaee", "#d6f2dc", "#ace6b8", "#7fd993", "#5ace73", "#44c760", "#34c153", "#24ab45", "#17993b", "#07852f"],
    "alarm-warn": ["#fff4e4", "#fee7cb", "#fdd09b", "#fcb767", "#fba23b", "#fb9320", "#fb8a11", "#e07704", "#c86a00", "#ae5b00"],
    "alarm-alarm": ["#ffe9e9", "#ffd1d1", "#fda2a2", "#f97171", "#f54848", "#f43131", "#f42525", "#d91515", "#c20f0f", "#a90000"],
    "alarm-uncertain": ["#e8f4ff", "#cfe8ff", "#9fd0ff", "#6db7ff", "#46a3ff", "#2f96ff", "#228fff", "#0f7be5", "#006ece", "#005fb5"],
    "alarm-bad": ["#edf1f3", "#dce2e6", "#bcc6cd", "#9aa9b3", "#7f929f", "#6d8594", "#627e8e", "#516d7c", "#445f6d", "#34505d"],
    "alarm-comm": ["#e9eef0", "#d4dde1", "#a8bcc4", "#799aa5", "#587f8e", "#456f80", "#3b677a", "#2d5669", "#244c5f", "#174154"],
    "alarm-offline": ["#e8ebec", "#d2d8da", "#a5b0b4", "#75878d", "#51686f", "#3f5961", "#354f58", "#273f47", "#1c363f", "#0d2c35"],
    "alarm-ack": ["#f1ebfa", "#e0d5f0", "#c1a9e1", "#a27cd4", "#8859c8", "#7a44c2", "#7339c0", "#632ca9", "#572699", "#4a1f86"]
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
  }
});
