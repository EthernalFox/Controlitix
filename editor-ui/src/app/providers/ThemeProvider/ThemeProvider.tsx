import {
  MantineProvider,
  localStorageColorSchemeManager,
  useMantineColorScheme
} from "@mantine/core";

import {
  COLOR_SCHEME_STORAGE_KEY,
  defaultColorScheme,
  themeOverrides
} from "@shared/libs/theme";

import { ThemeProviderProps } from "./types";

const colorSchemeManager = localStorageColorSchemeManager({
  key: COLOR_SCHEME_STORAGE_KEY
});

const ThemeOverridesProvider = ({ children }: ThemeProviderProps) => {
  const { colorScheme } = useMantineColorScheme();
  const resolvedScheme =
    colorScheme === "light" || colorScheme === "dark"
      ? colorScheme
      : defaultColorScheme;

  return (
    <MantineProvider theme={themeOverrides[resolvedScheme]}>
      {children}
    </MantineProvider>
  );
};

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  return (
    <MantineProvider
      defaultColorScheme={defaultColorScheme}
      colorSchemeManager={colorSchemeManager}
    >
      <ThemeOverridesProvider>{children}</ThemeOverridesProvider>
    </MantineProvider>
  );
};
