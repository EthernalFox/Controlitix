import { createTheme, rem, type MantineThemeOverride } from "@mantine/core";

import { DARK_PALETTE, LIGHT_PALETTE } from "./constants";

type ThemeName = "light" | "dark";

export const defaultColorScheme: ThemeName = "dark";

const FONT_FAMILY = "Roboto, sans-serif";

const baseTheme = createTheme({
  primaryColor: "primary",

  shadows: {
    md: "1px 1px 3px rgba(0, 0, 0, .25)",
    xl: "5px 5px 3px rgba(0, 0, 0, .25)"
  },

  headings: {
    fontFamily: FONT_FAMILY,
    sizes: {
      h1: { fontSize: rem(36) }
    }
  },
  fontFamily: FONT_FAMILY
});

export const themeOverrides = {
  light: createTheme({
    ...baseTheme,
    colors: LIGHT_PALETTE
  }),
  dark: createTheme({
    ...baseTheme,
    colors: DARK_PALETTE
  })
} satisfies Record<ThemeName, MantineThemeOverride>;
