import type { MantineColorsTuple } from "@mantine/core";

export type ThemeColors = {
  primary: MantineColorsTuple;
  secondary: MantineColorsTuple;
  deepBlue: MantineColorsTuple;
  alarmOk: MantineColorsTuple;
  alarmWarn: MantineColorsTuple;
  alarmCrit: MantineColorsTuple;
  alarmUncertain: MantineColorsTuple;
  alarmBad: MantineColorsTuple;
  alarmComm: MantineColorsTuple;
  alarmOffline: MantineColorsTuple;
  alarmAck: MantineColorsTuple;
  light: MantineColorsTuple;
  dark: MantineColorsTuple;
};

export type AlarmColors = {
  ok: string;
  warn: string;
  crit: string;
  uncertain: string;
  bad: string;
  comm: string;
  offline: string;
  ack: string;
};
