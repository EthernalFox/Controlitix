import { MantineProvider, localStorageColorSchemeManager } from "@mantine/core";

import {
  COLOR_SCHEME_STORAGE_KEY,
  cssVariablesResolver,
  defaultColorScheme,
  theme
} from "@shared/libs/theme";

import { ThemeProviderProps } from "./types";

const colorSchemeManager = localStorageColorSchemeManager({
  key: COLOR_SCHEME_STORAGE_KEY
});

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  return (
    <MantineProvider
      theme={theme}
      cssVariablesResolver={cssVariablesResolver}
      defaultColorScheme={defaultColorScheme}
      colorSchemeManager={colorSchemeManager}
    >
      {children}
    </MantineProvider>
  );
};
