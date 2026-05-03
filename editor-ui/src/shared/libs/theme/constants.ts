import type { AlarmColors, ThemeColors } from "./types";

const DEEP_BLUE_LIGHT = [
  "#e7f5ff",
  "#d0ebff",
  "#a5d8ff",
  "#74c0fc",
  "#4dabf7",
  "#339af0",
  "#228be6",
  "#1c7ed6",
  "#1971c2",
  "#1864ab"
] as const;

const GRAY_LIGHT = [
  "#F8F9FA",
  "#F1F3F5",
  "#E9ECEF",
  "#DEE2E6",
  "#CED4DA",
  "#ADB5BD",
  "#868E96",
  "#495057",
  "#343A40",
  "#212529"
] as const;

const GRAY_DARK = [
  "#C1C2C5",
  "#A6A7AB",
  "#909296",
  "#5C5F66",
  "#373A40",
  "#2C2E33",
  "#25262B",
  "#1A1B1E",
  "#141517",
  "#101113"
] as const;

const ALARM_OK = [
  "#e8f5e9",
  "#c8e6c9",
  "#a5d6a7",
  "#81c784",
  "#66bb6a",
  "#4caf50",
  "#4caf50",
  "#43a047",
  "#388e3c",
  "#2e7d32"
] as const;

const ALARM_WARN = [
  "#fff8e1",
  "#ffecb3",
  "#ffe082",
  "#ffd54f",
  "#ffca28",
  "#ffa726",
  "#ffa726",
  "#fb8c00",
  "#f57c00",
  "#ef6c00"
] as const;

const ALARM_CRIT = [
  "#ffebee",
  "#ffcdd2",
  "#ef9a9a",
  "#e57373",
  "#ef5350",
  "#ef5350",
  "#ef5350",
  "#e53935",
  "#d32f2f",
  "#c62828"
] as const;

const ALARM_UNCERTAIN = [
  "#e3f2fd",
  "#bbdefb",
  "#90caf9",
  "#64b5f6",
  "#42a5f5",
  "#42a5f5",
  "#42a5f5",
  "#1e88e5",
  "#1976d2",
  "#1565c0"
] as const;

const ALARM_BAD = [
  "#eceff1",
  "#cfd8dc",
  "#b0bec5",
  "#90a4ae",
  "#78909c",
  "#78909c",
  "#78909c",
  "#607d8b",
  "#546e7a",
  "#455a64"
] as const;

const ALARM_COMM = [
  "#eceff1",
  "#cfd8dc",
  "#b0bec5",
  "#90a4ae",
  "#78909c",
  "#607d8b",
  "#546e7a",
  "#455a64",
  "#37474f",
  "#263238"
] as const;

const ALARM_OFFLINE = [
  "#eceff1",
  "#cfd8dc",
  "#b0bec5",
  "#90a4ae",
  "#78909c",
  "#546e7a",
  "#37474f",
  "#2f3e46",
  "#263238",
  "#1c252b"
] as const;

const ALARM_ACK = [
  "#ede7f6",
  "#d1c4e9",
  "#b39ddb",
  "#9575cd",
  "#7e57c2",
  "#7e57c2",
  "#7e57c2",
  "#6a4fb3",
  "#5e35b1",
  "#4527a0"
] as const;

export const ALARM_COLORS: AlarmColors = {
  ok: "#4caf50",
  warn: "#ffa726",
  crit: "#ef5350",
  uncertain: "#42a5f5",
  bad: "#78909c",
  comm: "#546e7a",
  offline: "#37474f",
  ack: "#7e57c2"
};

export const THEME_PALETTE: ThemeColors = {
  primary: [
    "#E7F5FF",
    "#D0EBFF",
    "#A5D8FF",
    "#74C0FC",
    "#4DABF7",
    "#339AF0",
    "#228BE6",
    "#1C7ED6",
    "#1971C2",
    "#1864AB"
  ],
  secondary: [
    "#EEF3FF",
    "#DCE4F5",
    "#B9C7E2",
    "#94A8D0",
    "#748DC1",
    "#5F7CB8",
    "#5474B4",
    "#44639F",
    "#39588F",
    "#2D4B81"
  ],
  light: [...GRAY_LIGHT],
  dark: [...GRAY_DARK],
  deepBlue: [...DEEP_BLUE_LIGHT],
  alarmOk: [...ALARM_OK],
  alarmWarn: [...ALARM_WARN],
  alarmCrit: [...ALARM_CRIT],
  alarmUncertain: [...ALARM_UNCERTAIN],
  alarmBad: [...ALARM_BAD],
  alarmComm: [...ALARM_COMM],
  alarmOffline: [...ALARM_OFFLINE],
  alarmAck: [...ALARM_ACK]
};
