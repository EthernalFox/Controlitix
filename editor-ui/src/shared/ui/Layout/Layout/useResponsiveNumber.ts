import { useMantineTheme } from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";

import type { ResponsiveNumber } from "./types";

export const useResponsiveNumber = (
  value: ResponsiveNumber | undefined,
  fallback: number
): number => {
  if (typeof value === "number") return value;

  const theme = useMantineTheme();

  const xs = useMediaQuery(`(min-width: ${theme.breakpoints.xs})`);
  const sm = useMediaQuery(`(min-width: ${theme.breakpoints.sm})`);
  const md = useMediaQuery(`(min-width: ${theme.breakpoints.md})`);
  const lg = useMediaQuery(`(min-width: ${theme.breakpoints.lg})`);
  const xl = useMediaQuery(`(min-width: ${theme.breakpoints.xl})`);

  const baseValue = value?.base ?? fallback;

  if (xl && value?.xl !== undefined) return value.xl;
  if (lg && value?.lg !== undefined) return value.lg;
  if (md && value?.md !== undefined) return value.md;
  if (sm && value?.sm !== undefined) return value.sm;
  if (xs && value?.xs !== undefined) return value.xs;

  return baseValue;
};

